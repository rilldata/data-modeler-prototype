package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// If changing this value also update maxAcquiredConnDuration in runtime/drivers/duckdb/duckdb.go
	_modelDefaultTimeout = 3 * time.Hour

	_modelSyncPartitionsBatchSize    = 1000
	_modelPendingPartitionsBatchSize = 1000
)

var errPartitionsHaveErrors = errors.New("some partitions have errors")

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindModel, newModelReconciler)
}

type ModelReconciler struct {
	C       *runtime.Controller
	execSem *semaphore.Weighted
}

func newModelReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	cfg, err := c.Runtime.InstanceConfig(ctx, c.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model execution concurrency limit: %w", err)
	}
	if cfg.ModelConcurrentExecutionLimit <= 0 {
		return nil, errors.New("model_concurrent_execution_limit must be greater than zero")
	}
	return &ModelReconciler{
		C:       c,
		execSem: semaphore.NewWeighted(int64(cfg.ModelConcurrentExecutionLimit)),
	}, nil
}

func (r *ModelReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ModelReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetModel()
	b := to.GetModel()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ModelReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetModel()
	b := to.GetModel()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ModelReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetModel().State = &runtimev1.ModelState{}
	return nil
}

func (r *ModelReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	model := self.GetModel()
	if model == nil {
		return runtime.ReconcileResult{Err: errors.New("not a model")}
	}

	// If the model's state indicates that the last execution produced valid output, create a manager for the previous result
	var prevManager drivers.ModelManager
	var prevResult *drivers.ModelResult
	if model.State.ResultConnector != "" {
		conn, release, err := r.C.AcquireConn(ctx, model.State.ResultConnector)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		defer release()

		m, ok := conn.AsModelManager(r.C.InstanceID)
		if !ok {
			return runtime.ReconcileResult{Err: fmt.Errorf("connector %q does not support model management", model.State.ResultConnector)}
		}
		prevManager = m

		prevResult = &drivers.ModelResult{
			Connector:  model.State.ResultConnector,
			Properties: model.State.ResultProperties.AsMap(),
			Table:      model.State.ResultTable,
		}
	}

	// Fetch contextual config
	modelEnv, err := r.newModelEnv(ctx)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Handle deletion
	if self.Meta.DeletedOn != nil {
		if prevManager != nil {
			err := r.execSem.Acquire(ctx, 1)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
			defer r.execSem.Release(1)

			err = prevManager.Delete(ctx, prevResult)
			return runtime.ReconcileResult{Err: err}
		}

		err := r.clearPartitions(ctx, model)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}

		return runtime.ReconcileResult{}
	}

	// Handle renames
	if self.Meta.RenamedFrom != nil {
		if prevManager != nil {
			// Using a nested scope to ensure the execSem is safely acquired and released.
			func() {
				err := r.execSem.Acquire(ctx, 1)
				if err != nil {
					return // Safe to ignore because the err can only be ctx.Err()
				}
				defer r.execSem.Release(1)

				renameRes, err := prevManager.Rename(ctx, prevResult, self.Meta.Name.Name, modelEnv)
				if err == nil {
					err = r.updateStateWithResult(ctx, self, renameRes)
				}
				if err != nil {
					r.C.Logger.Warn("failed to rename model", zap.String("model", n.Name), zap.String("renamed_from", self.Meta.RenamedFrom.Name), zap.Error(err))
				}
			}()
			if ctx.Err() != nil { // Handle if the error was a ctx error
				return runtime.ReconcileResult{Err: ctx.Err()}
			}

			// Note: Not exiting early. We may need to retrigger the model in some cases. We also need to set the correct retrigger time.
		}
	}

	// Exit early if disabled
	if model.Spec.RefreshSchedule != nil && model.Spec.RefreshSchedule.Disable {
		return runtime.ReconcileResult{}
	}

	// Check refs - stop if any of them are invalid
	err = checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		// If not staging changes, we need to drop the previous output (if any) before returning
		if !modelEnv.StageChanges && prevManager != nil {
			err := r.execSem.Acquire(ctx, 1)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
			defer r.execSem.Release(1)

			err2 := prevManager.Delete(ctx, prevResult)
			if err2 != nil {
				r.C.Logger.Warn("failed to delete model output", zap.String("model", n.Name), zap.Error(err2))
			}

			err = r.clearPartitions(ctx, model)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}

			err2 = r.updateStateClear(ctx, self)
			if err2 != nil {
				r.C.Logger.Warn("refs check: failed to update state", zap.Any("error", err2))
			}
		}

		return runtime.ReconcileResult{Err: err}
	}

	// Compute hashes to determine if something has changes.
	// If the specHash changes, a full model reset is required (because the config changed).
	// If the refsHash changes, an incremental model run is sufficient (because the refs only went through a regular refresh).
	specHash, err := r.executionSpecHash(ctx, self.Meta.Refs, model.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute spec hash: %w", err)}
	}
	refsHash, err := r.refsStateHash(ctx, self.Meta.Refs, model.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute refs hash: %w", err)}
	}

	// Compute next time to refresh based on the RefreshSchedule (if any)
	var refreshOn time.Time
	if model.State.RefreshedOn != nil {
		refreshOn, err = nextRefreshTime(model.State.RefreshedOn.AsTime(), model.Spec.RefreshSchedule)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Check if the output still exists (might have been corrupted/lost somehow)
	var exists bool
	if prevManager != nil {
		exists, err = prevManager.Exists(ctx, prevResult)
		if err != nil {
			r.C.Logger.Warn("failed to check if model output exists", zap.String("model", n.Name), zap.Error(err))
		}
	}

	// Decide if we should trigger a reset
	triggerReset := model.Spec.TriggerFull
	triggerReset = triggerReset || model.State.ResultConnector == "" // If its nil, ResultProperties/ResultTable will also be nil
	triggerReset = triggerReset || model.State.RefreshedOn == nil
	triggerReset = triggerReset || model.State.SpecHash != specHash
	triggerReset = triggerReset || !exists

	// Decide if we should trigger
	trigger := triggerReset
	trigger = trigger || model.Spec.Trigger
	trigger = trigger || !refreshOn.IsZero() && time.Now().After(refreshOn)
	trigger = trigger || model.State.RefsHash != refsHash

	// Reschedule if we're not triggering
	if !trigger {
		// Show if any partitions errored
		if model.State.PartitionsHaveErrors {
			return runtime.ReconcileResult{Err: errPartitionsHaveErrors, Retrigger: refreshOn}
		}
		return runtime.ReconcileResult{Retrigger: refreshOn}
	}

	// Acquire the execution semaphore for the remainder of the function.
	err = r.execSem.Acquire(ctx, 1)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	defer r.execSem.Release(1)

	// If the output connector has changed, drop data in the old output connector (if any).
	// If only the output properties have changed, the executor will handle dropping existing data (to comply with StageChanges).
	if prevManager != nil && model.State.ResultConnector != model.Spec.OutputConnector {
		err = prevManager.Delete(ctx, prevResult)
		if err != nil {
			r.C.Logger.Warn("failed to delete model output", zap.String("model", n.Name), zap.Error(err))
		}
	}

	// Build the model
	executorConnector, execRes, execErr := r.executeAll(ctx, self, model, modelEnv, triggerReset, prevResult)

	// After the model has executed successfully, we re-evaluate the model's incremental state (not to be confused with the resource state)
	var newIncrementalState *structpb.Struct
	var newIncrementalStateSchema *runtimev1.StructType
	if execErr == nil {
		newIncrementalState, newIncrementalStateSchema, execErr = r.resolveIncrementalState(ctx, model)
	}

	// If the model is partitioned, track if any of the partitions have errors
	var partitionsHaveErrors bool
	if model.State.PartitionsModelId != "" {
		catalog, release, err := r.C.Runtime.Catalog(ctx, r.C.InstanceID)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		defer release()

		partitionsHaveErrors, err = catalog.CheckModelPartitionsHaveErrors(ctx, model.State.PartitionsModelId)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// If the build succeeded, update the model's state accodingly
	if execErr == nil {
		model.State.ExecutorConnector = executorConnector
		model.State.SpecHash = specHash
		model.State.RefsHash = refsHash
		model.State.RefreshedOn = timestamppb.Now()
		model.State.IncrementalState = newIncrementalState
		model.State.IncrementalStateSchema = newIncrementalStateSchema
		model.State.PartitionsHaveErrors = partitionsHaveErrors
		err := r.updateStateWithResult(ctx, self, execRes)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// If the build failed, clear the state only if we're not staging changes
	if execErr != nil {
		if !modelEnv.StageChanges {
			err := r.clearPartitions(ctx, model)
			if err != nil {
				return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
			}

			err = r.updateStateClear(ctx, self)
			if err != nil {
				return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
			}
		}
	}

	// If the context was cancelled, we return now since we don't want to clear the trigger or set a next refresh time.
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: errors.Join(ctx.Err(), execErr)}
	}

	// Reset spec.Trigger and spec.TriggerFull
	if model.Spec.Trigger || model.Spec.TriggerFull {
		err := r.updateTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
		}
	}

	// Compute next refresh time
	refreshOn, err = nextRefreshTime(time.Now(), model.Spec.RefreshSchedule)
	if err != nil {
		return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
	}

	// Note: If the build failed, this is where we return the error.
	if execErr != nil {
		return runtime.ReconcileResult{Err: execErr, Retrigger: refreshOn}
	}

	// Show if any partitions errored
	if model.State.PartitionsHaveErrors {
		return runtime.ReconcileResult{Err: errPartitionsHaveErrors, Retrigger: refreshOn}
	}

	// Return the next refresh time
	return runtime.ReconcileResult{Retrigger: refreshOn}
}

