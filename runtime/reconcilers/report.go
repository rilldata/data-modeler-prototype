package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/queries"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	reportExecutionHistoryLimit = 10
	reportCheckDefaultTimeout   = 5 * time.Minute
	reportDefaultIntervalsLimit = 25
	reportQueryPriority         = 1
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindReport, newReportReconciler)
}

type ReportReconciler struct {
	C *runtime.Controller
}

func newReportReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ReportReconciler{C: c}, nil
}

func (r *ReportReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ReportReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetReport()
	b := to.GetReport()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ReportReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetReport()
	b := to.GetReport()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ReportReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetReport().State = &runtimev1.ReportState{}
	return nil
}

func (r *ReportReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	rep := self.GetReport()
	if rep == nil {
		return runtime.ReconcileResult{Err: errors.New("not a report")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// NOTE: refs not supported for reports.
	// Not supporting them simplifies report spec generation, improves performance (there may be many reports),
	// and it's anyway acceptable (maybe even expected) that a report fails with an execution error if the underlying query errors.

	// If CurrentExecution is not nil, a catastrophic failure occurred during the last execution.
	// Clean up to ensure CurrentExecution is nil.
	if rep.State.CurrentExecution != nil {
		rep.State.CurrentExecution.ErrorMessage = "Internal: report execution was interrupted unexpectedly"
		rep.State.CurrentExecution.FinishedOn = timestamppb.Now()
		err = r.popCurrentExecution(ctx, self, rep)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Exit early if disabled
	if rep.Spec.RefreshSchedule != nil && rep.Spec.RefreshSchedule.Disable {
		return runtime.ReconcileResult{}
	}

	// Determine whether to trigger
	adhocTrigger := rep.Spec.Trigger
	scheduleTrigger := rep.State.NextRunOn != nil && !rep.State.NextRunOn.AsTime().After(time.Now())
	trigger := adhocTrigger || scheduleTrigger

	// If not triggering now, update NextRunOn and retrigger when it falls due
	if !trigger {
		err = r.updateNextRunOn(ctx, self, rep)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		if rep.State.NextRunOn != nil {
			return runtime.ReconcileResult{Retrigger: rep.State.NextRunOn.AsTime()}
		}
		return runtime.ReconcileResult{}
	}

	// Determine time to evaluate the report relative to.
	// We use the "clean" scheduled time unless it's an ad-hoc trigger.
	var reportTime *timestamppb.Timestamp
	if scheduleTrigger && !adhocTrigger {
		reportTime = rep.State.NextRunOn
	} else {
		reportTime = timestamppb.Now()
	}

	retry, executeErr := r.executeAll(ctx, self, rep, reportTime.AsTime(), adhocTrigger)

	// If we want to retry, exit without advancing NextRunOn or clearing spec.Trigger.
	// NOTE: We don't set Retrigger here because we'll leave re-scheduling to whatever cancelled the reconciler.
	if retry || errors.Is(executeErr, context.Canceled) {
		return runtime.ReconcileResult{Err: executeErr}
	}

	// Advance NextRunOn
	err = r.updateNextRunOn(ctx, self, rep)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Clear ad-hoc trigger
	if rep.Spec.Trigger {
		err := r.setTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to clear trigger: %w", err)}
		}
	}

	// Done
	if rep.State.NextRunOn != nil {
		return runtime.ReconcileResult{Err: executeErr, Retrigger: rep.State.NextRunOn.AsTime()}
	}
	return runtime.ReconcileResult{Err: executeErr}
}

// updateNextRunOn evaluates the report's schedule relative to the current time, and updates the NextRunOn state accordingly.
// If the schedule is nil, it will set NextRunOn to nil.
func (r *ReportReconciler) updateNextRunOn(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report) error {
	next, err := nextRefreshTime(time.Now(), rep.Spec.RefreshSchedule)
	if err != nil {
		return err
	}

	var curr time.Time
	if rep.State.NextRunOn != nil {
		curr = rep.State.NextRunOn.AsTime()
	}

	if next == curr {
		return nil
	}

	if next.IsZero() {
		rep.State.NextRunOn = nil
	} else {
		rep.State.NextRunOn = timestamppb.New(next)
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// popCurrentExecution moves the current execution into the execution history, and persists the updated state.
// At a certain limit, it trims old executions from the history to prevent it from growing unboundedly.
func (r *ReportReconciler) popCurrentExecution(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report) error {
	if rep.State.CurrentExecution == nil {
		panic(fmt.Errorf("attempting to pop current execution when there is none"))
	}

	rep.State.ExecutionHistory = slices.Insert(rep.State.ExecutionHistory, 0, rep.State.CurrentExecution)
	rep.State.CurrentExecution = nil

	if len(rep.State.ExecutionHistory) > reportExecutionHistoryLimit {
		rep.State.ExecutionHistory = rep.State.ExecutionHistory[:reportExecutionHistoryLimit]
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// setTriggerFalse sets the report's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *ReportReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return err
	}

	rep := self.GetReport()
	if rep == nil {
		return fmt.Errorf("not a report")
	}

	rep.Spec.Trigger = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// executeAll runs queries and sends reports. It also adds entries to rep.State.ExecutionHistory.
// By default, a report is checked once for the current watermark, but if rep.Spec.IntervalsIsoDuration is set, it will be checked *for each* interval that has elapsed since the previous execution watermark.
func (r *ReportReconciler) executeAll(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report, triggerTime time.Time, adhocTrigger bool) (bool, error) {
	// Enforce timeout
	timeout := reportCheckDefaultTimeout
	if rep.Spec.TimeoutSeconds > 0 {
		timeout = time.Duration(rep.Spec.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Run report queries and send notifications
	retry, executeErr := r.executeAllWrapped(ctx, self, rep, triggerTime, adhocTrigger)
	if executeErr == nil {
		return false, nil
	}

	// If it's a cancellation, don't add the error to the execution history.
	// The controller may for example cancel if the runtime is restarting or the underlying source is scheduled to refresh.
	if retry || errors.Is(executeErr, context.Canceled) {
		// If there's a CurrentExecution, pretend it never happened
		if rep.State.CurrentExecution != nil {
			rep.State.CurrentExecution = nil
			err := r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				return false, err
			}
		}
		return retry, executeErr
	}

	// There was an execution error. Add it to the execution history.
	if rep.State.CurrentExecution == nil {
		// CurrentExecution will only be nil if we never made it to the point of checking the report query.
		rep.State.CurrentExecution = &runtimev1.ReportExecution{
			Adhoc:      adhocTrigger,
			ReportTime: nil, // NOTE: Setting execution time to nil. The only alternative is using triggerTime, but a) it might not be the reportTime, b) it might lead to previousWatermark being advanced too far on the next invocation.
			StartedOn:  timestamppb.Now(),
		}
	}
	rep.State.CurrentExecution.ErrorMessage = executeErr.Error()
	rep.State.CurrentExecution.FinishedOn = timestamppb.Now()
	err := r.popCurrentExecution(ctx, self, rep)
	if err != nil {
		return false, err
	}

	return retry, executeErr
}

// executeAllWrapped is called by executeAll, which wraps it with timeout and writing of errors to the execution history.
func (r *ReportReconciler) executeAllWrapped(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report, triggerTime time.Time, adhocTrigger bool) (bool, error) {
	// Check refs
	err := checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return false, err
	}

	// Evaluate watermark unless refs check failed.
	watermark := triggerTime
	if rep.Spec.WatermarkInherit {
		t, ok, err := r.computeInheritedWatermark(ctx, self.Meta.Refs)
		if err != nil {
			return false, err
		}
		if ok {
			watermark = t
		}
		// If !ok, no watermark could be computed. So we'll just stick to triggerTime.
	}

	// Evaluate previous watermark (if any)
	var previousWatermark time.Time
	for _, e := range rep.State.ExecutionHistory {
		if e.ReportTime != nil {
			previousWatermark = e.ReportTime.AsTime()
			break
		}
	}

	// Evaluate intervals
	ts, err := calculateReportExecutionTimes(rep, watermark, previousWatermark)
	if err != nil {
		skipErr := &skipError{}
		if errors.As(err, skipErr) {
			r.C.Logger.Info("Skipped report", zap.String("name", self.Meta.Name.Name), zap.String("reason", skipErr.reason), zap.Time("current_watermark", watermark), zap.Time("previous_watermark", previousWatermark), zap.String("interval", rep.Spec.IntervalsIsoDuration))
			return false, nil
		}
		r.C.Logger.Error("Internal: failed to calculate execution times", zap.String("name", self.Meta.Name.Name), zap.Error(err))
		return false, err
	}
	if len(ts) == 0 {
		// This should never happen
		r.C.Logger.Error("Internal: no execution times found", zap.String("name", self.Meta.Name.Name), zap.Error(err))
		return false, nil
	}

	// Evaluate report for each execution time
	for _, t := range ts {
		retry, err := r.executeSingle(ctx, self, rep, t, adhocTrigger)
		if err != nil {
			return retry, err
		}
	}

	return false, nil
}

// executeSingle runs the report query and sends notifications for a single execution time.
func (r *ReportReconciler) executeSingle(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report, executionTime time.Time, adhocTrigger bool) (bool, error) {
	// Create new execution and save in State.CurrentExecution
	rep.State.CurrentExecution = &runtimev1.ReportExecution{
		Adhoc:      adhocTrigger,
		ReportTime: timestamppb.New(executionTime),
		StartedOn:  timestamppb.Now(),
	}
	err := r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return false, err
	}

	// Execute report
	dirtyErr, reportErr := r.sendReport(ctx, self, rep, executionTime)

	// Set execution error and determine whether to retry.
	// We're only going to retry on non-dirty cancellations.
	retry := false
	if reportErr != nil {
		if errors.Is(reportErr, context.Canceled) || errors.Is(reportErr, context.DeadlineExceeded) {
			if dirtyErr {
				rep.State.CurrentExecution.ErrorMessage = "Report run was interrupted after some notifications were sent. The report will not automatically retry."
			} else {
				retry = true
				rep.State.CurrentExecution.ErrorMessage = "Report run was interrupted. It will automatically retry."
			}
		} else {
			rep.State.CurrentExecution.ErrorMessage = fmt.Sprintf("Report run failed: %v", reportErr.Error())
		}
		reportErr = fmt.Errorf("last report run failed with error: %v", reportErr.Error())
	}

	// Log it
	if reportErr != nil {
		r.C.Logger.Error("Report run failed", zap.Any("report", self.Meta.Name), zap.Any("error", reportErr.Error()))
	}

	// Commit CurrentExecution to history
	rep.State.CurrentExecution.FinishedOn = timestamppb.Now()
	err = r.popCurrentExecution(ctx, self, rep)
	if err != nil {
		return false, err
	}

	return retry, reportErr
}

// sendReport composes and sends the actual report to the configured recipients.
// It returns true if an error occurred after some or all notifications were sent.
func (r *ReportReconciler) sendReport(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report, t time.Time) (bool, error) {
	r.C.Logger.Info("Sending report", zap.String("report", self.Meta.Name.Name), zap.Time("report_time", t))

	admin, release, err := r.C.Runtime.Admin(ctx, r.C.InstanceID)
	if err != nil {
		if errors.Is(err, runtime.ErrAdminNotConfigured) {
			r.C.Logger.Info("Skipped sending report because an admin service is not configured", zap.String("report", self.Meta.Name.Name))
			return false, nil
		}
		return false, fmt.Errorf("failed to get admin client: %w", err)
	}
	defer release()

	var ownerID, explore, canvas, webOpenMode string
	if id, ok := rep.Spec.Annotations["admin_owner_user_id"]; ok {
		ownerID = id
	}
	if e, ok := rep.Spec.Annotations["explore"]; ok {
		explore = e
	}
	if c, ok := rep.Spec.Annotations["canvas"]; ok {
		canvas = c
	}
	if w, ok := rep.Spec.Annotations["web_open_mode"]; ok {
		webOpenMode = w
	} else {
		webOpenMode = "legacy" // backward compatibility
		if _, ok = rep.Spec.Annotations["web_open_path"]; !ok {
			webOpenMode = "none" // for older reports if web_open_path is not set
		}
	}

	var reqWebOpenMode adminv1.ReportOptions_OpenMode
	switch webOpenMode {
	case "legacy":
		reqWebOpenMode = adminv1.ReportOptions_OPEN_MODE_LEGACY
	case "creator":
		reqWebOpenMode = adminv1.ReportOptions_OPEN_MODE_CREATOR
	case "none":
		reqWebOpenMode = adminv1.ReportOptions_OPEN_MODE_NONE
	case "filtered":
		reqWebOpenMode = adminv1.ReportOptions_OPEN_MODE_FILTERED
	default:
		return false, fmt.Errorf("invalid web open mode: %s", webOpenMode)
	}

	anonRecipients := false
	var emailRecipients []string
	for _, notifier := range rep.Spec.Notifiers {
		if notifier.Connector == "email" {
			emailRecipients = pbutil.ToSliceString(notifier.Properties.AsMap()["recipients"])
		} else {
			anonRecipients = true
		}
	}

	meta, err := admin.GetReportMetadata(ctx, self.Meta.Name.Name, ownerID, explore, canvas, reqWebOpenMode, emailRecipients, anonRecipients, t)
	if err != nil {
		return false, fmt.Errorf("failed to get report metadata: %w", err)
	}

	sent := false
	for _, notifier := range rep.Spec.Notifiers {
		switch notifier.Connector {
		case "email":
			recipients := pbutil.ToSliceString(notifier.Properties.AsMap()["recipients"])
			for _, recipient := range recipients {
				opts := &email.ScheduledReport{
					ToEmail:        recipient,
					ToName:         "",
					DisplayName:    rep.Spec.DisplayName,
					ReportTime:     t,
					DownloadFormat: formatExportFormat(rep.Spec.ExportFormat),
				}
				urls, ok := meta.RecipientURLs[recipient]
				if !ok {
					return false, fmt.Errorf("failed to get recipient URLs for %q", recipient)
				}
				opts.OpenLink = urls.OpenURL
				u, err := createExportURL(urls.ExportURL, rep.Spec.ExportFormat.String(), t, int(rep.Spec.ExportLimit))
				if err != nil {
					return false, err
				}
				opts.DownloadLink = u.String()
				opts.EditLink = urls.EditURL
				opts.UnsubscribeLink = urls.UnsubscribeURL
				err = r.C.Runtime.Email.SendScheduledReport(opts)
				sent = true
				if err != nil {
					return true, fmt.Errorf("failed to generate report for %q: %w", recipient, err)
				}
			}
		default:
			err := func() (outErr error) {
				conn, release, err := r.C.Runtime.AcquireHandle(ctx, r.C.InstanceID, notifier.Connector)
				if err != nil {
					return err
				}
				defer release()
				n, err := conn.AsNotifier(notifier.Properties.AsMap())
				if err != nil {
					return err
				}
				urls, ok := meta.RecipientURLs[""]
				if !ok {
					return fmt.Errorf("failed to get recipient URLs for anon user")
				}
				u, err := createExportURL(urls.ExportURL, rep.Spec.ExportFormat.String(), t, int(rep.Spec.ExportLimit))
				if err != nil {
					return err
				}
				msg := &drivers.ScheduledReport{
					DisplayName:     rep.Spec.DisplayName,
					ReportTime:      t,
					DownloadFormat:  formatExportFormat(rep.Spec.ExportFormat),
					OpenLink:        urls.OpenURL,
					DownloadLink:    u.String(),
					UnsubscribeLink: urls.UnsubscribeURL,
				}
				start := time.Now()
				defer func() {
					totalLatency := time.Since(start).Milliseconds()

					if r.C.Activity != nil {
						r.C.Activity.RecordMetric(ctx, "notifier_total_latency_ms", float64(totalLatency),
							attribute.Bool("failed", outErr != nil),
							attribute.String("connector", notifier.Connector),
							attribute.String("notification_type", "scheduled_report"),
						)
					}
				}()
				err = n.SendScheduledReport(msg)
				sent = true
				if err != nil {
					return fmt.Errorf("failed to send %s notification: %w", notifier.Connector, err)
				}
				return nil
			}()
			if err != nil {
				return sent, err
			}
		}
	}

	return false, nil
}

func createExportURL(inURL, format string, executionTime time.Time, exportLimit int) (*url.URL, error) {
	exportURL, err := url.Parse(inURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse export URL %q: %w", inURL, err)
	}

	exportURLQry := exportURL.Query()
	exportURLQry.Set("format", format)
	exportURLQry.Set("execution_time", executionTime.Format(time.RFC3339))
	if exportLimit > 0 {
		exportURLQry.Set("limit", strconv.Itoa(exportLimit))
	}
	exportURL.RawQuery = exportURLQry.Encode()

	return exportURL, nil
}

func formatExportFormat(f runtimev1.ExportFormat) string {
	switch f {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return "CSV"
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return "Excel"
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return "Parquet"
	default:
		return f.String()
	}
}

// computeInheritedWatermark computes the inherited watermark for the report.
// It returns false if the watermark could not be computed.
func (r *ReportReconciler) computeInheritedWatermark(ctx context.Context, refs []*runtimev1.ResourceName) (time.Time, bool, error) {
	var t time.Time
	for _, ref := range refs {
		q := &queries.ResourceWatermark{
			ResourceKind: ref.Kind,
			ResourceName: ref.Name,
		}
		err := r.C.Runtime.Query(ctx, r.C.InstanceID, q, reportQueryPriority)
		if err != nil {
			return t, false, fmt.Errorf("failed to resolve watermark for %s/%s: %w", ref.Kind, ref.Name, err)
		}

		if q.Result != nil && (t.IsZero() || q.Result.Before(t)) {
			t = *q.Result
		}
	}

	return t, !t.IsZero(), nil
}

// calculateReportExecutionTimes calculates the execution times for a report, taking into consideration the report's intervals configuration and previous executions.
// If the report is not configured to run on intervals, it will return a slice containing only the current watermark.
// If the alert should not be executed, it returns a skipError explaining why.
func calculateReportExecutionTimes(r *runtimev1.Report, watermark, previousWatermark time.Time) ([]time.Time, error) {
	// If the watermark is unchanged, skip the check.
	// NOTE: It might make sense to make this configurable in the future, but the use cases seem limited.
	// The watermark can only be unchanged if watermark="inherit" and since that indicates watermarks can be trusted, why check for the same watermark?
	if watermark.Equal(previousWatermark) {
		return nil, skipError{reason: "watermark is unchanged"}
	}

	// If the report is not configured to run on intervals, check it just for the current watermark.
	if r.Spec.IntervalsIsoDuration == "" {
		return []time.Time{watermark}, nil
	}

	// Note: The watermark and previousWatermark may be unaligned with the report's interval duration.

	// Parse the interval duration
	// The YAML parser validates it as a StandardDuration, so this shouldn't fail.
	di, err := duration.ParseISO8601(r.Spec.IntervalsIsoDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval duration: %w", err)
	}
	d, ok := di.(duration.StandardDuration)
	if !ok {
		return nil, fmt.Errorf("interval duration %q is not a standard ISO 8601 duration", r.Spec.IntervalsIsoDuration)
	}

	// Extract time zone
	tz := time.UTC
	if r.Spec.RefreshSchedule != nil && r.Spec.RefreshSchedule.TimeZone != "" {
		tz, err = time.LoadLocation(r.Spec.RefreshSchedule.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("failed to load time zone %q: %w", r.Spec.RefreshSchedule.TimeZone, err)
		}
	}

	// Compute the last end time (rounded to the interval duration)
	// NOTE: Hardcoding firstDayOfWeek and firstMonthOfYear. We might consider inferring these from the underlying metrics view (or just customizing in the `intervals:` clause) in the future.
	end := watermark.In(tz)
	if r.Spec.IntervalsCheckUnclosed {
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
	if !r.Spec.IntervalsCheckUnclosed && !end.After(previousWatermark) {
		return nil, skipError{reason: "watermark has not advanced by a full interval"}
	}

	// Set a limit on the number of intervals to check
	limit := int(r.Spec.IntervalsLimit)
	if limit <= 0 {
		limit = reportDefaultIntervalsLimit
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
