package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetBillingSubscription(ctx context.Context, req *adminv1.GetBillingSubscriptionRequest) (*adminv1.GetBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org subscriptions")
	}

	if org.BillingCustomerID == "" {
		return &adminv1.GetBillingSubscriptionResponse{Organization: s.organizationToDTO(org, true)}, nil
	}

	sub, org, err := s.getSubscriptionAndUpdateOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	if sub == nil {
		return &adminv1.GetBillingSubscriptionResponse{Organization: s.organizationToDTO(org, true)}, nil
	}

	return &adminv1.GetBillingSubscriptionResponse{
		Organization:     s.organizationToDTO(org, true),
		Subscription:     subscriptionToDTO(sub),
		BillingPortalUrl: sub.Customer.PortalURL,
	}, nil
}

func (s *Server) UpdateBillingSubscription(ctx context.Context, req *adminv1.UpdateBillingSubscriptionRequest) (*adminv1.UpdateBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	observability.AddRequestAttributes(ctx, attribute.String("args.plan_name", req.PlanName))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org billing plan")
	}

	if req.PlanName == "" {
		return nil, status.Error(codes.InvalidArgument, "plan name must be provided")
	}

	if org.BillingCustomerID == "" {
		return nil, status.Error(codes.FailedPrecondition, "billing not yet initialized for the organization")
	}

	bisc, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
	}

	if bisc != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "plan cannot be changed on existing subscription as it was cancelled, please renew the subscription")
	}

	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess

	plan, err := s.admin.Biller.GetPlanByName(ctx, req.PlanName)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "plan not found")
		}
		return nil, err
	}
	// if its a trial plan, start trial only if its a new org
	if plan.Default {
		bi, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeNeverSubscribed)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Errorf(codes.FailedPrecondition, "only new organizations can subscribe to the trial plan %s", plan.Name)
			}
			return nil, err
		}
		if bi != nil {
			// check against trial orgs quota, skip for superusers
			if org.CreatedByUserID != nil && !claims.Superuser(ctx) {
				u, err := s.admin.DB.FindUser(ctx, *org.CreatedByUserID)
				if err != nil {
					return nil, err
				}
				if u.QuotaTrialOrgs >= 0 && u.CurrentTrialOrgsCount >= u.QuotaTrialOrgs {
					return nil, status.Errorf(codes.FailedPrecondition, "trial orgs quota of %d reached for user %s", u.QuotaTrialOrgs, u.Email)
				}
			}

			updatedOrg, sub, err := s.admin.StartTrial(ctx, org)
			if err != nil {
				return nil, err
			}

			// send trial started email
			err = s.admin.Email.SendTrialStarted(&email.TrialStarted{
				ToEmail:      org.BillingEmail,
				ToName:       org.Name,
				OrgName:      org.Name,
				FrontendURL:  s.admin.URLs.Frontend(),
				TrialEndDate: sub.TrialEndDate,
			})
			if err != nil {
				s.logger.Named("billing").Error("failed to send trial started email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
			}

			return &adminv1.UpdateBillingSubscriptionResponse{
				Organization: s.organizationToDTO(updatedOrg, true),
				Subscription: subscriptionToDTO(sub),
			}, nil
		}
	}

	if !plan.Public && !forceAccess {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot assign a private plan %q", plan.Name)
	}

	// check for validation errors
	err = s.planChangeValidationChecks(ctx, org, forceAccess)
	if err != nil {
		return nil, err
	}

	if planDowngrade(plan, org) {
		if !forceAccess {
			return nil, status.Errorf(codes.FailedPrecondition, "plan downgrade not supported")
		}
		s.logger.Named("billing").Warn("plan downgrade request", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("plan_name", plan.Name))
	}

	sub, err := s.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		if !errors.Is(err, billing.ErrNotFound) {
			return nil, err
		}
	}

	planChange := false
	if sub == nil {
		// create new subscription
		sub, err = s.admin.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, err
		}
		planChange = true
		s.logger.Named("billing").Info("new subscription created",
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name),
			zap.String("plan_id", sub.Plan.ID),
			zap.String("plan_name", sub.Plan.Name),
		)
	} else {
		// schedule plan change
		oldPlan := sub.Plan
		if oldPlan.ID != plan.ID {
			sub, err = s.admin.Biller.ChangeSubscriptionPlan(ctx, sub.ID, plan)
			if err != nil {
				return nil, err
			}
			planChange = true
			s.logger.Named("billing").Info("plan changed",
				zap.String("org_id", org.ID),
				zap.String("org_name", org.Name),
				zap.String("old_plan_id", oldPlan.ID),
				zap.String("old_plan_name", oldPlan.Name),
				zap.String("new_plan_id", sub.Plan.ID),
				zap.String("new_plan_name", sub.Plan.Name),
			)
		}
	}

	org, err = s.updateQuotasAndHandleBillingIssues(ctx, org, sub)
	if err != nil {
		return nil, err
	}

	if planChange {
		// send plan changed email

		if plan.PlanType == billing.TeamPlanType {
			s.logger.Named("billing").Info("upgraded to team plan",
				zap.String("org_id", org.ID),
				zap.String("org_name", org.Name),
				zap.String("plan_id", sub.Plan.ID),
				zap.String("plan_name", sub.Plan.Name),
			)

			// special handling for team plan to send custom email
			err = s.admin.Email.SendTeamPlanStarted(&email.TeamPlan{
				ToEmail:          org.BillingEmail,
				ToName:           org.Name,
				OrgName:          org.Name,
				FrontendURL:      s.admin.URLs.Frontend(),
				PlanName:         plan.DisplayName,
				BillingStartDate: sub.CurrentBillingCycleEndDate,
			})
		} else {
			err = s.admin.Email.SendPlanUpdate(&email.PlanUpdate{
				ToEmail:  org.BillingEmail,
				ToName:   org.Name,
				OrgName:  org.Name,
				PlanName: plan.DisplayName,
			})
		}

		if err != nil {
			s.logger.Named("billing").Error("failed to send plan update email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
		}
	}

	return &adminv1.UpdateBillingSubscriptionResponse{
		Organization: s.organizationToDTO(org, true),
		Subscription: subscriptionToDTO(sub),
	}, nil
}