// executionSpecHash computes a hash of those model properties that impact execution.
// It also incorporates the spec hashes of the model's refs.
// If the spec hash changes, it means the model should be reset and fully re-executed.
func (r *ModelReconciler) executionSpecHash(ctx context.Context, refs []*runtimev1.ResourceName, spec *runtimev1.ModelSpec) (string, error) {
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

		if ref.Kind != runtime.ResourceKindSource && ref.Kind != runtime.ResourceKindModel {
			continue
		}

		r, err := r.C.Get(ctx, ref, false)
		if err != nil {
			continue
		}

		var refSpechHash string
		switch ref.Kind {
		case runtime.ResourceKindSource:
			refSpechHash = r.GetSource().State.SpecHash
		case runtime.ResourceKindModel:
			refSpechHash = r.GetModel().State.SpecHash
		}

		_, err = hash.Write([]byte(refSpechHash))
		if err != nil {
			return "", err
		}
	}

	err := binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.Incremental)
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.IncrementalStateResolver))
	if err != nil {
		return "", err
	}

	if spec.IncrementalStateResolverProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.IncrementalStateResolverProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.IncrementalStateResolverProperties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}

	_, err = hash.Write([]byte(spec.PartitionsResolver))
	if err != nil {
		return "", err
	}

	if spec.PartitionsResolverProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.PartitionsResolverProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.PartitionsResolverProperties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}

	_, err = hash.Write([]byte(spec.PartitionsWatermarkField))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.InputConnector))
	if err != nil {
		return "", err
	}

	if spec.InputProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.InputProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.InputProperties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}

	_, err = hash.Write([]byte(spec.OutputConnector))
	if err != nil {
		return "", err
	}

	if spec.OutputProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.OutputProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.OutputProperties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// refsStateHash computes a hash of the state of the model's refs.
