package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type TrialEndingSoonArgs struct{}

func (TrialEndingSoonArgs) Kind() string { return "trial_ending_soon" }

type TrialEndingSoonWorker struct {
	river.WorkerDefaults[TrialEndingSoonArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *TrialEndingSoonWorker) Work(ctx context.Context, job *river.Job[TrialEndingSoonArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.trialEndingSoon)
}

func (w *TrialEndingSoonWorker) trialEndingSoon(ctx context.Context) error {
	onTrialOrgs, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeOnTrial, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	for _, o := range onTrialOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataOnTrial)
		if time.Now().UTC().Before(m.EndDate.AddDate(0, 0, -7)) {
			// trial end date is more than 7 days away, move to next org
			continue
		}

		// trial period ending soon, log warn and send email
		org, err := w.admin.DB.FindOrganization(ctx, o.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		// remaining days in the trial period
		daysRemaining := int(m.EndDate.Sub(time.Now().UTC()).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		// number of projects for the org
		projects, err := w.admin.DB.CountProjectsForOrganization(ctx, org.ID)
		if err != nil {
			return fmt.Errorf("failed to count projects for org %q: %w", org.Name, err)
		}

		w.logger.Warn("trial ending soon",
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name),
			zap.Time("trial_end_date", m.EndDate),
			zap.String("user_email", org.BillingEmail),
			zap.Int("count_of_projects", projects),
			zap.Int("count_of_days_remaining", daysRemaining),
		)

		err = w.admin.Email.SendTrialEndingSoon(&email.TrialEndingSoon{
			ToEmail:      org.BillingEmail,
			ToName:       org.Name,
			OrgName:      org.Name,
			UpgradeURL:   w.admin.URLs.Billing(org.Name, true),
			TrialEndDate: m.EndDate,
		})
		if err != nil {
			return fmt.Errorf("failed to send trial ending soon email for org %q: %w", org.Name, err)
		}

		// mark the billing issue as processed
		err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("failed to update billing issue as processed: %w", err)
		}
	}

	return nil
}

type TrialEndCheckArgs struct{}

func (TrialEndCheckArgs) Kind() string { return "trial_end_check" }