// CancelBillingSubscription cancels the billing subscription for the organization
func (s *Server) CancelBillingSubscription(ctx context.Context, req *adminv1.CancelBillingSubscriptionRequest) (*adminv1.CancelBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to cancel org subscription")
	}

	if org.BillingCustomerID == "" {
		return nil, status.Error(codes.FailedPrecondition, "billing not yet initialized for the organization")
	}

	sub, err := s.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, err
	}

	endDate, err := s.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionEndOfSubscriptionTerm)
	if err != nil {
		return nil, err
	}

	if !endDate.IsZero() {
		// raise a billing issue of the subscription cancellation
		_, err = s.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeSubscriptionCancelled,
			Metadata: database.BillingIssueMetadataSubscriptionCancelled{
				EndDate: endDate,
			},
			EventTime: time.Now(),
		})
		if err != nil {
			return nil, err
		}
	}

	// clean up any trial related billing issues if present
	err = s.admin.CleanupTrialBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Named("billing").Warn("subscription cancelled", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	err = s.admin.Email.SendSubscriptionCancelled(&email.SubscriptionCancelled{
		ToEmail:    org.BillingEmail,
		ToName:     org.Name,
		OrgName:    org.Name,
		PlanName:   sub.Plan.DisplayName,
		EndDate:    endDate,
		BillingURL: s.admin.URLs.Billing(org.Name, false),
	})
	if err != nil {
		s.logger.Named("billing").Error("failed to send subscription cancelled email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
	}

	return &adminv1.CancelBillingSubscriptionResponse{}, nil
}

func (s *Server) RenewBillingSubscription(ctx context.Context, req *adminv1.RenewBillingSubscriptionRequest) (*adminv1.RenewBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	observability.AddRequestAttributes(ctx, attribute.String("args.plan_name", req.PlanName))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to renew org subscription")
	}

	if org.BillingCustomerID == "" {
		return nil, status.Error(codes.FailedPrecondition, "billing not yet initialized for the organization")
	}

	bisc, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Errorf(codes.FailedPrecondition, "subscription not cancelled for the organization %s", org.Name)
		}
		return nil, err
	}

	plan, err := s.admin.Biller.GetPlanByName(ctx, req.PlanName)
	if err != nil {
		return nil, err
	}

	if plan.Default {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot renew to trial plan %s", plan.Name)
	}

	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess

	if !plan.Public && !forceAccess {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot renew to a private plan %q", plan.Name)
	}

	// check for validation errors
	err = s.planChangeValidationChecks(ctx, org, forceAccess)
	if err != nil {
		return nil, err
	}

	sub, err := s.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		if !errors.Is(err, billing.ErrNotFound) {
			return nil, err
		}
	}

	if sub == nil {
		sub, err = s.admin.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, err
		}
	} else if sub.EndDate == sub.CurrentBillingCycleEndDate {
		// To make request idempotent, if subscription is still on cancellation schedule, unschedule it
		sub, err = s.admin.Biller.UnscheduleCancellation(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
	}

	if sub.Plan.ID != plan.ID {
		// change the plan, won't happen for new subscriptions
		sub, err = s.admin.Biller.ChangeSubscriptionPlan(ctx, sub.ID, plan)
		if err != nil {
			return nil, err
		}
	}

	// update quotas
	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       valOrDefault(sub.Plan.Quotas.NumProjects, org.QuotaProjects),
		QuotaDeployments:                    valOrDefault(sub.Plan.Quotas.NumDeployments, org.QuotaDeployments),
		QuotaSlotsTotal:                     valOrDefault(sub.Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
		QuotaSlotsPerDeployment:             valOrDefault(sub.Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
		QuotaOutstandingInvites:             valOrDefault(sub.Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
		QuotaStorageLimitBytesPerDeployment: valOrDefault(sub.Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		BillingPlanName:                     &sub.Plan.Name,
		BillingPlanDisplayName:              &sub.Plan.DisplayName,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}

	// delete the billing issue
	err = s.admin.DB.DeleteBillingIssue(ctx, bisc.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Named("billing").Info("subscription renewed", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("plan_id", sub.Plan.ID), zap.String("plan_name", sub.Plan.Name))

	// send subscription renewed email
	if sub.Plan.PlanType == billing.TeamPlanType {
		// special handling for team plan to send custom email
		err = s.admin.Email.SendTeamPlanRenewal(&email.TeamPlan{
			ToEmail:          org.BillingEmail,
			ToName:           org.Name,
			OrgName:          org.Name,
			FrontendURL:      s.admin.URLs.Frontend(),
			PlanName:         sub.Plan.DisplayName,
			BillingStartDate: sub.CurrentBillingCycleEndDate,
		})
	} else {
		err = s.admin.Email.SendSubscriptionRenewed(&email.SubscriptionRenewed{
			ToEmail:  org.BillingEmail,
			ToName:   org.Name,
			OrgName:  org.Name,
			PlanName: sub.Plan.DisplayName,
		})
	}
	if err != nil {
		s.logger.Named("billing").Error("failed to send subscription renewed email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.Error(err))
	}

	return &adminv1.RenewBillingSubscriptionResponse{
		Organization: s.organizationToDTO(org, true),
		Subscription: subscriptionToDTO(sub),
	}, nil
}

func (s *Server) GetPaymentsPortalURL(ctx context.Context, req *adminv1.GetPaymentsPortalURLRequest) (*adminv1.GetPaymentsPortalURLResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	observability.AddRequestAttributes(ctx, attribute.String("args.return_url", req.ReturnUrl))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage org billing")
	}

	if org.PaymentCustomerID == "" {
		return nil, status.Error(codes.FailedPrecondition, "payment customer not initialized yet for the organization")
	}

	// returnUrl is mandatory so if not passed default to home page
	if req.ReturnUrl == "" {
		req.ReturnUrl = s.admin.URLs.Frontend()
	}

	url, err := s.admin.PaymentProvider.GetBillingPortalURL(ctx, org.PaymentCustomerID, req.ReturnUrl)
	if err != nil {
		return nil, err
	}

	return &adminv1.GetPaymentsPortalURLResponse{Url: url}, nil
}

// SudoUpdateOrganizationBillingCustomer updates the billing customer id for an organization. May be useful if customer is initialized manually in billing system
func (s *Server) SudoUpdateOrganizationBillingCustomer(ctx context.Context, req *adminv1.SudoUpdateOrganizationBillingCustomerRequest) (*adminv1.SudoUpdateOrganizationBillingCustomerResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)
	if req.BillingCustomerId != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.billing_customer_id", *req.BillingCustomerId))
	}
	if req.PaymentCustomerId != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.payment_customer_id", *req.PaymentCustomerId))
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage billing customer")
	}

	if req.BillingCustomerId == nil && req.PaymentCustomerId == nil {
		return nil, status.Error(codes.InvalidArgument, "either or both billing and payment customer id must be provided")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, err
	}

	opts := &database.UpdateOrganizationOptions{
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
		BillingCustomerID:                   valOrDefault(req.BillingCustomerId, org.BillingCustomerID),
		PaymentCustomerID:                   valOrDefault(req.PaymentCustomerId, org.PaymentCustomerID),
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	}

	var sub *billing.Subscription
	if req.BillingCustomerId != nil {
		// get active subscriptions if present
		sub, err = s.admin.Biller.GetActiveSubscription(ctx, *req.BillingCustomerId)
		if err != nil {
			if !errors.Is(err, billing.ErrNotFound) {
				return nil, err
			}
		}

		if sub != nil {
			opts.QuotaProjects = biggerOfInt(sub.Plan.Quotas.NumProjects, org.QuotaProjects)
			opts.QuotaDeployments = biggerOfInt(sub.Plan.Quotas.NumDeployments, org.QuotaDeployments)
			opts.QuotaSlotsTotal = biggerOfInt(sub.Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal)
			opts.QuotaSlotsPerDeployment = biggerOfInt(sub.Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment)
			opts.QuotaOutstandingInvites = biggerOfInt(sub.Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites)
			opts.QuotaStorageLimitBytesPerDeployment = biggerOfInt64(sub.Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment)
		}
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, opts)
	if err != nil {
		return nil, err
	}

	if req.PaymentCustomerId != nil {
		// fetch the customer
		pc, err := s.admin.PaymentProvider.FindCustomer(ctx, *req.PaymentCustomerId)
		if err != nil {
			return nil, err
		}

		// link the payment customer to the billing customer
		err = s.admin.Biller.UpdateCustomerPaymentID(ctx, org.BillingCustomerID, billing.PaymentProviderStripe, *req.PaymentCustomerId)
		if err != nil {
			return nil, err
		}

		if !pc.HasPaymentMethod {
			_, err := s.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
				OrgID:     org.ID,
				Type:      database.BillingIssueTypeNoPaymentMethod,
				Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
				EventTime: time.Now(),
			})
			if err != nil {
				return nil, err
			}
		}

		if !pc.HasBillableAddress {
			_, err := s.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
				OrgID:     org.ID,
				Type:      database.BillingIssueTypeNoBillableAddress,
				Metadata:  &database.BillingIssueMetadataNoBillableAddress{},
				EventTime: time.Now(),
			})
			if err != nil {
				return nil, err
			}
		}
	}

	if sub == nil {
		return &adminv1.SudoUpdateOrganizationBillingCustomerResponse{
			Organization: s.organizationToDTO(org, true),
		}, nil
	}

	return &adminv1.SudoUpdateOrganizationBillingCustomerResponse{
		Organization: s.organizationToDTO(org, true),
		Subscription: subscriptionToDTO(sub),
	}, nil
}