// It is used to check if the model's refs have been updated, which should trigger an (incremental) model execution.
// (Note that the refs state hash identifies when to trigger incremental runs, whereas the the execution spec hash identifies when to trigger full resets.)
func (r *ModelReconciler) refsStateHash(ctx context.Context, refs []*runtimev1.ResourceName, spec *runtimev1.ModelSpec) (string, error) {
	if spec.RefreshSchedule == nil || !spec.RefreshSchedule.RefUpdate {
		return "", nil
	}

	hash := md5.New()

	for _, ref := range refs {
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

// updateStateWithResult updates the model resource's state with the result of a model execution.
// It only updates the result-related fields. If changing other fields, such as RefreshedOn and SpecHash, they must be assigned before calling this function.
func (r *ModelReconciler) updateStateWithResult(ctx context.Context, self *runtimev1.Resource, res *drivers.ModelResult) error {
	mdl := self.GetModel()

	props, err := structpb.NewStruct(res.Properties)
	if err != nil {
		return fmt.Errorf("failed to serialize result properties: %w", err)
	}

	mdl.State.ResultConnector = res.Connector
	mdl.State.ResultProperties = props
	mdl.State.ResultTable = res.Table

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// updateStateClear clears the model resource's state.
func (r *ModelReconciler) updateStateClear(ctx context.Context, self *runtimev1.Resource) error {
	mdl := self.GetModel()

	mdl.State.ExecutorConnector = ""
	mdl.State.ResultConnector = ""
	mdl.State.ResultProperties = nil
	mdl.State.ResultTable = ""
	mdl.State.SpecHash = ""
	mdl.State.RefsHash = ""
	mdl.State.RefreshedOn = nil
	mdl.State.IncrementalState = nil
	mdl.State.IncrementalStateSchema = nil
	mdl.State.PartitionsModelId = ""
	mdl.State.PartitionsHaveErrors = false

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// updateTriggerFalse sets the model's spec.Trigger and spec.TriggerFull to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *ModelReconciler) updateTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return err
	}

	model := self.GetModel()
	if model == nil {
		return fmt.Errorf("not a model")
	}

	model.Spec.Trigger = false
	model.Spec.TriggerFull = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// resolveIncrementalState resolves the incremental state of a model using its configured incremental state resolver.
// Note the ambiguity around "state" in models – all resources have a "spec" and a "state",
// but models also have a resolver for "incremental state" that enables incremental/stateful computation by persisting data from the previous execution.
// It returns nil results if an incremental state resolver is not configured or does not return any data.
func (r *ModelReconciler) resolveIncrementalState(ctx context.Context, mdl *runtimev1.ModelV2) (*structpb.Struct, *runtimev1.StructType, error) {
	if !mdl.Spec.Incremental {
		return nil, nil, nil
	}

	if mdl.Spec.IncrementalStateResolver == "" {
		return nil, nil, nil
	}

	res, err := r.C.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         r.C.InstanceID,
		Resolver:           mdl.Spec.IncrementalStateResolver,
		ResolverProperties: mdl.Spec.IncrementalStateResolverProperties.AsMap(),
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	if err != nil {
		return nil, nil, err
	}
	defer res.Close()

	row, err := res.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Not returning any rows will clear the state
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to read state resolver output: %w", err)
	}

	state, err := structpb.NewStruct(row)
	if err != nil {
		return nil, nil, fmt.Errorf("state resolver produced invalid output: %w", err)
	}

	return state, res.Schema(), nil
}

// resolveAndSyncPartitions resolves the model's partitions using its configured partitions resolver and inserts or updates them in the catalog.
func (r *ModelReconciler) resolveAndSyncPartitions(ctx context.Context, self *runtimev1.Resource, mdl *runtimev1.ModelV2, incrementalState map[string]any) error {
	// Log
	r.C.Logger.Debug("Resolving model partitions", zap.String("model", self.Meta.Name.Name), zap.String("resolver", mdl.Spec.PartitionsResolver))

	// Ensure a model ID is set. We use it to track the model's partitions in the catalog.
	if mdl.State.PartitionsModelId == "" {
		mdl.State.PartitionsModelId = uuid.NewString()
		err := r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return err
		}
	}

	// Resolve partition rows
	res, err := r.C.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         r.C.InstanceID,
		Resolver:           mdl.Spec.PartitionsResolver,
		ResolverProperties: mdl.Spec.PartitionsResolverProperties.AsMap(),
		Args:               map[string]any{"state": incrementalState},
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	if err != nil {
		return err
	}
	defer res.Close()

	// Consume the rows and sync them in batches
	var batch []map[string]any
	var batchStartIdx int
	for {
		// Read a row
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to read partitions resolver output: %w", err)
		}
		batch = append(batch, row)

		// Flush a batch of rows
		if len(batch) >= _modelSyncPartitionsBatchSize {
			// Sync the partitions
			err = r.syncPartitions(ctx, mdl, batchStartIdx, batch)
			if err != nil {
				return err
			}

			// Advance the row index of the first row in the batch
			batchStartIdx += len(batch)

			// Reset the batch without reallocating
			for i := range batch {
				batch[i] = nil
			}
			batch = batch[:0]
		}
	}

	// Log
	count := batchStartIdx + len(batch)
	defer r.C.Logger.Info("Resolved model partitions", zap.String("model", self.Meta.Name.Name), zap.Int("partitions", count))

	// Flush the remaining rows not handled in the loop
	return r.syncPartitions(ctx, mdl, batchStartIdx, batch)
}

