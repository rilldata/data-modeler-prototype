package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _defaultIngestTimeout = 60 * time.Minute

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindSource, newSourceReconciler)
}

type SourceReconciler struct {
	C *runtime.Controller
}

func newSourceReconciler(c *runtime.Controller) runtime.Reconciler {
	return &SourceReconciler{C: c}
}

func (r *SourceReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *SourceReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	src := self.GetSource()

	// The table name to ingest into is derived from the resource name.
	// We only set src.State.Table after ingestion is complete.
	// The value of tableName and src.State.Table will differ until initial successful ingestion and when renamed.
	tableName := self.Meta.Name.Name

	// Handle deletion
	if self.Meta.Deleted {
		olapDropTableIfExists(ctx, r.C, src.State.Connector, src.State.Table, false)
		olapDropTableIfExists(ctx, r.C, src.State.Connector, r.stagingTableName(tableName), false)
		return runtime.ReconcileResult{}
	}

	// Handle renames
	if self.Meta.RenamedFrom != nil {
		// Check if the table exists (it should, but might somehow have been corrupted)
		t, ok := olapTableInfo(ctx, r.C, src.State.Connector, src.State.Table)
		if ok && !t.View { // Checking View only out of caution (would indicate very corrupted DB)
			// Clear any existing table with the new name
			if t2, ok := olapTableInfo(ctx, r.C, src.State.Connector, tableName); ok {
				olapDropTableIfExists(ctx, r.C, src.State.Connector, tableName, t2.View)
			}

			// Rename and update state
			err = olapRenameTable(ctx, r.C, src.State.Connector, src.State.Table, tableName, false)
			if err != nil {
				return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename table: %w", err)}
			}
			src.State.Table = tableName
			err = r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
		}
		// Note: Not exiting early. It might need to be (re-)ingested, and we need to set the correct retrigger time based on the refresh schedule.
	}

	// TODO: Exit if refs have errors

	// Use a hash of ingestion-related fields from the spec to determine if we need to re-ingest
	hash, err := r.ingestionSpecHash(src.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}

	// Compute next time to refresh based on the RefreshSchedule (if any)
	var refreshOn time.Time
	if src.State.RefreshedOn != nil {
		refreshOn, err = nextRefreshTime(src.State.RefreshedOn.AsTime(), src.Spec.RefreshSchedule)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Check if the table still exists (might have been corrupted/lost somehow)
	tableExists := false
	if src.State.Table != "" {
		t, ok := olapTableInfo(ctx, r.C, src.State.Connector, src.State.Table)
		tableExists = ok && !t.View
	}

	// Decide if we should trigger a refresh
	trigger := src.Spec.Trigger                                             // If Trigger is set
	trigger = trigger || src.State.Table == ""                              // If table is missing
	trigger = trigger || src.State.RefreshedOn == nil                       // If never refreshed
	trigger = trigger || src.State.SpecHash != hash                         // If the spec has changed
	trigger = trigger || !tableExists                                       // If the table has disappeared
	trigger = trigger || !refreshOn.IsZero() && time.Now().After(refreshOn) // If the schedule says it's time

	// Exit early if no trigger
	if !trigger {
		return runtime.ReconcileResult{Retrigger: refreshOn}
	}

	// If the SinkConnector was changed, drop data in the old connector
	if src.State.Table != "" && src.State.Connector != src.Spec.SinkConnector {
		olapDropTableIfExists(ctx, r.C, src.State.Connector, src.State.Table, false)
		olapDropTableIfExists(ctx, r.C, src.State.Connector, r.stagingTableName(src.State.Table), false)
	}

	// Prepare for ingestion
	stagingTableName := tableName
	connector := src.Spec.SinkConnector
	if src.Spec.StageChanges {
		stagingTableName = r.stagingTableName(tableName)
	}

	// Should never happen, but if somehow the staging table was corrupted into a view, drop it
	if t, ok := olapTableInfo(ctx, r.C, connector, stagingTableName); ok && t.View {
		olapDropTableIfExists(ctx, r.C, connector, stagingTableName, t.View)
	}

	// Execute ingestion
	ingestErr := r.ingestSource(ctx, src.Spec, stagingTableName)
	if ingestErr != nil {
		ingestErr = fmt.Errorf("failed to ingest source: %w", ingestErr)
	}
	if ingestErr == nil && src.Spec.StageChanges {
		// Drop the main table name
		if t, ok := olapTableInfo(ctx, r.C, connector, tableName); ok {
			olapDropTableIfExists(ctx, r.C, connector, tableName, t.View)
		}
		// Rename staging table to main table
		err = olapRenameTable(ctx, r.C, connector, stagingTableName, tableName, false)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename staging table: %w", err)}
		}
	}

	// How we handle ingestErr depends on several things:
	// If ctx was cancelled, we cleanup and exit
	// If StageChanges is true, we retain the existing table, but still return the error.
	// If StageChanges is false, we clear the existing table and return the error.

	// ctx will only be cancelled in cases where the Controller guarantees a new call to Reconcile.
	// We just clean up temp tables and state, then return.
	cleanupCtx := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		cleanupCtx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
	}

	// Update state
	update := false
	if ingestErr == nil {
		// Successful ingestion
		update = true
		src.State.Connector = connector
		src.State.Table = tableName
		src.State.SpecHash = hash
		src.State.RefreshedOn = timestamppb.Now()
	} else if src.Spec.StageChanges {
		// Failed ingestion to staging table
		olapDropTableIfExists(cleanupCtx, r.C, connector, stagingTableName, false)
	} else {
		// Failed ingestion to main table
		update = true
		olapDropTableIfExists(cleanupCtx, r.C, connector, tableName, false)
		src.State.Connector = ""
		src.State.Table = ""
		src.State.SpecHash = ""
		src.State.RefreshedOn = nil
	}
	if update {
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// See earlier note – essential cleanup is done, we can return now.
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: ingestErr}
	}

	// Reset spec.Trigger
	if src.Spec.Trigger {
		src.Spec.Trigger = false
		err = r.C.UpdateSpec(ctx, self.Meta.Name, self.Meta.Refs, self.Meta.Owner, self.Meta.FilePaths, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Compute next refresh time
	refreshOn, err = nextRefreshTime(time.Now(), src.Spec.RefreshSchedule)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: ingestErr, Retrigger: refreshOn}
}