func (s *Server) SudoExtendTrial(ctx context.Context, req *adminv1.SudoExtendTrialRequest) (*adminv1.SudoExtendTrialResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	days := int(req.Days)
	observability.AddRequestAttributes(ctx, attribute.Int("args.days", days))

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can extend trial")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ns, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeNeverSubscribed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
	}

	if ns != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "organization %s never subscribed to a plan", org.Name)
	}

	// find existing trial end date
	currentEndDate := time.Time{}
	onTrial, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeOnTrial)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
	}
	if onTrial != nil {
		currentEndDate = onTrial.Metadata.(*database.BillingIssueMetadataOnTrial).GracePeriodEndDate
	}

	if currentEndDate.IsZero() {
		trialEnded, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeTrialEnded)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return nil, err
			}
		}
		if trialEnded != nil {
			currentEndDate = trialEnded.Metadata.(*database.BillingIssueMetadataTrialEnded).GracePeriodEndDate
		}
	}

	if currentEndDate.IsZero() {
		subCancelled, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeSubscriptionCancelled)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return nil, err
			}
		}
		if subCancelled != nil {
			currentEndDate = subCancelled.Metadata.(*database.BillingIssueMetadataSubscriptionCancelled).EndDate
		}
	}

	if currentEndDate.IsZero() || currentEndDate.Before(time.Now()) {
		currentEndDate = time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	}

	newEndDate := currentEndDate.AddDate(0, 0, days)

	// start a new trial, if already on trial plan, this will not create new subscription, if not on trial plan it will error
	_, sub, err := s.admin.StartTrial(ctx, org)
	if err != nil {
		return nil, err
	}

	if sub.ID != "" {
		// update on trial metadata with new end date
		_, err = s.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeOnTrial,
			Metadata: database.BillingIssueMetadataOnTrial{
				SubID:              sub.ID,
				PlanID:             sub.Plan.ID,
				EndDate:            newEndDate,
				GracePeriodEndDate: newEndDate,
			},
			EventTime: time.Now(),
		})
		if err != nil {
			return nil, err
		}

		// send trial extended email
		err = s.admin.Email.SendTrialExtended(&email.TrialExtended{
			ToEmail:      org.BillingEmail,
			ToName:       org.Name,
			OrgName:      org.Name,
			TrialEndDate: newEndDate,
		})
		if err != nil {
			s.logger.Named("billing").Error("failed to send trial extended email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
		}
	}

	// if trial subscription was cancelled then unschedule the cancellation
	if sub.EndDate == sub.CurrentBillingCycleEndDate {
		// if trial subscription was cancelled then unschedule the cancellation
		_, err = s.admin.Biller.UnscheduleCancellation(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
	}

	return &adminv1.SudoExtendTrialResponse{TrialEnd: timestamppb.New(newEndDate)}, nil
}

