package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"go.uber.org/zap"
)

func (s *Service) InitOrganizationBilling(ctx context.Context, org *database.Organization) (*database.Organization, error) {
	// create payment customer
	pc, err := s.PaymentProvider.CreateCustomer(ctx, org)
	if err != nil {
		return nil, err
	}
	s.Logger.Info("created payment customer",
		zap.String("org_id", org.ID),
		zap.String("org_name", org.Name),
		zap.String("payment_customer_id", pc.ID),
		zap.String("user_email", org.BillingEmail),
	)

	org.PaymentCustomerID = pc.ID

	// create billing customer
	bc, err := s.Biller.CreateCustomer(ctx, org, billing.PaymentProviderStripe)
	if err != nil {
		return nil, err
	}
	s.Logger.Info("created billing customer",
		zap.String("org", org.Name),
		zap.String("org_id", org.ID),
		zap.String("billing_customer_id", bc.ID),
		zap.String("payment_customer_id", pc.ID),
		zap.String("user_email", org.BillingEmail),
	)

	org.BillingCustomerID = bc.ID

	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	err = s.RaiseNewOrgBillingIssues(ctx, org.ID, org.CreatedOn, pc.HasPaymentMethod, pc.HasBillableAddress, org.BillingCustomerID == "") // noop biller will have customer id as "" so do not raise never subscribed billing error for them
	if err != nil {
		return nil, err
	}

	return org, nil
}

