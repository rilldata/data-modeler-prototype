package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/queries"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const alertExecutionHistoryLimit = 25

const alertDefaultIntervalsLimit = 25

const alertQueryPriority = 1

const alertCheckDefaultTimeout = 5 * time.Minute

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindAlert, newAlertReconciler)
}

type AlertReconciler struct {
	C *runtime.Controller
}

func newAlertReconciler(c *runtime.Controller) runtime.Reconciler {
	return &AlertReconciler{C: c}
}

func (r *AlertReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *AlertReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetAlert()
	b := to.GetAlert()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *AlertReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetAlert()
	b := to.GetAlert()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *AlertReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetAlert().State = &runtimev1.AlertState{}
	return nil
}

func (r *AlertReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	a := self.GetAlert()
	if a == nil {
		return runtime.ReconcileResult{Err: errors.New("not an alert")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// If CurrentExecution is not nil, a catastrophic failure occurred during the last execution.
	// Clean up to ensure CurrentExecution is nil.
	if a.State.CurrentExecution != nil {
		// Don't pop it, just pretend it never happened
		a.State.CurrentExecution = nil
		err := r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Exit early if disabled
	if a.Spec.RefreshSchedule != nil && a.Spec.RefreshSchedule.Disable {
		return runtime.ReconcileResult{}
	}

	// Unlike other resources, alerts have different hashes for the spec and the refs' state.
	// This enables differentiating behavior between changes to the spec and changes to the refs.
	// When the spec changes, we clear all alert state. When the refs change, we just use it to trigger the alert ()
	specHash, err := r.executionSpecHash(ctx, a.Spec, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}
	refsHash, err := r.refsStateHash(ctx, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}

	// Determine whether to trigger
	adhocTrigger := a.Spec.Trigger
	specHashTrigger := a.State.SpecHash != specHash
	refsTrigger := a.State.RefsHash != refsHash && a.Spec.RefreshSchedule != nil && a.Spec.RefreshSchedule.RefUpdate
	scheduleTrigger := a.State.NextRunOn != nil && !a.State.NextRunOn.AsTime().After(time.Now())
	trigger := adhocTrigger || specHashTrigger || refsTrigger || scheduleTrigger

	// If not triggering now, update NextRunOn and retrigger when it falls due
	if !trigger {
		err = r.updateNextRunOn(ctx, self, a)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		if a.State.NextRunOn != nil {
			return runtime.ReconcileResult{Retrigger: a.State.NextRunOn.AsTime()}
		}
		return runtime.ReconcileResult{}
	}

	// If the spec hash changed, clear all alert state
	if specHashTrigger {
		a.State.SpecHash = specHash
		a.State.RefsHash = ""
		a.State.NextRunOn = nil
		a.State.CurrentExecution = nil
		a.State.ExecutionHistory = nil
		a.State.ExecutionCount = 0
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Evaluate the trigger time of the alert. If triggered by schedule, we use the "clean" scheduled time.
	// Note: Correction for watermarks and intervals is done in checkAlert.
	var triggerTime time.Time
	if scheduleTrigger && !adhocTrigger && !specHashTrigger && !refsTrigger {
		triggerTime = a.State.NextRunOn.AsTime()
	} else {
		triggerTime = time.Now()
	}

	// Run alert queries and send emails
	executeErr := r.executeAll(ctx, self, a, triggerTime, adhocTrigger)

	// If we were cancelled, exit without updating any other trigger-related state.
	// NOTE: We don't set Retrigger here because we'll leave re-scheduling to whatever cancelled the reconciler.
	if errors.Is(executeErr, context.Canceled) {
		return runtime.ReconcileResult{Err: executeErr}
	}

	// Advance NextRunOn
	err = r.updateNextRunOn(ctx, self, a)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Clear ad-hoc trigger
	if a.Spec.Trigger {
		err := r.setTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to clear trigger: %w", err)}
		}
	}

	// Update refs hash
	if refsTrigger {
		a.State.RefsHash = refsHash
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Done
	if a.State.NextRunOn != nil {
		return runtime.ReconcileResult{Err: executeErr, Retrigger: a.State.NextRunOn.AsTime()}
	}
	return runtime.ReconcileResult{Err: executeErr}
}

// executionSpecHash computes a hash of the alert properties that impact execution.
// NOTE: Unlike other resources, we don't include the refs' state version in the hash since it's managed separately using refsStateHash.
func (r *AlertReconciler) executionSpecHash(ctx context.Context, spec *runtimev1.AlertSpec, refs []*runtimev1.ResourceName) (string, error) {
	hash := md5.New()

	for _, ref := range refs { // Refs are always sorted
		// Write name
		_, err := hash.Write([]byte(ref.Kind))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(ref.Name))
		if err != nil {
			return "", err
		}
	}

	if spec.RefreshSchedule != nil {
		_, err := hash.Write([]byte(spec.RefreshSchedule.TimeZone))
		if err != nil {
			return "", err
		}
	}

	err := binary.Write(hash, binary.BigEndian, spec.WatermarkInherit)
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.IntervalsIsoDuration))
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.IntervalsCheckUnclosed)
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.QueryName))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.QueryArgsJson))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.GetQueryForUserId()))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.GetQueryForUserEmail()))
	if err != nil {
		return "", err
	}

	if spec.GetQueryForAttributes() != nil {
		v := structpb.NewStructValue(spec.GetQueryForAttributes())
		err = pbutil.WriteHash(v, hash)
		if err != nil {
			return "", err
		}
	}

	err = binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// refsStateHash computes a hash of the refs and their state versions.