func (s *Server) SudoTriggerBillingRepair(ctx context.Context, req *adminv1.SudoTriggerBillingRepairRequest) (*adminv1.SudoTriggerBillingRepairResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can trigger billing repair")
	}

	ids, err := s.admin.DB.FindOrganizationIDsWithoutBilling(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations without billing id: %w", err)
	}

	for _, orgID := range ids {
		_, err := s.admin.Jobs.RepairOrgBilling(ctx, orgID)
		if err != nil {
			s.logger.Named("billing").Error("failed to submit repair billing job", zap.String("org_id", orgID), zap.Error(err))
			continue
		}
	}

	return &adminv1.SudoTriggerBillingRepairResponse{}, nil
}

func (s *Server) ListPublicBillingPlans(ctx context.Context, req *adminv1.ListPublicBillingPlansRequest) (*adminv1.ListPublicBillingPlansResponse, error) {
	observability.AddRequestAttributes(ctx)

	// no permissions required to list public billing plans
	plans, err := s.admin.Biller.GetPublicPlans(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []*adminv1.BillingPlan
	for _, plan := range plans {
		dtos = append(dtos, billingPlanToDTO(plan))
	}

	return &adminv1.ListPublicBillingPlansResponse{
		Plans: dtos,
	}, nil
}

func (s *Server) GetBillingProjectCredentials(ctx context.Context, req *adminv1.GetBillingProjectCredentialsRequest) (*adminv1.GetBillingProjectCredentialsResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to get metrics for this org")
	}

	if s.admin.MetricsProjectID == "" {
		return nil, status.Error(codes.FailedPrecondition, "metrics project not configured")
	}

	metricsProj, err := s.admin.DB.FindProject(ctx, s.admin.MetricsProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if metricsProj.ProdDeploymentID == nil {
		return nil, status.Error(codes.InvalidArgument, "project does not have a deployment")
	}

	prodDepl, err := s.admin.DB.FindDeployment(ctx, *metricsProj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Generate JWT
	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: prodDepl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         runtimeAccessTokenDefaultTTL,
		InstancePermissions: map[string][]runtimeauth.Permission{
			prodDepl.RuntimeInstanceID: {
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadAPI,
			},
		},
		Attributes: map[string]any{"organization_id": org.ID, "is_embed": true},
	})
	if err != nil {
		return nil, fmt.Errorf("could not issue jwt: %w", err)
	}

	s.admin.Used.Deployment(prodDepl.ID)

	return &adminv1.GetBillingProjectCredentialsResponse{
		RuntimeHost: prodDepl.RuntimeHost,
		InstanceId:  prodDepl.RuntimeInstanceID,
		AccessToken: jwt,
		TtlSeconds:  uint32(runtimeAccessTokenDefaultTTL.Seconds()),
	}, nil
}