// RepairOrganizationBilling repairs billing for an organization by checking if customer exists in billing systems, if not creating new. Useful for migrating existing orgs to billing system and in rare case when InitOrganizationBilling fails in the middle
func (s *Service) RepairOrganizationBilling(ctx context.Context, org *database.Organization, initSub bool) (*database.Organization, *billing.Subscription, error) {
	var bc *billing.Customer
	var pc *payment.Customer
	var err error

	bcid := org.BillingCustomerID // for safety in case the method is called with org which has billing customer id, currently there is no such call
	if bcid == "" {
		bcid = org.ID
	}
	bc, err = s.Biller.FindCustomer(ctx, bcid)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return nil, nil, fmt.Errorf("error finding billing customer: %w", err)
	}

	if bc != nil {
		org.BillingCustomerID = bc.ID
		if bc.PaymentProviderID != "" {
			org.PaymentCustomerID = bc.PaymentProviderID
		}
	}

	if org.PaymentCustomerID == "" {
		pc, err = s.PaymentProvider.FindCustomerForOrg(ctx, org)
		if err != nil {
			if errors.Is(err, billing.ErrNotFound) {
				// Create a new customer
				pc, err = s.PaymentProvider.CreateCustomer(ctx, org)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create payment customer: %w", err)
				}
				s.Logger.Info("created payment customer",
					zap.String("org_id", org.ID),
					zap.String("org_name", org.Name),
					zap.String("payment_customer_id", pc.ID),
					zap.String("user_email", pc.Email),
				)
			} else {
				return nil, nil, fmt.Errorf("error finding payment customer: %w", err)
			}
		}
		org.PaymentCustomerID = pc.ID
	}

	if pc == nil {
		pc, err = s.PaymentProvider.FindCustomer(ctx, org.PaymentCustomerID)
		if err != nil {
			return nil, nil, fmt.Errorf("error finding payment customer: %w", err)
		}
	}

	if bc == nil {
		bc, err = s.Biller.CreateCustomer(ctx, org, billing.PaymentProviderStripe)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create billing customer: %w", err)
		}
		s.Logger.Info("created billing customer",
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name),
			zap.String("billing_customer_id", bc.Email),
			zap.String("payment_customer_id", org.PaymentCustomerID),
			zap.String("user_email", org.BillingEmail),
		)
		org.BillingCustomerID = bc.ID
	} else if bc.PaymentProviderID == "" {
		// update payment customer id in billing system
		err = s.Biller.UpdateCustomerPaymentID(ctx, org.BillingCustomerID, billing.PaymentProviderStripe, org.PaymentCustomerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update payment customer id: %w", err)
		}
	}

	// update billing and payment customer id
	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update organization: %w", err)
	}

	sub, err := s.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		if !errors.Is(err, billing.ErrNotFound) {
			return nil, nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
		}
	}

	err = s.RaiseNewOrgBillingIssues(ctx, org.ID, org.CreatedOn, pc.HasPaymentMethod, pc.HasBillableAddress, sub != nil)
	if err != nil {
		return nil, nil, err
	}

	if !initSub {
		return org, nil, nil
	}

	var updatedOrg *database.Organization
	if sub == nil {
		updatedOrg, sub, err = s.StartTrial(ctx, org)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to start trial: %w", err)
		}

		// send trial started email
		err = s.Email.SendTrialStarted(&email.TrialStarted{
			ToEmail:      org.BillingEmail,
			ToName:       org.Name,
			OrgName:      org.Name,
			FrontendURL:  s.URLs.Frontend(),
			TrialEndDate: sub.TrialEndDate,
		})
		if err != nil {
			s.Logger.Named("billing").Error("failed to send trial started email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
		}
	} else {
		s.Logger.Named("billing").Warn("subscription already exists for org", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		// update org quotas, this subscription might have been manually created
		updatedOrg, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			DisplayName:                         org.DisplayName,
			Description:                         org.Description,
			LogoAssetID:                         org.LogoAssetID,
			FaviconAssetID:                      org.FaviconAssetID,
			CustomDomain:                        org.CustomDomain,
			QuotaProjects:                       biggerOfInt(sub.Plan.Quotas.NumProjects, org.QuotaProjects),
			QuotaDeployments:                    biggerOfInt(sub.Plan.Quotas.NumDeployments, org.QuotaDeployments),
			QuotaSlotsTotal:                     biggerOfInt(sub.Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
			QuotaSlotsPerDeployment:             biggerOfInt(sub.Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
			QuotaOutstandingInvites:             biggerOfInt(sub.Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
			QuotaStorageLimitBytesPerDeployment: biggerOfInt64(sub.Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
			BillingEmail:                        org.BillingEmail,
			BillingPlanName:                     org.BillingPlanName,
			BillingPlanDisplayName:              org.BillingPlanDisplayName,
			CreatedByUserID:                     org.CreatedByUserID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update organization: %w", err)
		}
	}

	return updatedOrg, sub, nil
}

func (s *Service) StartTrial(ctx context.Context, org *database.Organization) (*database.Organization, *billing.Subscription, error) {
	// get default plan
	plan, err := s.Biller.GetDefaultPlan(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get default plan: %w", err)
	}

	sub, err := s.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		if !errors.Is(err, billing.ErrNotFound) {
			if errors.Is(err, billing.ErrCustomerIDRequired) {
				return nil, nil, fmt.Errorf("org billing not initialized yet")
			}
			return nil, nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
		}
	}

	if sub != nil && sub.Plan.ID != plan.ID {
		return nil, nil, errors.New("subscription exists with non-trial plan")
	}

	if sub == nil {
		sub, err = s.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
		}

		if org.CreatedByUserID != nil {
			err = s.DB.IncrementCurrentTrialOrgCount(ctx, *org.CreatedByUserID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to increment current trial org count: %w", err)
			}
		}
	}

	if sub.ID == "" || sub.Plan.ID == "" {
		// happens with noop biller
		return org, sub, nil
	}

	s.Logger.Named("billing").Info("started trial for organization",
		zap.String("org_name", org.Name),
		zap.String("org_id", org.ID),
		zap.String("trial_end_date", sub.TrialEndDate.String()),
		zap.String("email", *org.CreatedByUserID),
	)

	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       biggerOfInt(plan.Quotas.NumProjects, org.QuotaProjects),
		QuotaDeployments:                    biggerOfInt(plan.Quotas.NumDeployments, org.QuotaDeployments),
		QuotaSlotsTotal:                     biggerOfInt(plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
		QuotaSlotsPerDeployment:             biggerOfInt(plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
		QuotaOutstandingInvites:             biggerOfInt(plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
		QuotaStorageLimitBytesPerDeployment: biggerOfInt64(plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     &plan.Name,
		BillingPlanDisplayName:              &plan.DisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update organization: %w", err)
	}

	err = s.CleanupSubscriptionBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, nil, err
	}

	err = s.CleanupTrialBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, nil, err
	}

	// raise on-trial billing warning
	_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID: org.ID,
		Type:  database.BillingIssueTypeOnTrial,
		Metadata: &database.BillingIssueMetadataOnTrial{
			SubID:              sub.ID,
			PlanID:             sub.Plan.ID,
			EndDate:            sub.TrialEndDate,
			GracePeriodEndDate: sub.TrialEndDate.AddDate(0, 0, database.BillingGracePeriodDays),
		},
		EventTime: sub.StartDate,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upsert billing warning: %w", err)
	}

	return org, sub, nil
}

// RaiseNewOrgBillingIssues raises billing issues for a new organization
func (s *Service) RaiseNewOrgBillingIssues(ctx context.Context, orgID string, creationTime time.Time, hasPaymentMethod, hasBillableAddress, hasSubscription bool) error {
	if !hasPaymentMethod {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     orgID,
			Type:      database.BillingIssueTypeNoPaymentMethod,
			Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
			EventTime: creationTime,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	if !hasBillableAddress {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     orgID,
			Type:      database.BillingIssueTypeNoBillableAddress,
			Metadata:  &database.BillingIssueMetadataNoBillableAddress{},
			EventTime: creationTime,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	if !hasSubscription {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     orgID,
			Type:      database.BillingIssueTypeNeverSubscribed,
			Metadata:  database.BillingIssueMetadataNeverSubscribed{},
			EventTime: creationTime,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	return nil
}

// CleanupTrialBillingIssues removes trial related billing issues
func (s *Service) CleanupTrialBillingIssues(ctx context.Context, orgID string) error {
	err := s.DB.DeleteBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	err = s.DB.DeleteBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeOnTrial)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	return nil
}

// CleanupSubscriptionBillingIssues removes subscription related billing issues
func (s *Service) CleanupSubscriptionBillingIssues(ctx context.Context, orgID string) error {
	err := s.DB.DeleteBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeNeverSubscribed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	err = s.DB.DeleteBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to delete billing errors: %w", err)
		}
	}

	return nil
}

func (s *Service) CheckBlockingBillingErrors(ctx context.Context, orgID string) error {
	be, err := s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil {
		return fmt.Errorf("trial has ended")
	}

	be, err = s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil {
		earliestGracePeriodEndDate := time.Time{}
		invoices := be.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices
		for _, inv := range invoices {
			if inv.GracePeriodEndDate.Before(earliestGracePeriodEndDate) || earliestGracePeriodEndDate.IsZero() {
				earliestGracePeriodEndDate = inv.GracePeriodEndDate
			}
		}

		if earliestGracePeriodEndDate.Before(time.Now()) || earliestGracePeriodEndDate.IsZero() {
			return fmt.Errorf("payment overdue")
		}
	}

	be, err = s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil && be.Metadata.(*database.BillingIssueMetadataSubscriptionCancelled).EndDate.Before(time.Now()) {
		return fmt.Errorf("subscription cancelled")
	}

	return nil
}

func biggerOfInt(ptr *int, def int) int {
	if ptr == nil {
		return def
	}

	if *ptr < 0 || def < 0 {
		return -1
	}

	if *ptr > def {
		return *ptr
	}

	return def
}

func biggerOfInt64(ptr *int64, def int64) int64 {
	if ptr == nil {
		return def
	}

	if *ptr < 0 || def < 0 {
		return -1
	}

	if *ptr > def {
		return *ptr
	}

	return def
}
