package worker

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/pkg/riverworker"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const jobTimeout = 60 * time.Minute

var (
	tracer              = otel.Tracer("github.com/rilldata/rill/admin/worker")
	meter               = otel.Meter("github.com/rilldata/rill/admin/worker")
	jobLatencyHistogram = observability.Must(meter.Int64Histogram("job_latency", metric.WithUnit("ms")))
)

type Worker struct {
	logger        *zap.Logger
	admin         *admin.Service
	riverMigrator *rivermigrate.Migrator[*sql.Tx]
	riverDBPool   *sql.DB
	riverClient   *river.Client[*sql.Tx]
}

func New(logger *zap.Logger, adm *admin.Service, driver riverdriver.Driver[*sql.Tx], riverDBPool *sql.DB) *Worker {
	client, err := river.NewClient[*sql.Tx](driver, &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers:      riverworker.Workers,
		JobTimeout:   10 * time.Minute,
		MaxAttempts:  3,
		ErrorHandler: &riverworker.ErrorHandler{Logger: logger},
		// TODO set logger as well but it requires slog instead of zap
	})
	if err != nil {
		panic(err)
	}

	migrator := rivermigrate.New[*sql.Tx](driver, nil)

	return &Worker{
		logger:        logger,
		admin:         adm,
		riverDBPool:   riverDBPool,
		riverMigrator: migrator,
		riverClient:   client,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		// Migrate the database
		tx, err := w.riverDBPool.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		res, err := w.riverMigrator.MigrateTx(ctx, tx, rivermigrate.DirectionUp, nil)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
		for _, version := range res.Versions {
			w.logger.Info("Migrated river database", zap.String("direction", string(res.Direction)), zap.Int("version", version.Version))
		}

		if err := w.riverClient.Start(ctx); err != nil {
			panic(err)
		}
		<-ctx.Done()
		_ = w.riverClient.Stop(ctx) // ignore error
		return nil
	})

	group.Go(func() error {
		return w.schedule(ctx, "check_provisioner_capacity", w.checkProvisionerCapacity, 15*time.Minute)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_tokens", w.deleteExpiredAuthTokens, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_device_auth_codes", w.deleteExpiredDeviceAuthCodes, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_auth_codes", w.deleteExpiredAuthCodes, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_virtual_files", w.deleteExpiredVirtualFiles, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "hibernate_expired_deployments", w.hibernateExpiredDeployments, 15*time.Minute)
	})
	group.Go(func() error {
		return w.schedule(ctx, "validate_deployments", w.validateDeployments, 6*time.Hour)
	})
	group.Go(func() error {
		return w.scheduleCron(ctx, "run_autoscaler", w.runAutoscaler, w.admin.AutoscalerCron)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_unused_assets", w.deleteUnusedAssets, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "deployments_health_check", w.deploymentsHealthCheck, 10*time.Minute)
	})

	if w.admin.Biller.GetReportingWorkerCron() != "" {
		group.Go(func() error {
			return w.scheduleCron(ctx, "run_billing_reporter", w.reportUsage, w.admin.Biller.GetReportingWorkerCron())
		})
	}

	if w.admin.Biller.Name() != "noop" {
		group.Go(func() error {
			return w.schedule(ctx, "run_billing_repair", w.repairOrgBilling, 10*time.Minute)
		})

		group.Go(func() error {
			// run every midnight
			return w.scheduleCron(ctx, "run_trial_end_check", w.trialEndCheck, "0 0 * * *")
		})
	}
	// NOTE: Add new scheduled jobs here

	w.logger.Info("worker started")
	defer w.logger.Info("worker stopped")
	return group.Wait()
}

func (w *Worker) RunJob(ctx context.Context, name string) error {
	switch name {
	case "check_provisioner_capacity":
		return w.runJob(ctx, name, w.checkProvisionerCapacity)
	case "reset_all_deployments":
		return w.runJob(ctx, name, w.resetAllDeployments)
	case "validate_deployments":
		return w.runJob(ctx, name, w.validateDeployments)
	// NOTE: Add new ad-hoc jobs here
	default:
		return fmt.Errorf("unknown job: %s", name)
	}
}

func (w *Worker) schedule(ctx context.Context, name string, fn func(context.Context) error, every time.Duration) error {
	for {
		err := w.runJob(ctx, name, fn)
		if err != nil {
			w.logger.Error("Failed to run the job", zap.String("job_name", name), zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(every):
		}
	}
}

func (w *Worker) scheduleCron(ctx context.Context, name string, fn func(context.Context) error, cronExpr string) error {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return err
	}

	for {
		nextRun := schedule.Next(time.Now())
		waitDuration := time.Until(nextRun)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(waitDuration):
			err := w.runJob(ctx, name, fn)
			if err != nil {
				w.logger.Error("Failed to run the cronjob", zap.String("cronjob_name", name), zap.Error(err))
			}
		}
	}
}

func (w *Worker) runJob(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	ctx, span := tracer.Start(ctx, fmt.Sprintf("runJob %s", name), trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	start := time.Now()
	w.logger.Info("job started", zap.String("name", name), observability.ZapCtx(ctx))
	err := fn(ctx)
	jobLatencyHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attribute.String("name", name), attribute.Bool("failed", err != nil)))
	if err != nil {
		w.logger.Error("job failed", zap.String("name", name), zap.Error(err), zap.Duration("duration", time.Since(start)), observability.ZapCtx(ctx))
		return err
	}
	w.logger.Info("job completed", zap.String("name", name), zap.Duration("duration", time.Since(start)), observability.ZapCtx(ctx))
	return nil
}

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {
		panic(err)
	}
}

// StartPingServer starts a http server that returns 200 OK on /ping
func StartPingServer(ctx context.Context, port int) error {
	httpMux := http.NewServeMux()
	httpMux.Handle("/ping", &pingHandler{})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: httpMux,
	}

	return graceful.ServeHTTP(ctx, srv, graceful.ServeOptions{
		Port: port,
	})
}
