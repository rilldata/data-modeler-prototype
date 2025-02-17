package river

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type InitOrgBillingArgs struct {
	OrgID string
}

func (InitOrgBillingArgs) Kind() string { return "init_org_billing" }

type InitOrgBillingWorker struct {
	river.WorkerDefaults[InitOrgBillingArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker initializes the billing for an organization
func (w *InitOrgBillingWorker) Work(ctx context.Context, job *river.Job[InitOrgBillingArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization %s: %w", job.Args.OrgID, err)
	}

	if job.Attempt > 1 {
		// rare case but if its retried, we should repair the billing as it might be in some inconsistent state
		_, _, err = w.admin.RepairOrganizationBilling(ctx, org, false)
		if err != nil {
			return fmt.Errorf("failed to repair billing for organization %s: %w", org.Name, err)
		}
		return nil
	}

	_, err = w.admin.InitOrganizationBilling(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to init billing for organization %s: %w", org.Name, err)
	}
	return nil
}

type RepairOrgBillingArgs struct {
	OrgID string
}

func (RepairOrgBillingArgs) Kind() string { return "repair_org_billing" }

type RepairOrgBillingWorker struct {
	river.WorkerDefaults[RepairOrgBillingArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker repairs the billing for an organization
func (w *RepairOrgBillingWorker) Work(ctx context.Context, job *river.Job[RepairOrgBillingArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization %s: %w", job.Args.OrgID, err)
	}

	_, _, err = w.admin.RepairOrganizationBilling(ctx, org, true)
	if err != nil {
		return fmt.Errorf("failed to repair billing for organization %s: %w", org.Name, err)
	}
	return nil
}

type StartTrialArgs struct {
	OrgID string
}

func (StartTrialArgs) Kind() string { return "start_trial" }

type StartTrialWorker struct {
	river.WorkerDefaults[StartTrialArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker starts the trial for an organization
func (w *StartTrialWorker) Work(ctx context.Context, job *river.Job[StartTrialArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return err
	}

	trialOrg, sub, err := w.admin.StartTrial(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to start trial for organization %s: %w", org.Name, err)
	}

	// send trial started email
	err = w.admin.Email.SendTrialStarted(&email.TrialStarted{
		ToEmail:      trialOrg.BillingEmail,
		ToName:       trialOrg.Name,
		OrgName:      trialOrg.Name,
		FrontendURL:  w.admin.URLs.Frontend(),
		TrialEndDate: sub.TrialEndDate,
	})
	if err != nil {
		w.logger.Error("failed to send trial started email", zap.String("org_name", trialOrg.Name), zap.String("org_id", trialOrg.ID), zap.String("billing_email", trialOrg.BillingEmail), zap.Error(err))
	}

	return nil
}

type DeleteOrgArgs struct {
	OrgID string
}

func (DeleteOrgArgs) Kind() string { return "delete_org" }

type DeleteOrgWorker struct {
	river.WorkerDefaults[DeleteOrgArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker handles the deletion of an organization and cancels all subscriptions related to it
func (w *DeleteOrgWorker) Work(ctx context.Context, job *river.Job[DeleteOrgArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return err
	}

	// cancel all subscriptions for the customer immediately but keep the customer in billing and payment system for issued invoices
	if org.BillingCustomerID != "" {
		_, err = w.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate)
		if err != nil {
			w.logger.Error("failed to cancel subscriptions for customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
		}
	}

	// delete org, billing issues will be cascade deleted
	err = w.admin.DB.DeleteOrganization(ctx, org.Name)
	if err != nil {
		return err
	}

	w.logger.Warn("organization deleted", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	return nil
}