// syncPartitions syncs a batch of partition rows to the catalog.
// If a partition doesn't exist, it is inserted and marked for execution.
// If a partition already exists, it will be ignored unless its watermark field has advanced, in which case it will be marked for execution.
//
// The startIdx should be the index of the first row in the batch in the full partitions dataset.
// Partition indexes only inform the order that partitions are executed in, so they don't need to be very consistent across invocations.
//
// NOTE: This implementation inserts/updates partitions one-by-one in the catalog.
// If we start using another DB than SQLite for the catalog, it may make sense to implement batched writes.
func (r *ModelReconciler) syncPartitions(ctx context.Context, mdl *runtimev1.ModelV2, startIdx int, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}

	catalog, release, err := r.C.Runtime.Catalog(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}
	defer release()

	// Build ModelPartition objects indexed by their Key
	partitions := make(map[string]drivers.ModelPartition, len(rows))
	for i, row := range rows {
		// If a watermark field is configured, we extract and remove it from the map.
		// It is necessary to remove it to ensure the key is deterministic.
		var watermark *time.Time
		if mdl.Spec.PartitionsWatermarkField != "" {
			if v, ok := row[mdl.Spec.PartitionsWatermarkField]; ok {
				t, ok := v.(time.Time)
				if !ok {
					return fmt.Errorf(`expected a timestamp for partition watermark field %q, got type %T`, mdl.Spec.PartitionsWatermarkField, v)
				}

				watermark = &t
				delete(row, mdl.Spec.PartitionsWatermarkField)
			}
		}

		// Marshal the rest of the row
		rowJSON, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("failed to marshal partition row at index %d: %w", i, err)
		}

		// JSON serialization is deterministic in Go, so we can hash it to get a key
		key, err := md5Hash(rowJSON)
		if err != nil {
			return fmt.Errorf("failed to hash partition row at index %d: %w", i, err)
		}

		partitions[key] = drivers.ModelPartition{
			Key:        key,
			DataJSON:   rowJSON,
			Index:      startIdx + i,
			Watermark:  watermark,
			ExecutedOn: nil,
			Error:      "",
			Elapsed:    0,
		}
	}

	// Find those partitions that already exist in the catalog
	keys := make([]string, 0, len(partitions))
	for key := range partitions {
		keys = append(keys, key)
	}
	existing, err := catalog.FindModelPartitionsByKeys(ctx, mdl.State.PartitionsModelId, keys)
	if err != nil {
		return fmt.Errorf("failed to find existing partitions: %w", err)
	}

	// Handle the existing partitions by skipping or updating them.
	// We remove the handled partitions from the partitions map. The ones that remain are new and should be inserted.
	for _, old := range existing {
		// Pop the matching partition from the map
		partition := partitions[old.Key]
		delete(partitions, old.Key)

		// If the watermark hasn't advanced, there's nothing to do
		if partition.Watermark == nil {
			continue
		}
		if old.Watermark != nil && !old.Watermark.Before(*partition.Watermark) {
			continue
		}

		// Update the partition (the new partition's ExecutedOn will be nil, so it will be marked for execution).
		err = catalog.UpdateModelPartition(ctx, mdl.State.PartitionsModelId, partition)
		if err != nil {
			return fmt.Errorf("failed to update existing partition: %w", err)
		}
	}

	// The remaining partitions are new and should be inserted
	for _, partition := range partitions {
		err = catalog.InsertModelPartition(ctx, mdl.State.PartitionsModelId, partition)
		if err != nil {
			return fmt.Errorf("failed to insert new partition: %w", err)
		}
	}
	return nil
}

// clearPartitions drops all partitions for a model from the catalog.
func (r *ModelReconciler) clearPartitions(ctx context.Context, mdl *runtimev1.ModelV2) error {
	if mdl.State.PartitionsModelId == "" {
		return nil
	}

	catalog, release, err := r.C.Runtime.Catalog(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}
	defer release()

	return catalog.DeleteModelPartitions(ctx, mdl.State.PartitionsModelId)
}