func (s *Server) ListOrganizationBillingIssues(ctx context.Context, req *adminv1.ListOrganizationBillingIssuesRequest) (*adminv1.ListOrganizationBillingIssuesResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org billing errors")
	}

	issues, err := s.admin.DB.FindBillingIssuesForOrg(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	var dtos []*adminv1.BillingIssue
	for _, i := range issues {
		dtos = append(dtos, &adminv1.BillingIssue{
			Organization: org.Name,
			Type:         billingIssueTypeToDTO(i.Type),
			Level:        billingIssueLevelToDTO(i.Level),
			Metadata:     billingIssueMetadataToDTO(i.Type, i.Metadata),
			EventTime:    timestamppb.New(i.EventTime),
			CreatedOn:    timestamppb.New(i.CreatedOn),
		})
	}

	return &adminv1.ListOrganizationBillingIssuesResponse{
		Issues: dtos,
	}, nil
}

func (s *Server) SudoDeleteOrganizationBillingIssue(ctx context.Context, req *adminv1.SudoDeleteOrganizationBillingIssueRequest) (*adminv1.SudoDeleteOrganizationBillingIssueResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization), attribute.String("args.type", req.Type.String()))

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can delete billing errors")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	t, err := dtoBillingIssueTypeToDB(req.Type)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteBillingIssueByTypeForOrg(ctx, org.ID, t)
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoDeleteOrganizationBillingIssueResponse{}, nil
}