func (r *AlertReconciler) refsStateHash(ctx context.Context, refs []*runtimev1.ResourceName) (string, error) {
	hash := md5.New()

	for _, ref := range refs { // Refs are always sorted
		// Write name
		_, err := hash.Write([]byte(ref.Kind))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(ref.Name))
		if err != nil {
			return "", err
		}

		// Note: Only writing the state info to the hash, not spec version, because it doesn't matter whether the spec/meta changes, only whether the state changes.
		// Note: Also using StateUpdatedOn because the state version is reset when the resource is deleted and recreated.
		r, err := r.C.Get(ctx, ref, false)
		var stateVersion, stateUpdatedOn int64
		if err == nil {
			stateVersion = r.Meta.StateVersion
			stateUpdatedOn = r.Meta.StateUpdatedOn.Seconds
		} else {
			stateVersion = -1
		}
		err = binary.Write(hash, binary.BigEndian, stateVersion)
		if err != nil {
			return "", err
		}
		err = binary.Write(hash, binary.BigEndian, stateUpdatedOn)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// updateNextRunOn evaluates the alert's schedule relative to the current time, and updates the NextRunOn state accordingly.
// If the schedule is nil, it will set NextRunOn to nil.
func (r *AlertReconciler) updateNextRunOn(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert) error {
	next, err := nextRefreshTime(time.Now(), a.Spec.RefreshSchedule)
	if err != nil {
		return err
	}

	var curr time.Time
	if a.State.NextRunOn != nil {
		curr = a.State.NextRunOn.AsTime()
	}

	if next == curr {
		return nil
	}

	if next.IsZero() {
		a.State.NextRunOn = nil
	} else {
		a.State.NextRunOn = timestamppb.New(next)
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// setTriggerFalse sets the alert's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *AlertReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, false)
	if err != nil {
		return err
	}

	a := self.GetAlert()
	if a == nil {
		return fmt.Errorf("not an alert")
	}

	a.Spec.Trigger = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// executeAll runs queries and (maybe) sends emails for the alert. It also adds entries to a.State.ExecutionHistory.
// By default, an alert is checked once for the current watermark, but if a.Spec.IntervalsIsoDuration is set, it will be checked *for each* interval that has elapsed since the previous execution watermark.
func (r *AlertReconciler) executeAll(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, triggerTime time.Time, adhocTrigger bool) error {
	// Enforce timeout
	timeout := alertCheckDefaultTimeout
	if a.Spec.TimeoutSeconds > 0 {
		timeout = time.Duration(a.Spec.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get admin metadata for the alert (if an admin service is not configured, alerts will still work, the emails just won't have open/edit links).
	var adminMeta *drivers.AlertMetadata
	admin, release, err := r.C.Runtime.Admin(ctx, r.C.InstanceID)
	if err != nil && !errors.Is(err, runtime.ErrAdminNotConfigured) {
		return fmt.Errorf("failed to get admin client: %w", err)
	}
	if err == nil { // Connected successfully
		defer release()
		adminMeta, err = admin.GetAlertMetadata(ctx, a.Spec.QueryName, a.Spec.Annotations, a.Spec.GetQueryForUserId(), a.Spec.GetQueryForUserEmail())
		if err != nil {
			return fmt.Errorf("failed to get alert metadata: %w", err)
		}
	}

	// Run alert queries and send emails
	executeErr := r.executeAllWrapped(ctx, self, a, adminMeta, triggerTime, adhocTrigger)
	if executeErr == nil {
		return nil
	}

	// If it's a cancellation, don't add the error to the execution history.
	// The controller may for example cancel if the runtime is restarting or the underlying source is scheduled to refresh.
	if errors.Is(executeErr, context.Canceled) {
		// If there's a CurrentExecution, pretend it never happened
		if a.State.CurrentExecution != nil {
			a.State.CurrentExecution = nil
			err := r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				return err
			}
		}
		return executeErr
	}

	// There was an execution error. Add it to the execution history.
	if a.State.CurrentExecution == nil {
		// CurrentExecution will only be nil if we never made it to the point of checking the alert query.
		a.State.CurrentExecution = &runtimev1.AlertExecution{
			Adhoc:         adhocTrigger,
			ExecutionTime: nil, // NOTE: Setting execution time to nil. The only alternative is using triggerTime, but a) it might not be the executionTime, b) it might lead to previousWatermark being advanced too far on the next invocation.
			StartedOn:     timestamppb.Now(),
		}
	}
	a.State.CurrentExecution.Result = &runtimev1.AssertionResult{
		Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
		ErrorMessage: executeErr.Error(),
	}
	a.State.CurrentExecution.FinishedOn = timestamppb.Now()
	err = r.popCurrentExecution(ctx, self, a, adminMeta)
	if err != nil {
		return err
	}

	return executeErr
}

// executeAllWrapped is called by executeAll, which wraps it with timeout and writing of errors to the execution history.
func (r *AlertReconciler) executeAllWrapped(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, adminMeta *drivers.AlertMetadata, triggerTime time.Time, adhocTrigger bool) error {
	// Check refs
	err := checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return err
	}

	// Evaluate watermark unless refs check failed.
	watermark := triggerTime
	if a.Spec.WatermarkInherit {
		t, ok, err := r.computeInheritedWatermark(ctx, self.Meta.Refs)
		if err != nil {
			return err
		}
		if ok {
			watermark = t
		}
		// If !ok, no watermark could be computed. So we'll just stick to triggerTime.
	}

	// Evaluate previous watermark (if any)
	var previousWatermark time.Time
	for _, e := range a.State.ExecutionHistory {
		if e.ExecutionTime != nil {
			previousWatermark = e.ExecutionTime.AsTime()
			break
		}
	}

	// Evaluate intervals
	ts, err := calculateExecutionTimes(self, a, watermark, previousWatermark)
	if err != nil {
		// This should not usually error
		r.C.Logger.Error("Internal: failed to calculate execution times", zap.String("name", self.Meta.Name.Name), zap.Error(err))
		return err
	}

	if len(ts) == 0 {
		r.C.Logger.Debug("Skipped alert check because watermark has not advanced by a full interval", zap.String("name", self.Meta.Name.Name), zap.Time("current_watermark", watermark), zap.Time("previous_watermark", previousWatermark), zap.String("interval", a.Spec.IntervalsIsoDuration))
		return nil
	}

	// Evaluate alert for each execution time
	for _, t := range ts {
		err := r.executeSingle(ctx, self, a, adminMeta, t, adhocTrigger)
		if err != nil {
			return err
		}
	}

	return nil
}

// executeSingleAlert runs the alert query and maybe sends emails for a single execution time.
func (r *AlertReconciler) executeSingle(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, adminMeta *drivers.AlertMetadata, executionTime time.Time, adhocTrigger bool) error {
	// Create new execution and save in State.CurrentExecution
	a.State.CurrentExecution = &runtimev1.AlertExecution{
		Adhoc:         adhocTrigger,
		ExecutionTime: timestamppb.New(executionTime),
		StartedOn:     timestamppb.Now(),
	}
	err := r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return err
	}

	// Check the alert and get the result
	res, executeErr := r.executeSingleWrapped(ctx, self, a, adminMeta, executionTime)

	// If the error is a cancellation/timeout, return (will be retried)
	if errors.Is(executeErr, context.Canceled) || errors.Is(executeErr, context.DeadlineExceeded) {
		return executeErr
	}

	// The error is not a cancellation/timeout. Add it to the execution history. (We don't return it since we want to continue evaluating other execution times.)
	if executeErr != nil {
		res = &runtimev1.AssertionResult{
			Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
			ErrorMessage: fmt.Sprintf("Alert check failed: %s", executeErr.Error()),
		}
	}

	// Finalize and pop current execution.
	a.State.CurrentExecution.Result = res
	a.State.CurrentExecution.FinishedOn = timestamppb.Now()
	err = r.popCurrentExecution(ctx, self, a, adminMeta)
	if err != nil {
		return err
	}
	return nil
}

// checkAlert runs the alert query and returns the result.
func (r *AlertReconciler) executeSingleWrapped(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, adminMeta *drivers.AlertMetadata, executionTime time.Time) (*runtimev1.AssertionResult, error) {
	// Log
	r.C.Logger.Debug("Checking alert", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))

	// Build query proto
	qpb, err := queries.ProtoFromJSON(a.Spec.QueryName, a.Spec.QueryArgsJson, &executionTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Evaluate query attributes
	var queryForAttrs map[string]any
	if adminMeta != nil {
		queryForAttrs = adminMeta.QueryForAttributes
	}
	if a.Spec.GetQueryForAttributes() != nil { // Explicit attributes take precedence
		queryForAttrs = a.Spec.GetQueryForAttributes().AsMap()
	}

	// Create and execute query
	q, err := queries.ProtoToQuery(qpb, queryForAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = r.C.Runtime.Query(ctx, r.C.InstanceID, q, alertQueryPriority)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Extract result row
	row, ok, err := extractQueryResultFirstRow(q)
	if err != nil {
		return nil, fmt.Errorf("failed to extract query result: %w", err)
	}
	if !ok {
		r.C.Logger.Info("Alert passed", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))
		return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS}, nil
	}

	r.C.Logger.Info("Alert failed", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))

	// Return fail row
	failRow, err := structpb.NewStruct(row)
	if err != nil {
		return nil, fmt.Errorf("failed to convert fail row to proto: %w", err)
	}
	return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, FailRow: failRow}, nil
}