// executeAll executes all partitions (if any) of a model with the given execution options.
// Note that triggerReset only denotes if a reset is required. Even if it is false, the model will still be reset if it's not an incremental model.
func (r *ModelReconciler) executeAll(ctx context.Context, self *runtimev1.Resource, model *runtimev1.ModelV2, env *drivers.ModelEnv, triggerReset bool, prevResult *drivers.ModelResult) (string, *drivers.ModelResult, error) {
	// Prepare the incremental state to pass to the executor
	usePartitions := model.Spec.PartitionsResolver != ""
	incrementalRun := false
	incrementalState := map[string]any{}
	if !triggerReset && model.Spec.Incremental && prevResult != nil {
		// This is an incremental run!
		incrementalRun = true
		if model.State.IncrementalState != nil {
			incrementalState = model.State.IncrementalState.AsMap()
		}
	}
	incrementalState["incremental"] = incrementalRun // The incremental flag is hard-coded by convention

	// Build log message
	logArgs := []zap.Field{zap.String("model", self.Meta.Name.Name)}
	if incrementalRun {
		logArgs = append(logArgs, zap.String("run_type", "incremental"))
	} else {
		logArgs = append(logArgs, zap.String("run_type", "reset"))
	}
	if usePartitions {
		logArgs = append(logArgs, zap.Bool("partition", true))
	}
	if model.Spec.InputConnector == model.Spec.OutputConnector {
		logArgs = append(logArgs, zap.String("connector", model.Spec.InputConnector))
	} else {
		logArgs = append(logArgs, zap.String("input_connector", model.Spec.InputConnector), zap.String("output_connector", model.Spec.OutputConnector))
	}
	if model.Spec.StageConnector != "" {
		logArgs = append(logArgs, zap.String("stage_connector", model.Spec.StageConnector))
	}
	r.C.Logger.Info("Executing model", logArgs...)

	// Apply the timeout to the ctx
	timeout := _modelDefaultTimeout
	if model.Spec.TimeoutSeconds > 0 {
		timeout = time.Duration(model.Spec.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// On non-incremental runs, we need to clear all partition state from the catalog
	if !incrementalRun {
		err := r.clearPartitions(ctx, model)
		if err != nil {
			return "", nil, err
		}
	}

	// Get executor(s)
	executor, release, err := r.acquireExecutor(ctx, self, model, env)
	if err != nil {
		return "", nil, err
	}
	defer release()

	// For safety, double check the ctx before executing the model (there may be some code paths where it's not checked)
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}

	// If we're not partitionting execution, run the executor directly and return
	if !usePartitions {
		res, err := r.executeSingle(ctx, executor, self, model, prevResult, incrementalRun, incrementalState, nil)
		if err != nil {
			return "", nil, err
		}
		return executor.finalConnector, res, err
	}

	// At this point, we know we're running with partitions configured.

	// Discover number of concurrent partitions to process at a time
	concurrency, ok := executor.final.Concurrency(int(model.Spec.PartitionsConcurrencyLimit))
	if !ok {
		return "", nil, fmt.Errorf("invalid concurrency limit %d for model executor %q", model.Spec.PartitionsConcurrencyLimit, executor.finalConnector)
	}
	if executor.stage != nil {
		stageConcurrency, ok := executor.stage.Concurrency(int(model.Spec.PartitionsConcurrencyLimit))
		if !ok {
			return "", nil, fmt.Errorf("invalid concurrency limit %d for model stage executor %q", model.Spec.PartitionsConcurrencyLimit, executor.stageConnector)
		}
		if stageConcurrency < concurrency {
			concurrency = stageConcurrency
		}
	}
	if concurrency < 1 {
		return "", nil, fmt.Errorf("invalid concurrency limit %d for model executor %q", model.Spec.PartitionsConcurrencyLimit, executor.finalConnector)
	}

	// Prepare catalog which tracks partitions
	catalog, release, err := r.C.Runtime.Catalog(ctx, r.C.InstanceID)
	if err != nil {
		return "", nil, err
	}
	defer release()

	// First step is to resolve and sync the partitions.
	err = r.resolveAndSyncPartitions(ctx, self, model, incrementalState)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sync partitions: %w", err)
	}

	// We run the first partition without concurrency to ensure that only incremental runs are executed concurrently.
	// This enables the first partition to create the initial result (such as a table) that the other partitions incrementally build upon.
	if !incrementalRun {
		// Find the first partition
		partitions, err := catalog.FindModelPartitions(ctx, &drivers.FindModelPartitionsOptions{
			ModelID:      model.State.PartitionsModelId,
			WherePending: true,
			Limit:        1,
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to load first partition: %w", err)
		}
		if len(partitions) == 0 {
			return "", nil, fmt.Errorf("no partitions found")
		}
		partition := partitions[0]

		// Execute the first partition (with returnErr=true because for the first partition, we do not log and skip erroring partitions)
		res, ok, err := r.executePartition(ctx, catalog, executor, self, model, prevResult, incrementalRun, incrementalState, partition, true)
		if err != nil {
			return "", nil, err
		}
		if !ok {
			panic("executePartition returned false despite returnErr being set to true") // Can't happen
		}

		// Update the state so the next invocations will be incremental
		prevResult = res
		incrementalRun = true
	}

	// Repeatedly load a batch of pending partitions and execute it with a pool of worker goroutines.
	for {
		// Get a batch of pending partitions
		// Note: We do this when no workers are running because partitions are considered pending if they have not completed execution yet.
		// This reduces concurrency when processing the last handful of partitions in each batch, but with large batch sizes it's worth the simplicity for now.
		partitions, err := catalog.FindModelPartitions(ctx, &drivers.FindModelPartitionsOptions{
			ModelID:      model.State.PartitionsModelId,
			WherePending: true,
			Limit:        _modelPendingPartitionsBatchSize,
		})
		if err != nil {
			return "", nil, err
		}
		if len(partitions) == 0 {
			break
		}

		// Determine how many workers goroutines to start
		workers := concurrency
		if len(partitions) < concurrency {
			workers = len(partitions)
		}

		// Prepare the results of each worker.
		// For incremental runs, we need to pass the previous result to the executor, but for partition runs, we do not guarantee that the result is the most *recent* previous result.
		// We do guarantee that no result is discarded, and that all results are either passed as a previous result to the executor or passed into MergePartitionResults.
		// To that end, we can start all the workers off with the same initial previous result.
		results := make([]*drivers.ModelResult, workers)
		for workerID := 0; workerID < workers; workerID++ {
			results[workerID] = prevResult
		}

		// Start the worker goroutines
		grp, ctx := errgroup.WithContext(ctx)
		counter := &atomic.Int64{}
		for workerID := 0; workerID < workers; workerID++ {
			workerID := workerID
			grp.Go(func() error {
				for {
					// Atomically grab the index of a partition to process
					idx := counter.Add(1) - 1
					if idx >= int64(len(partitions)) {
						return nil
					}

					// Check the context in case the executor doesn't
					if ctx.Err() != nil {
						return ctx.Err()
					}

					// Execute the partition and capture the result in results[workerID]
					partition := partitions[idx]
					res, ok, err := r.executePartition(ctx, catalog, executor, self, model, results[workerID], incrementalRun, incrementalState, partition, false)
					if err != nil {
						return err
					}
					if ok {
						results[workerID] = res
					}
				}
			})
		}

		// Wait for all workers to finish
		err = grp.Wait()
		if err != nil {
			return "", nil, err
		}

		// Finally combine the results of each worker into the prevResult
		for _, r := range results {
			if r == nil {
				continue
			}
			if prevResult == nil {
				prevResult = r
				continue
			}

			prevResult, err = executor.finalResultManager.MergePartitionResults(prevResult, r)
			if err != nil {
				return "", nil, fmt.Errorf("failed to merge partition task results: %w", err)
			}
		}

		// If we got fewer partitions than the batch size, we've processed all pending partitions and can stop.
		if len(partitions) < _modelPendingPartitionsBatchSize {
			break
		}
	}

	// Should not happen, could also have been a panic
	if prevResult == nil {
		return "", nil, fmt.Errorf("partition execution succeeded but did not produce a non-nil result")
	}

	// We have continuously updated prevResult with new partition results, so we return it here
	return executor.finalConnector, prevResult, nil
}

// executePartition processes a drivers.ModelPartition by calling executeSingle and then updating the partition's state in the catalog.
// The returned bool will be false if execution failed, but the error was written to the partition in the catalog instead of being returned.
func (r *ModelReconciler) executePartition(ctx context.Context, catalog drivers.CatalogStore, executor *wrappedModelExecutor, self *runtimev1.Resource, mdl *runtimev1.ModelV2, prevResult *drivers.ModelResult, incrementalRun bool, incrementalState map[string]any, partition drivers.ModelPartition, returnErr bool) (*drivers.ModelResult, bool, error) {
	// Get partition data
	data := map[string]any{}
	err := json.Unmarshal(partition.DataJSON, &data)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal partition data: %w", err)
	}

	// Log
	logArgs := []zap.Field{zap.String("model", self.Meta.Name.Name), zap.String("key", partition.Key)}
	if len(partition.DataJSON) < 256 {
		logArgs = append(logArgs, zap.Any("data", data))
	}
	r.C.Logger.Debug("Executing model partition", logArgs...)
	defer func() { r.C.Logger.Info("Executed model partition", logArgs...) }()

	// Execute the partition.
	start := time.Now()
	errStr := ""
	res, err := r.executeSingle(ctx, executor, self, mdl, prevResult, incrementalRun, incrementalState, data)
	if err != nil {
		// Unless cancelled or explicitly told to return the error, we save the error in the partition and continue.
		if returnErr {
			return nil, false, err
		}
		if errors.Is(err, ctx.Err()) {
			return nil, false, err
		}
		errStr = err.Error()
		logArgs = append(logArgs, zap.Error(err))
	}

	// Mark the partition as executed
	now := time.Now()
	partition.ExecutedOn = &now
	partition.Error = errStr
	partition.Elapsed = time.Since(start)
	logArgs = append(logArgs, zap.Duration("elapsed", partition.Elapsed))

	err = catalog.UpdateModelPartition(ctx, mdl.State.PartitionsModelId, partition)
	if err != nil {
		return nil, false, fmt.Errorf("failed to update partition: %w", err)
	}
	return res, res != nil, nil
}