func (s *Server) updateQuotasAndHandleBillingIssues(ctx context.Context, org *database.Organization, sub *billing.Subscription) (*database.Organization, error) {
	org, err := s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       valOrDefault(sub.Plan.Quotas.NumProjects, org.QuotaProjects),
		QuotaDeployments:                    valOrDefault(sub.Plan.Quotas.NumDeployments, org.QuotaDeployments),
		QuotaSlotsTotal:                     valOrDefault(sub.Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
		QuotaSlotsPerDeployment:             valOrDefault(sub.Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
		QuotaOutstandingInvites:             valOrDefault(sub.Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
		QuotaStorageLimitBytesPerDeployment: valOrDefault(sub.Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		BillingPlanName:                     &sub.Plan.Name,
		BillingPlanDisplayName:              &sub.Plan.DisplayName,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}

	// delete any trial related billing issues, irrespective of the new plan.
	err = s.admin.CleanupTrialBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup trial billing errors and warnings: %w", err)
	}

	// delete any subscription related billing issues
	err = s.admin.CleanupSubscriptionBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup subscription cancellation errors: %w", err)
	}

	return org, nil
}

func (s *Server) planChangeValidationChecks(ctx context.Context, org *database.Organization, forceAccess bool) error {
	// not a trial plan, check for a payment method and a valid billing address
	var validationErrs []string
	pc, err := s.admin.PaymentProvider.FindCustomer(ctx, org.PaymentCustomerID)
	if err != nil {
		return err
	}
	if !pc.HasPaymentMethod {
		validationErrs = append(validationErrs, "no payment method found")
	}

	if !pc.HasBillableAddress {
		validationErrs = append(validationErrs, "no billing address found")
	}

	be, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}
	if be != nil {
		validationErrs = append(validationErrs, "a previous payment is due")
	}

	if len(validationErrs) > 0 && !forceAccess {
		return status.Errorf(codes.FailedPrecondition, "please fix following by visiting billing portal: %s", strings.Join(validationErrs, ", "))
	}

	return nil
}

func (s *Server) getSubscriptionAndUpdateOrg(ctx context.Context, org *database.Organization) (*billing.Subscription, *database.Organization, error) {
	sub, err := s.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return nil, nil, err
	}

	var planDisplayName string
	var planName string
	if sub == nil {
		planDisplayName = ""
		planName = ""
	} else {
		planDisplayName = sub.Plan.DisplayName
		planName = sub.Plan.Name
	}

	// update the cached plan
	if org.BillingPlanName == nil || *org.BillingPlanName != planName {
		org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
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
			BillingPlanName:                     &planName,
			BillingPlanDisplayName:              &planDisplayName,
			CreatedByUserID:                     org.CreatedByUserID,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return sub, org, nil
}