type TrialEndCheckWorker struct {
	river.WorkerDefaults[TrialEndCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *TrialEndCheckWorker) Work(ctx context.Context, job *river.Job[TrialEndCheckArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.trialEndCheck)
}

func (w *TrialEndCheckWorker) trialEndCheck(ctx context.Context) error {
	onTrialOrgs, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeOnTrial, true)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	for _, o := range onTrialOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataOnTrial)
		if time.Now().UTC().Before(m.EndDate) {
			// trial end date is not finished yet, move to next org
			continue
		}

		// trial period has ended, log warn and send email
		org, err := w.admin.DB.FindOrganization(ctx, o.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
		if err != nil {
			return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
		}
		if sub.ID != m.SubID || sub.Plan.ID != m.PlanID {
			w.logger.Warn("trial period has ended, but org has different active subscription, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("sub_id", sub.ID), zap.String("sub_plan_id", sub.Plan.ID), zap.String("expected_sub_id", m.SubID), zap.String("expected_sub_plan_id", m.PlanID))
			continue
		}

		w.logger.Warn("trial period has ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

		cctx, tx, err := w.admin.DB.NewTx(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		_, err = w.admin.DB.UpsertBillingIssue(cctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeTrialEnded,
			Metadata: &database.BillingIssueMetadataTrialEnded{
				SubID:              m.SubID,
				PlanID:             m.PlanID,
				EndDate:            m.EndDate,
				GracePeriodEndDate: m.GracePeriodEndDate,
			},
			EventTime: m.EndDate,
		})
		if err != nil {
			prevErr := err
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to add billing error: %w", prevErr)
		}

		// delete the on-trial billing issue
		err = w.admin.DB.DeleteBillingIssue(cctx, o.ID)
		if err != nil {
			prevErr := err
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to delete billing issue: %w", prevErr)
		}

		// send email
		err = w.admin.Email.SendTrialEnded(&email.TrialEnded{
			ToEmail:            org.BillingEmail,
			ToName:             org.Name,
			OrgName:            org.Name,
			UpgradeURL:         w.admin.URLs.Billing(org.Name, true),
			GracePeriodEndDate: m.GracePeriodEndDate,
		})
		if err != nil {
			prevErr := err
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to send trial period ended email for org %q: %w", org.Name, prevErr)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

type TrialGracePeriodCheckArgs struct{}

func (TrialGracePeriodCheckArgs) Kind() string { return "trial_grace_period_check" }

type TrialGracePeriodCheckWorker struct {
	river.WorkerDefaults[TrialGracePeriodCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *TrialGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[TrialGracePeriodCheckArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.trialGracePeriodCheck)
}

func (w *TrialGracePeriodCheckWorker) trialGracePeriodCheck(ctx context.Context) error {
	trailEndedOrgs, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeTrialEnded, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	for _, o := range trailEndedOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataTrialEnded)
		if time.Now().UTC().Before(m.GracePeriodEndDate) {
			// grace period end date is not finished yet, move to next org
			continue
		}

		org, err := w.admin.DB.FindOrganization(ctx, o.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		// get active subscription for the org
		sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
		if err != nil {
			if !errors.Is(err, billing.ErrNotFound) {
				return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
			}
		}
		if sub == nil {
			// might happen if previous job failed in middle
			w.logger.Warn("trial grace period has ended, but org has no active subscription, please check", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		} else if sub.ID != m.SubID || sub.Plan.ID != m.PlanID {
			w.logger.Warn("trial grace period has ended, but org has different active subscription, doing nothing, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("sub_id", sub.ID), zap.String("sub_plan_id", sub.Plan.ID), zap.String("expected_sub_id", m.SubID), zap.String("expected_sub_plan_id", m.PlanID))
			continue
		}

		// cancel the subscription
		_, err = w.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate)
		if err != nil {
			return fmt.Errorf("failed to cancel subscription for org %q: %w", org.Name, err)
		}

		// trial grace period ended, update quotas to 0 and hibernate all projects
		_, err = w.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			DisplayName:                         org.DisplayName,
			Description:                         org.Description,
			LogoAssetID:                         org.LogoAssetID,
			FaviconAssetID:                      org.FaviconAssetID,
			CustomDomain:                        org.CustomDomain,
			QuotaProjects:                       0,
			QuotaDeployments:                    0,
			QuotaSlotsTotal:                     0,
			QuotaSlotsPerDeployment:             0,
			QuotaOutstandingInvites:             0,
			QuotaStorageLimitBytesPerDeployment: 0,
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
			BillingEmail:                        org.BillingEmail,
			BillingPlanName:                     org.BillingPlanName,
			BillingPlanDisplayName:              org.BillingPlanDisplayName,
			CreatedByUserID:                     org.CreatedByUserID,
		})
		if err != nil {
			return err
		}

		limit := 10
		afterProjectName := ""
		for {
			projs, err := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, afterProjectName, limit)
			if err != nil {
				return err
			}

			for _, proj := range projs {
				_, err = w.admin.HibernateProject(ctx, proj)
				if err != nil {
					return fmt.Errorf("failed to hibernate project %q: %w", proj.Name, err)
				}
				afterProjectName = proj.Name
			}

			if len(projs) < limit {
				break
			}
		}
		w.logger.Warn("projects hibernated due to trial grace period ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

		// send email
		err = w.admin.Email.SendTrialGracePeriodEnded(&email.TrialGracePeriodEnded{
			ToEmail:    org.BillingEmail,
			ToName:     org.Name,
			OrgName:    org.Name,
			UpgradeURL: w.admin.URLs.Billing(org.Name, true),
		})
		if err != nil {
			return fmt.Errorf("failed to send trial grace period ended email for org %q: %w", org.Name, err)
		}

		// mark the billing issue as processed
		err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("failed to update billing issue as processed: %w", err)
		}
	}

	return nil
}