// executeSingle executes a single step of a model. Passing a previous result, incremental state, and/or a partition is optional.
func (r *ModelReconciler) executeSingle(ctx context.Context, executor *wrappedModelExecutor, self *runtimev1.Resource, mdl *runtimev1.ModelV2, prevResult *drivers.ModelResult, incrementalRun bool, incrementalState, partition map[string]any) (*drivers.ModelResult, error) {
	// Resolve templating in the input and output props
	inputProps, err := r.resolveTemplatedProps(ctx, self, incrementalState, partition, mdl.Spec.InputConnector, mdl.Spec.InputProperties.AsMap())
	if err != nil {
		return nil, err
	}
	outputProps, err := r.resolveTemplatedProps(ctx, self, incrementalState, partition, mdl.Spec.OutputConnector, mdl.Spec.OutputProperties.AsMap())
	if err != nil {
		return nil, err
	}

	tempDir, err := r.C.Runtime.TempDir(r.C.InstanceID)
	if err != nil {
		return nil, err
	}

	// Execute the stage step if configured
	if executor.stage != nil {
		// Also resolve templating in the stage props
		stageProps, err := r.resolveTemplatedProps(ctx, self, incrementalState, partition, mdl.Spec.StageConnector, mdl.Spec.StageProperties.AsMap())
		if err != nil {
			return nil, err
		}

		// Execute the stage step
		stageResult, err := executor.stage.Execute(ctx, &drivers.ModelExecuteOptions{
			ModelExecutorOptions: executor.stageOpts,
			InputProperties:      inputProps,
			OutputProperties:     stageProps,
			Priority:             0,
			Incremental:          mdl.Spec.Incremental,
			IncrementalRun:       incrementalRun,
			PartitionRun:         partition != nil,
			PreviousResult:       prevResult,
			TempDir:              tempDir,
		})
		if err != nil {
			return nil, err
		}

		// We change the inputProps to be the result properties of the stage step
		inputProps = stageResult.Properties

		// Drop the stage result after the final step has executed.
		// We do this using the same ctx, which means we may leak data in the staging connector in case of context cancellations.
		// This is acceptable since the staging connector is assumed to be configured for temporary data.
		defer func() {
			err := executor.stageResultManager.Delete(ctx, stageResult)
			if err != nil {
				r.C.Logger.Warn("Failed to clean up staged model output", zap.String("model", self.Meta.Name.Name), zap.Error(err))
			}
		}()
	}

	// Execute the final step
	finalResult, err := executor.final.Execute(ctx, &drivers.ModelExecuteOptions{
		ModelExecutorOptions: executor.finalOpts,
		InputProperties:      inputProps,
		OutputProperties:     outputProps,
		Priority:             0,
		Incremental:          mdl.Spec.Incremental,
		IncrementalRun:       incrementalRun,
		PartitionRun:         partition != nil,
		PreviousResult:       prevResult,
		TempDir:              tempDir,
	})
	if err != nil {
		return nil, err
	}
	return finalResult, nil
}