// ingestionSpecHash computes a hash of only those source spec properties that impact ingestion.
func (r *SourceReconciler) ingestionSpecHash(spec *runtimev1.SourceSpec) (string, error) {
	hash := md5.New()

	_, err := hash.Write([]byte(spec.SourceConnector))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.SinkConnector))
	if err != nil {
		return "", err
	}

	err = pbutil.WriteHash(structpb.NewStructValue(spec.Properties), hash)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func (r *SourceReconciler) stagingTableName(table string) string {
	return "__rill_tmp_src_" + table
}

// ingestSource ingests the source into a table with tableName.
// It does NOT drop the table if ingestion fails after the table has been created.
// It will return an error if the sink connector is not an OLAP.
func (r *SourceReconciler) ingestSource(ctx context.Context, src *runtimev1.SourceSpec, tableName string) error {
	// Get connections and transporter
	srcConn, release1, err := r.C.AcquireConn(ctx, src.SourceConnector)
	if err != nil {
		return err
	}
	defer release1()
	sinkConn, release2, err := r.C.AcquireConn(ctx, src.SinkConnector)
	if err != nil {
		return err
	}
	defer release2()
	t, ok := sinkConn.AsTransporter(srcConn, sinkConn)
	if !ok {
		t, ok = srcConn.AsTransporter(srcConn, sinkConn)
		if !ok {
			return fmt.Errorf("cannot transfer data between connectors %q and %q", src.SourceConnector, src.SinkConnector)
		}
	}

	// Get source and sink configs
	srcConfig, err := driversSource(srcConn, src.Properties)
	if err != nil {
		return err
	}
	sinkConfig, err := driversSink(sinkConn, tableName)
	if err != nil {
		return err
	}

	// Set timeout on ctx
	timeout := _defaultIngestTimeout
	if src.TimeoutSeconds > 0 {
		timeout = time.Duration(src.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Enforce storage limits
	// TODO: This code is pretty ugly. We should push storage limit tracking into the underlying driver and transporter.
	var ingestionLimit *int64
	var limitExceeded bool
	if olap, ok := sinkConn.AsOLAP(r.C.InstanceID); ok {
		// Get storage limit
		inst, err := r.C.Runtime.FindInstance(ctx, r.C.InstanceID)
		if err != nil {
			return err
		}
		storageLimit := inst.IngestionLimitBytes

		// Enforce storage limit if it's set
		if storageLimit > 0 {
			// Get ingestion limit (storage limit minus current size)
			bytes, ok := olap.EstimateSize()
			if ok {
				n := storageLimit - bytes
				if n <= 0 {
					return drivers.ErrIngestionLimitExceeded
				}
				ingestionLimit = &n

				// Start background goroutine to check size is not exceeded during ingestion
				go func() {
					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							if size, ok := olap.EstimateSize(); ok && size > storageLimit {
								limitExceeded = true
								cancel()
							}
						}
					}
				}()
			}
		}
	}

	// Execute the data transfer
	opts := drivers.NewTransferOpts()
	if ingestionLimit != nil {
		opts.LimitInBytes = *ingestionLimit
	}
	err = t.Transfer(ctx, srcConfig, sinkConfig, opts, drivers.NoOpProgress{})
	if limitExceeded {
		return drivers.ErrIngestionLimitExceeded
	}
	return err
}

func driversSource(conn drivers.Handle, propsPB *structpb.Struct) (drivers.Source, error) {
	props := propsPB.AsMap()
	switch conn.Driver() {
	case "s3":
		return &drivers.BucketSource{
			// ExtractPolicy: src.Policy, // TODO: Add
			Properties: props,
		}, nil
	case "gcs":
		return &drivers.BucketSource{
			// ExtractPolicy: src.Policy, // TODO: Add
			Properties: props,
		}, nil
	case "https":
		return &drivers.FileSource{
			Properties: props,
		}, nil
	case "local_file":
		return &drivers.FileSource{
			Properties: props,
		}, nil
	case "motherduck":
		query, ok := props["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"motherduck\"")
		}
		var db string
		if val, ok := props["db"].(string); ok {
			db = val
		}

		return &drivers.DatabaseSource{
			SQL:      query,
			Database: db,
		}, nil
	case "duckdb":
		query, ok := props["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"duckdb\"")
		}
		return &drivers.DatabaseSource{
			SQL: query,
		}, nil
	case "bigquery":
		query, ok := props["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"bigquery\"")
		}
		return &drivers.DatabaseSource{
			SQL:   query,
			Props: props,
		}, nil
	default:
		return nil, fmt.Errorf("source connector %q not supported", conn.Driver())
	}
}

func driversSink(conn drivers.Handle, tableName string) (drivers.Sink, error) {
	switch conn.Driver() {
	case "duckdb":
		return &drivers.DatabaseSink{
			Table: tableName,
		}, nil
	default:
		return nil, fmt.Errorf("sink connector %q not supported", conn.Driver())
	}
}