// popCurrentExecution moves the current execution into the execution history and sends emails if the execution matched the emailing criteria.
// At a certain limit, it trims old executions from the history to prevent it from growing unboundedly.
func (r *AlertReconciler) popCurrentExecution(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, adminMeta *drivers.AlertMetadata) error {
	if a.State.CurrentExecution == nil {
		panic(fmt.Errorf("attempting to pop current execution when there is none"))
	}

	current := a.State.CurrentExecution

	// td represents the amount of time since we last sent an email for the current status AND where all intervening executions have returned the same status.
	var td *time.Duration
	if current.ExecutionTime != nil {
		for _, prev := range a.State.ExecutionHistory {
			if prev.Result.Status != current.Result.Status {
				break
			}
			if !prev.SentEmails {
				continue
			}

			var prevT time.Time
			if prev.ExecutionTime != nil {
				prevT = prev.ExecutionTime.AsTime()
			} else {
				prevT = prev.FinishedOn.AsTime()
			}

			var currT time.Time
			if current.ExecutionTime != nil {
				currT = current.ExecutionTime.AsTime()
			} else {
				currT = current.FinishedOn.AsTime()
			}

			v := currT.Sub(prevT)
			td = &v
		}
	}

	// Determine if we should notify/renotify using td
	var notify bool
	if td == nil {
		// The status has changed since the last execution, so we should notify.
		// NOTE: This case may also match in an edge case of execution history limits, but that's fine.
		notify = true
	} else if a.Spec.EmailRenotify {
		if a.Spec.EmailRenotifyAfterSeconds == 0 {
			// The status has not changed since the last execution and there's no renotify suppression period, so we should notify.
			notify = true
		} else if int(td.Seconds()) >= int(a.Spec.EmailRenotifyAfterSeconds) {
			// The status has not changed since the last email and the last email was sent more than the renotify suppression period ago, so we should notify.
			notify = true
		}
	}

	// Get execution time
	var executionTime time.Time
	if current.ExecutionTime != nil {
		executionTime = current.ExecutionTime.AsTime()
	}

	// Generate the email message to send (if any)
	var msg *email.AlertStatus
	if notify {
		switch current.Result.Status {
		case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
			if !a.Spec.EmailOnRecover {
				break
			}

			// Check this is a recovery, i.e. that the previous status was something other than a PASS
			if len(a.State.ExecutionHistory) == 0 {
				break
			}
			prev := a.State.ExecutionHistory[0]
			if prev.Result.Status == runtimev1.AssertionStatus_ASSERTION_STATUS_PASS {
				break
			}

			msg = &email.AlertStatus{
				Title:         a.Spec.Title,
				ExecutionTime: executionTime,
				Status:        current.Result.Status,
				IsRecover:     true,
			}
		case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
			if !a.Spec.EmailOnFail {
				break
			}

			msg = &email.AlertStatus{
				Title:         a.Spec.Title,
				ExecutionTime: executionTime,
				Status:        current.Result.Status,
				FailRow:       current.Result.FailRow.AsMap(),
			}
		case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
			if !a.Spec.EmailOnError {
				break
			}

			msg = &email.AlertStatus{
				Title:          a.Spec.Title,
				ExecutionTime:  executionTime,
				Status:         current.Result.Status,
				ExecutionError: current.Result.ErrorMessage,
			}
		default:
			return fmt.Errorf("unexpected assertion result status: %v", current.Result.Status)
		}
	}

	// Send the email (if applicable)
	var emailErr error
	var sentEmails bool
	if msg != nil {
		if adminMeta != nil {
			// Note: adminMeta may not always be available (if outside of cloud). In those cases, we leave the links blank (no clickthrough available).
			msg.OpenLink = adminMeta.OpenURL
			msg.EditLink = adminMeta.EditURL
		}

		for _, recipient := range a.Spec.EmailRecipients {
			msg.ToEmail = recipient
			err := r.C.Runtime.Email.SendAlertStatus(msg)
			if err != nil {
				emailErr = fmt.Errorf("Failed to send email to %q: %w", recipient, err)
				break
			}
		}
		sentEmails = true
	}

	// If sending emails failed, add the error as an execution error.
	if emailErr != nil {
		a.State.CurrentExecution.Result = &runtimev1.AssertionResult{
			Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
			ErrorMessage: emailErr.Error(),
		}
	}

	a.State.CurrentExecution.SentEmails = sentEmails
	a.State.CurrentExecution.FinishedOn = timestamppb.Now()
	a.State.ExecutionHistory = slices.Insert(a.State.ExecutionHistory, 0, a.State.CurrentExecution)
	a.State.CurrentExecution = nil
	a.State.ExecutionCount++

	if len(a.State.ExecutionHistory) > alertExecutionHistoryLimit {
		a.State.ExecutionHistory = a.State.ExecutionHistory[:alertExecutionHistoryLimit]
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// computeInheritedWatermark computes the inherited watermark for the alert.
// It returns false if the watermark could not be computed.
func (r *AlertReconciler) computeInheritedWatermark(ctx context.Context, refs []*runtimev1.ResourceName) (time.Time, bool, error) {
	var t time.Time
	for _, ref := range refs {
		q := &queries.ResourceWatermark{
			ResourceKind: ref.Kind,
			ResourceName: ref.Name,
		}
		err := r.C.Runtime.Query(ctx, r.C.InstanceID, q, alertQueryPriority)
		if err != nil {
			return t, false, fmt.Errorf("failed to resolve watermark for %s/%s: %w", ref.Kind, ref.Name, err)
		}

		if q.Result != nil && (t.IsZero() || q.Result.Before(t)) {
			t = *q.Result
		}
	}

	return t, !t.IsZero(), nil
}

// calculateExecutionTimes calculates the execution times for an alert, taking into consideration the alert's intervals configuration and previous executions.
// If the alert is not configured to run on intervals, it will return a slice containing only the current watermark.
func calculateExecutionTimes(self *runtimev1.Resource, a *runtimev1.Alert, watermark, previousWatermark time.Time) ([]time.Time, error) {
	// If the alert is not configured to run on intervals, check it just for the current watermark.
	if a.Spec.IntervalsIsoDuration == "" {
		return []time.Time{watermark}, nil
	}

	// Note: The watermark and previousWatermark may be unaligned with the alert's interval duration.

	// Parse the interval duration
	// The YAML parser validates it as a StandardDuration, so this shouldn't fail.
	di, err := duration.ParseISO8601(a.Spec.IntervalsIsoDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval duration: %w", err)
	}
	d, ok := di.(duration.StandardDuration)
	if !ok {
		return nil, fmt.Errorf("interval duration %q is not a standard ISO 8601 duration", a.Spec.IntervalsIsoDuration)
	}

	// Extract time zone
	tz := time.UTC
	if a.Spec.RefreshSchedule != nil && a.Spec.RefreshSchedule.TimeZone != "" {
		tz, err = time.LoadLocation(a.Spec.RefreshSchedule.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("failed to load time zone %q: %w", a.Spec.RefreshSchedule.TimeZone, err)
		}
	}

	// Compute the last end time (rounded to the interval duration)
	// NOTE: Hardcoding firstDayOfWeek and firstMonthOfYear. We might consider inferring these from the underlying metrics view (or just customizing in the `intervals:` clause) in the future.
	end := watermark.In(tz)
	if a.Spec.IntervalsCheckUnclosed {
		// Ceil
		t := d.Truncate(end, 1, 1)
		if !t.Equal(end) {
			end = d.Add(t)
		}
	} else {
		// Floor
		end = d.Truncate(end, 1, 1)
	}

	// If there isn't a previous watermark, we'll just check the current watermark.
	if previousWatermark.IsZero() {
		return []time.Time{end}, nil
	}

	// Skip if end isn't past the previous watermark (unless we're supposed to check unclosed intervals)
	if !a.Spec.IntervalsCheckUnclosed && !end.After(previousWatermark) {
		return nil, nil
	}

	// Set a limit on the number of intervals to check
	limit := int(a.Spec.IntervalsLimit)
	if limit <= 0 {
		limit = alertDefaultIntervalsLimit
	}

	// Calculate the execution times
	ts := []time.Time{end}
	for i := 0; i < limit; i++ {
		t := ts[len(ts)-1]
		t = d.Sub(t)
		if !t.After(previousWatermark) {
			break
		}
		ts = append(ts, t)
	}

	// Reverse execution times so we run them in chronological order
	slices.Reverse(ts)

	return ts, nil
}

// extractQueryResultFirstRow extracts the first row from a query result.
// TODO: This should function more like an export, i.e. use dimension/measure labels instead of names.
func extractQueryResultFirstRow(q runtime.Query) (map[string]any, bool, error) {
	switch q := q.(type) {
	case *queries.MetricsViewAggregation:
		if q.Result != nil && len(q.Result.Data) > 0 {
			row := q.Result.Data[0]
			return row.AsMap(), true, nil
		}
		return nil, false, nil
	case *queries.MetricsViewComparison:
		if q.Result != nil && len(q.Result.Rows) > 0 {
			row := q.Result.Rows[0]
			res := make(map[string]any)
			res[q.DimensionName] = row.DimensionValue
			for _, v := range row.MeasureValues {
				res[v.MeasureName] = v.BaseValue.AsInterface()
				if v.ComparisonValue != nil {
					res[v.MeasureName+" (prev)"] = v.ComparisonValue.AsInterface()
				}
				if v.DeltaAbs != nil {
					res[v.MeasureName+" (Δ)"] = v.DeltaAbs.AsInterface()
				}
				if v.DeltaRel != nil {
					res[v.MeasureName+" (Δ%)"] = v.DeltaRel.AsInterface()
				}
			}
			return res, true, nil
		}
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("query type %T not supported for alerts", q)
	}
}