// wrappedModelExecutor is a ModelExecutor wraps one or two ModelExecutors. It is used to execute a model with a staging connector.
// If the model does not require a staging connector, the wrappedModelExecutor will only wrap the final executor.
type wrappedModelExecutor struct {
	finalConnector     string
	final              drivers.ModelExecutor
	finalOpts          *drivers.ModelExecutorOptions
	finalResultManager drivers.ModelManager
	stageConnector     string
	stage              drivers.ModelExecutor
	stageOpts          *drivers.ModelExecutorOptions
	stageResultManager drivers.ModelManager
}

// acquireExecutor acquires the executor(s) necessary for executing the given model.
// If the model has a stage connector, it will acquire and combine two executors: one from the input to the stage connector, and another from the stage to the output connector.
func (r *ModelReconciler) acquireExecutor(ctx context.Context, self *runtimev1.Resource, mdl *runtimev1.ModelV2, env *drivers.ModelEnv) (*wrappedModelExecutor, func(), error) {
	// Handle the simple case where there is no stage connector
	if mdl.Spec.StageConnector == "" {
		opts := &drivers.ModelExecutorOptions{
			Env:                         env,
			ModelName:                   self.Meta.Name.Name,
			InputHandle:                 nil,
			InputConnector:              mdl.Spec.InputConnector,
			PreliminaryInputProperties:  mdl.Spec.InputProperties.AsMap(),
			OutputHandle:                nil,
			OutputConnector:             mdl.Spec.OutputConnector,
			PreliminaryOutputProperties: mdl.Spec.OutputProperties.AsMap(),
		}

		connector, executor, release, err := r.acquireExecutorInner(ctx, opts)
		if err != nil {
			return nil, nil, err
		}

		finalResultManager, ok := opts.OutputHandle.AsModelManager(r.C.InstanceID)
		if !ok {
			release()
			return nil, nil, fmt.Errorf("output connector %q is not capable of managing model results", mdl.Spec.OutputConnector)
		}

		return &wrappedModelExecutor{
			finalConnector:     connector,
			final:              executor,
			finalOpts:          opts,
			finalResultManager: finalResultManager,
		}, release, nil
	}

	// Acquire the stage executor
	stageOpts := &drivers.ModelExecutorOptions{
		Env:                         env,
		ModelName:                   self.Meta.Name.Name,
		InputHandle:                 nil,
		InputConnector:              mdl.Spec.InputConnector,
		PreliminaryInputProperties:  mdl.Spec.InputProperties.AsMap(),
		OutputHandle:                nil,
		OutputConnector:             mdl.Spec.StageConnector,
		PreliminaryOutputProperties: mdl.Spec.StageProperties.AsMap(),
	}
	stageConnector, stage, stageRelease, err := r.acquireExecutorInner(ctx, stageOpts)
	if err != nil {
		return nil, nil, err
	}

	// Acquire the stage result manager
	stageResultManager, ok := stageOpts.OutputHandle.AsModelManager(r.C.InstanceID)
	if !ok {
		stageRelease()
		return nil, nil, fmt.Errorf("staging connector %q is not capable of managing model results", mdl.Spec.StageConnector)
	}

	// Acquire the final executor
	finalOpts := &drivers.ModelExecutorOptions{
		Env:                         env,
		ModelName:                   self.Meta.Name.Name,
		InputHandle:                 nil,
		InputConnector:              mdl.Spec.StageConnector,
		PreliminaryInputProperties:  mdl.Spec.StageProperties.AsMap(),
		OutputHandle:                nil,
		OutputConnector:             mdl.Spec.OutputConnector,
		PreliminaryOutputProperties: mdl.Spec.OutputProperties.AsMap(),
	}
	finalConnector, final, finalRelease, err := r.acquireExecutorInner(ctx, finalOpts)
	if err != nil {
		stageRelease()
		return nil, nil, err
	}

	// Acquire the final result manager
	finalResultManager, ok := finalOpts.OutputHandle.AsModelManager(r.C.InstanceID)
	if !ok {
		finalRelease()
		return nil, nil, fmt.Errorf("output connector %q is not capable of managing model results", mdl.Spec.OutputConnector)
	}

	// Wrap the executors
	wrapped := &wrappedModelExecutor{
		finalConnector:     finalConnector,
		final:              final,
		finalOpts:          finalOpts,
		stageConnector:     stageConnector,
		stage:              stage,
		stageOpts:          stageOpts,
		stageResultManager: stageResultManager,
		finalResultManager: finalResultManager,
	}
	release := func() {
		stageRelease()
		finalRelease()
	}
	return wrapped, release, nil
}