func subscriptionToDTO(sub *billing.Subscription) *adminv1.Subscription {
	return &adminv1.Subscription{
		Id:                           sub.ID,
		Plan:                         billingPlanToDTO(sub.Plan),
		StartDate:                    valOrNullTime(sub.StartDate),
		EndDate:                      valOrNullTime(sub.EndDate),
		CurrentBillingCycleStartDate: valOrNullTime(sub.CurrentBillingCycleStartDate),
		CurrentBillingCycleEndDate:   valOrNullTime(sub.CurrentBillingCycleEndDate),
		TrialEndDate:                 valOrNullTime(sub.TrialEndDate),
	}
}

func billingPlanToDTO(plan *billing.Plan) *adminv1.BillingPlan {
	return &adminv1.BillingPlan{
		Id:              plan.ID,
		Name:            plan.Name,
		PlanType:        planTypeToDTO(plan.PlanType),
		DisplayName:     plan.DisplayName,
		Description:     plan.Description,
		TrialPeriodDays: uint32(plan.TrialPeriodDays),
		Default:         plan.Default,
		Public:          plan.Public,
		Quotas: &adminv1.Quotas{
			Projects:                       valOrEmptyString(plan.Quotas.NumProjects),
			Deployments:                    valOrEmptyString(plan.Quotas.NumDeployments),
			SlotsTotal:                     valOrEmptyString(plan.Quotas.NumSlotsTotal),
			SlotsPerDeployment:             valOrEmptyString(plan.Quotas.NumSlotsPerDeployment),
			OutstandingInvites:             valOrEmptyString(plan.Quotas.NumOutstandingInvites),
			StorageLimitBytesPerDeployment: val64OrEmptyString(plan.Quotas.StorageLimitBytesPerDeployment),
		},
	}
}

func billingIssueTypeToDTO(t database.BillingIssueType) adminv1.BillingIssueType {
	switch t {
	case database.BillingIssueTypeOnTrial:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_ON_TRIAL
	case database.BillingIssueTypeTrialEnded:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_TRIAL_ENDED
	case database.BillingIssueTypeNoPaymentMethod:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD
	case database.BillingIssueTypeNoBillableAddress:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS
	case database.BillingIssueTypePaymentFailed:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_PAYMENT_FAILED
	case database.BillingIssueTypeSubscriptionCancelled:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED
	case database.BillingIssueTypeNeverSubscribed:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NEVER_SUBSCRIBED
	default:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_UNSPECIFIED
	}
}

func billingIssueLevelToDTO(l database.BillingIssueLevel) adminv1.BillingIssueLevel {
	switch l {
	case database.BillingIssueLevelError:
		return adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_ERROR
	case database.BillingIssueLevelWarning:
		return adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_WARNING
	default:
		return adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_UNSPECIFIED
	}
}

func dtoBillingIssueTypeToDB(t adminv1.BillingIssueType) (database.BillingIssueType, error) {
	switch t {
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_ON_TRIAL:
		return database.BillingIssueTypeOnTrial, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_TRIAL_ENDED:
		return database.BillingIssueTypeTrialEnded, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD:
		return database.BillingIssueTypeNoPaymentMethod, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS:
		return database.BillingIssueTypeNoBillableAddress, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_PAYMENT_FAILED:
		return database.BillingIssueTypePaymentFailed, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED:
		return database.BillingIssueTypeSubscriptionCancelled, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NEVER_SUBSCRIBED:
		return database.BillingIssueTypeNeverSubscribed, nil
	default:
		return database.BillingIssueTypeUnspecified, status.Error(codes.InvalidArgument, "invalid billing error type")
	}
}

func planTypeToDTO(t billing.PlanType) adminv1.BillingPlanType {
	switch t {
	case billing.TrailPlanType:
		return adminv1.BillingPlanType_BILLING_PLAN_TYPE_TRIAL
	case billing.TeamPlanType:
		return adminv1.BillingPlanType_BILLING_PLAN_TYPE_TEAM
	case billing.ManagedPlanType:
		return adminv1.BillingPlanType_BILLING_PLAN_TYPE_MANAGED
	case billing.EnterprisePlanType:
		return adminv1.BillingPlanType_BILLING_PLAN_TYPE_ENTERPRISE
	default:
		return adminv1.BillingPlanType_BILLING_PLAN_TYPE_UNSPECIFIED
	}
}

func billingIssueMetadataToDTO(t database.BillingIssueType, m database.BillingIssueMetadata) *adminv1.BillingIssueMetadata {
	switch t {
	case database.BillingIssueTypeOnTrial:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_OnTrial{
				OnTrial: &adminv1.BillingIssueMetadataOnTrial{
					EndDate:            valOrNullTime(m.(*database.BillingIssueMetadataOnTrial).EndDate),
					GracePeriodEndDate: valOrNullTime(m.(*database.BillingIssueMetadataOnTrial).GracePeriodEndDate),
				},
			},
		}
	case database.BillingIssueTypeTrialEnded:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_TrialEnded{
				TrialEnded: &adminv1.BillingIssueMetadataTrialEnded{
					EndDate:            valOrNullTime(m.(*database.BillingIssueMetadataTrialEnded).EndDate),
					GracePeriodEndDate: valOrNullTime(m.(*database.BillingIssueMetadataTrialEnded).GracePeriodEndDate),
				},
			},
		}
	case database.BillingIssueTypeNoPaymentMethod:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_NoPaymentMethod{
				NoPaymentMethod: &adminv1.BillingIssueMetadataNoPaymentMethod{},
			},
		}
	case database.BillingIssueTypeNoBillableAddress:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_NoBillableAddress{
				NoBillableAddress: &adminv1.BillingIssueMetadataNoBillableAddress{},
			},
		}
	case database.BillingIssueTypePaymentFailed:
		paymentFailed := m.(*database.BillingIssueMetadataPaymentFailed)
		invoices := make([]*adminv1.BillingIssueMetadataPaymentFailedMeta, 0)
		for k := range paymentFailed.Invoices {
			invoices = append(invoices, &adminv1.BillingIssueMetadataPaymentFailedMeta{
				InvoiceId:          paymentFailed.Invoices[k].ID,
				InvoiceNumber:      paymentFailed.Invoices[k].Number,
				InvoiceUrl:         paymentFailed.Invoices[k].URL,
				AmountDue:          paymentFailed.Invoices[k].Amount,
				Currency:           paymentFailed.Invoices[k].Currency,
				DueDate:            valOrNullTime(paymentFailed.Invoices[k].DueDate),
				GracePeriodEndDate: valOrNullTime(paymentFailed.Invoices[k].GracePeriodEndDate),
			})
		}
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_PaymentFailed{
				PaymentFailed: &adminv1.BillingIssueMetadataPaymentFailed{
					Invoices: invoices,
				},
			},
		}
	case database.BillingIssueTypeSubscriptionCancelled:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_SubscriptionCancelled{
				SubscriptionCancelled: &adminv1.BillingIssueMetadataSubscriptionCancelled{
					EndDate: valOrNullTime(m.(*database.BillingIssueMetadataSubscriptionCancelled).EndDate),
				},
			},
		}
	case database.BillingIssueTypeNeverSubscribed:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_NeverSubscribed{
				NeverSubscribed: &adminv1.BillingIssueMetadataNeverSubscribed{},
			},
		}
	default:
		return &adminv1.BillingIssueMetadata{}
	}
}

func planDowngrade(newPan *billing.Plan, org *database.Organization) bool {
	// nil or negative values are considered as unlimited
	if comparableInt(newPan.Quotas.NumProjects) < comparableInt(&org.QuotaProjects) {
		return true
	}
	if comparableInt(newPan.Quotas.NumDeployments) < comparableInt(&org.QuotaDeployments) {
		return true
	}
	if comparableInt(newPan.Quotas.NumSlotsTotal) < comparableInt(&org.QuotaSlotsTotal) {
		return true
	}
	if comparableInt(newPan.Quotas.NumSlotsPerDeployment) < comparableInt(&org.QuotaSlotsPerDeployment) {
		return true
	}
	if comparableInt(newPan.Quotas.NumOutstandingInvites) < comparableInt(&org.QuotaOutstandingInvites) {
		return true
	}
	if comparableInt64(newPan.Quotas.StorageLimitBytesPerDeployment) < comparableInt64(&org.QuotaStorageLimitBytesPerDeployment) {
		return true
	}
	return false
}

func comparableInt(v *int) int {
	if v == nil || *v < 0 {
		return math.MaxInt
	}
	return *v
}

func comparableInt64(v *int64) int64 {
	if v == nil || *v < 0 {
		return math.MaxInt64
	}
	return *v
}

func valOrNullTime(v time.Time) *timestamppb.Timestamp {
	if v.IsZero() {
		return nil
	}
	return timestamppb.New(v)
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