// acquireExecutorInner acquires a ModelExecutor by directly calling AsModelExecutor on the input and output connectors.
// It handles acquiring and setting opts.InputHandle and opts.OutputHandle.
func (r *ModelReconciler) acquireExecutorInner(ctx context.Context, opts *drivers.ModelExecutorOptions) (string, drivers.ModelExecutor, func(), error) {
	ic, ir, err := r.C.AcquireConn(ctx, opts.InputConnector)
	if err != nil {
		return "", nil, nil, err
	}

	if opts.InputConnector == opts.OutputConnector {
		opts.InputHandle = ic
		opts.OutputHandle = ic

		e, ok := ic.AsModelExecutor(r.C.InstanceID, opts)
		if !ok {
			return "", nil, nil, fmt.Errorf("connector %q is not capable of executing models", opts.InputConnector)
		}

		return opts.InputConnector, e, ir, nil
	}

	oc, or, err := r.C.AcquireConn(ctx, opts.OutputConnector)
	if err != nil {
		ir()
		return "", nil, nil, err
	}

	opts.InputHandle = ic
	opts.OutputHandle = oc

	executorName := opts.InputConnector
	e, ok := ic.AsModelExecutor(r.C.InstanceID, opts)
	if !ok {
		executorName = opts.OutputConnector
		e, ok = oc.AsModelExecutor(r.C.InstanceID, opts)
		if !ok {
			ir()
			or()
			return "", nil, nil, fmt.Errorf("cannot execute model: input connector %q and output connector %q are not compatible", opts.InputConnector, opts.OutputConnector)
		}
	}

	release := func() {
		ir()
		or()
	}

	return executorName, e, release, nil
}

// newModelEnv makes a ModelEnv configured using the current instance.
func (r *ModelReconciler) newModelEnv(ctx context.Context) (*drivers.ModelEnv, error) {
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to access instance config: %w", err)
	}

	repo, release, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to access repo: %w", err)
	}
	defer release()

	repoRoot, err := repo.Root(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo root: %w", err)
	}

	return &drivers.ModelEnv{
		AllowHostAccess:    r.C.Runtime.AllowHostAccess(),
		RepoRoot:           repoRoot,
		StageChanges:       cfg.StageChanges,
		DefaultMaterialize: cfg.ModelDefaultMaterialize,
		AcquireConnector:   r.C.AcquireConn,
	}, nil
}

// resolveTemplatedProps resolves template tags in strings nested in the provided props.
// Passing a connector is optional. If a connector is provided, it will be used to inform how values are escaped.
func (r *ModelReconciler) resolveTemplatedProps(ctx context.Context, self *runtimev1.Resource, incrementalState, partition map[string]any, connector string, props map[string]any) (map[string]any, error) {
	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return nil, err
	}

	// If we know the prop's connector AND it's an OLAP, we use its dialect to escape refs
	var dialect drivers.Dialect
	if connector != "" {
		olap, release, err := r.C.AcquireOLAP(ctx, connector)
		if err == nil {
			dialect = olap.Dialect()
			release()
		}
	}

	var extraProps map[string]any
	if partition != nil {
		extraProps = map[string]any{
			"partition": partition,
			"split":     partition, // Deprecated: use "partition" instead
		}
	}

	td := compilerv1.TemplateData{
		Environment: inst.Environment,
		User:        map[string]any{},
		Variables:   inst.ResolveVariables(false),
		State:       incrementalState,
		ExtraProps:  extraProps,
		Self: compilerv1.TemplateResource{
			Meta:  self.Meta,
			Spec:  self.GetModel().Spec,
			State: self.GetModel().State,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			if dialect == drivers.DialectUnspecified {
				return ref.Name, nil
			}
			return dialect.EscapeIdentifier(ref.Name), nil
		},
	}

	val, err := compilerv1.ResolveTemplateRecursively(props, td)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template: %w", err)
	}
	return val.(map[string]any), nil
}

// analyzeTemplatedVariables analyzes strings nested in the provided props for template tags that reference instance variables.
// It returns a map of variable names referenced in the props mapped to their current value (if known).
func (r *ModelReconciler) analyzeTemplatedVariables(ctx context.Context, props map[string]any) (map[string]string, error) {
	res := make(map[string]string)
	err := compilerv1.AnalyzeTemplateRecursively(props, res)
	if err != nil {
		return nil, err
	}

	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables(false)

	for k := range res {
		// Project variables are referenced with .env.name (current) or .vars.name (deprecated).
		// Other templated variable names are not project variable references.
		if k2 := strings.TrimPrefix(k, "env."); k != k2 {
			res[k] = vars[k2]
		} else if k2 := strings.TrimPrefix(k, "vars."); k != k2 {
			res[k] = vars[k2]
		}
	}

	return res, nil
}

// hashWriteMapOrdered writes the keys and values of a map to the writer in a deterministic order.
func hashWriteMapOrdered(w io.Writer, m map[string]string) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, err := w.Write([]byte(k))
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(m[k]))
		if err != nil {
			return err
		}
	}

	return nil
}

// md5Hash returns a hex-encoded SHA-256 hash of the provided byte slice.
func md5Hash(val []byte) (string, error) {
	hash := md5.New()
	_, err := hash.Write(val)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
