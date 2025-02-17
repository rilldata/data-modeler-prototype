package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/mitchellh/mapstructure"
	"github.com/orbcorp/orb-go"
	"github.com/orbcorp/orb-go/option"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"go.uber.org/zap"
)

const (
	paginationLimit      = 100
	eventIngestBatchSize = 500
	requestTimeout       = 10 * time.Second

	avalaraTaxProvider = "avalara"
	taxJarTaxProvider  = "taxjar"
	noneTaxProvider    = "none"

	avalaraTaxExemptionCode = "R" // code for NON-RESIDENT
)

var ErrCustomerIDRequired = errors.New("customer id is required")

var _ Biller = &Orb{}

type Orb struct {
	client        *orb.Client
	logger        *zap.Logger
	webhookSecret string
	taxProvider   string
}

func NewOrb(logger *zap.Logger, orbKey, webhookSecret, taxProvider string) Biller {
	c := orb.NewClient(option.WithAPIKey(orbKey), option.WithRequestTimeout(requestTimeout))

	return &Orb{client: c, logger: logger, webhookSecret: webhookSecret, taxProvider: taxProvider}
}

func (o *Orb) Name() string {
	return "orb"
}

func (o *Orb) GetDefaultPlan(ctx context.Context) (*Plan, error) {
	plans, err := o.GetPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.Default {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) GetPlans(ctx context.Context) ([]*Plan, error) {
	return o.getAllPlans(ctx)
}

func (o *Orb) GetPublicPlans(ctx context.Context) ([]*Plan, error) {
	all, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	var publicPlans []*Plan
	for _, p := range all {
		if p.Public {
			publicPlans = append(publicPlans, p)
		}
	}
	return publicPlans, nil
}

func (o *Orb) GetPlan(ctx context.Context, id string) (*Plan, error) {
	plans, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) GetPlanByName(ctx context.Context, name string) (*Plan, error) {
	if name == "" {
		return nil, ErrNotFound
	}
	plans, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) CreateCustomer(ctx context.Context, organization *database.Organization, provider PaymentProvider) (*Customer, error) {
	var paymentProviderType orb.CustomerNewParamsPaymentProvider
	switch provider {
	case PaymentProviderStripe:
		paymentProviderType = orb.CustomerNewParamsPaymentProviderStripeCharge
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}

	customer, err := o.client.Customers.New(ctx, orb.CustomerNewParams{
		Email:              orb.String(Email(organization)),
		Name:               orb.String(organization.Name),
		ExternalCustomerID: orb.String(organization.ID),
		Timezone:           orb.String(DefaultTimeZone),
		PaymentProvider:    orb.F(paymentProviderType),
		PaymentProviderID:  orb.String(organization.PaymentCustomerID),
	})
	if err != nil {
		return nil, err
	}

	return getBillingCustomerFromOrbCustomer(customer), nil
}

func (o *Orb) FindCustomer(ctx context.Context, customerID string) (*Customer, error) {
	customer, err := o.client.Customers.FetchByExternalID(ctx, customerID)
	if err != nil {
		var orbErr *orb.Error
		if errors.As(err, &orbErr) {
			if orbErr.Status == orb.ErrorStatus404 {
				return nil, ErrNotFound
			}
		}
		return nil, err
	}

	return getBillingCustomerFromOrbCustomer(customer), nil
}

func (o *Orb) UpdateCustomerPaymentID(ctx context.Context, customerID string, provider PaymentProvider, paymentProviderID string) error {
	var paymentProviderType orb.CustomerUpdateByExternalIDParamsPaymentProvider
	switch provider {
	case PaymentProviderStripe:
		paymentProviderType = orb.CustomerUpdateByExternalIDParamsPaymentProviderStripeCharge
	default:
		return fmt.Errorf("unsupported payment provider: %s", provider)
	}
	_, err := o.client.Customers.UpdateByExternalID(ctx, customerID, orb.CustomerUpdateByExternalIDParams{
		PaymentProvider:   orb.F(paymentProviderType),
		PaymentProviderID: orb.String(paymentProviderID),
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *Orb) UpdateCustomerEmail(ctx context.Context, customerID, email string) error {
	_, err := o.client.Customers.UpdateByExternalID(ctx, customerID, orb.CustomerUpdateByExternalIDParams{
		Email: orb.String(email),
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *Orb) DeleteCustomer(ctx context.Context, customerID string) error {
	c, err := o.client.Customers.FetchByExternalID(ctx, customerID)
	if err != nil {
		return err
	}

	err = o.client.Customers.Delete(ctx, c.ID)
	if err != nil {
		return err
	}
	return nil
}

func (o *Orb) CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error) {
	return o.createSubscription(ctx, customerID, plan)
}

func (o *Orb) GetActiveSubscription(ctx context.Context, customerID string) (*Subscription, error) {
	subs, err := o.getActiveSubscriptions(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if len(subs) == 0 {
		return nil, ErrNotFound
	}

	if len(subs) > 1 {
		return nil, fmt.Errorf("multiple active subscriptions (%d) found for customer %s", len(subs), customerID)
	}

	return subs[0], nil
}

func (o *Orb) ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan) (*Subscription, error) {
	s, err := o.client.Subscriptions.SchedulePlanChange(ctx, subscriptionID, orb.SubscriptionSchedulePlanChangeParams{
		PlanID:       orb.String(plan.ID),
		ChangeOption: orb.F(orb.SubscriptionSchedulePlanChangeParamsChangeOptionImmediate),
	})
	if err != nil {
		return nil, err
	}

	return &Subscription{
		ID:                           s.ID,
		Customer:                     getBillingCustomerFromOrbCustomer(&s.Customer),
		Plan:                         plan,
		StartDate:                    s.StartDate,
		EndDate:                      s.EndDate,
		CurrentBillingCycleStartDate: s.CurrentBillingPeriodStartDate,
		CurrentBillingCycleEndDate:   s.CurrentBillingPeriodEndDate,
		TrialEndDate:                 s.TrialInfo.EndDate,
		Metadata:                     s.Metadata,
	}, nil
}

func (o *Orb) UnscheduleCancellation(ctx context.Context, subscriptionID string) (*Subscription, error) {
	sub, err := o.client.Subscriptions.UnscheduleCancellation(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	return o.getBillingSubscriptionFromOrbSubscription(ctx, sub)
}

func (o *Orb) CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) (time.Time, error) {
	var cancelParams orb.SubscriptionCancelParams
	switch cancelOption {
	case SubscriptionCancellationOptionEndOfSubscriptionTerm:
		cancelParams = orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionEndOfSubscriptionTerm),
		}
	case SubscriptionCancellationOptionImmediate:
		cancelParams = orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionImmediate),
		}
	}

	// cancel all upcoming subscriptions for the customer immediately, there shouldn't be any but just in case
	upcomingSubs, err := o.getUpcomingSubscriptionsForCustomer(ctx, customerID)
	if err != nil {
		return time.Time{}, err
	}
	for _, s := range upcomingSubs {
		_, err := o.client.Subscriptions.Cancel(ctx, s.ID, orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionImmediate),
		})
		if err != nil {
			return time.Time{}, err
		}
	}

	// cancel all active subscriptions for the customer as per the cancel option
	subs, err := o.getActiveSubscriptions(ctx, customerID)
	if err != nil {
		return time.Time{}, err
	}
	cancelDate := time.Time{}
	for _, s := range subs {
		sub, err := o.client.Subscriptions.Cancel(ctx, s.ID, cancelParams)
		if err != nil {
			return time.Time{}, err
		}
		if sub.EndDate.After(cancelDate) {
			cancelDate = sub.EndDate
		}
	}

	return cancelDate, nil
}

func (o *Orb) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	invoice, err := o.client.Invoices.Fetch(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	return getBillingInvoiceFromOrbInvoice(invoice), nil
}

func (o *Orb) IsInvoiceValid(ctx context.Context, invoice *Invoice) bool {
	return !strings.EqualFold(invoice.Status, "void")
}

func (o *Orb) IsInvoicePaid(ctx context.Context, invoice *Invoice) bool {
	return strings.EqualFold(invoice.Status, "paid")
}

func (o *Orb) MarkCustomerTaxExempt(ctx context.Context, customerID string) error {
	switch o.taxProvider {
	case avalaraTaxProvider:
		_, err := o.client.Customers.UpdateByExternalID(ctx, customerID, orb.CustomerUpdateByExternalIDParams{
			TaxConfiguration: orb.F[orb.CustomerUpdateByExternalIDParamsTaxConfigurationUnion](orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewAvalaraTaxConfiguration{
				TaxExempt:        orb.F(true),
				TaxProvider:      orb.F(orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewAvalaraTaxConfigurationTaxProviderAvalara),
				TaxExemptionCode: orb.F(avalaraTaxExemptionCode), // code for NON-RESIDENT
			}),
		})
		if err != nil {
			return err
		}
	case taxJarTaxProvider:
		_, err := o.client.Customers.UpdateByExternalID(ctx, customerID, orb.CustomerUpdateByExternalIDParams{
			TaxConfiguration: orb.F[orb.CustomerUpdateByExternalIDParamsTaxConfigurationUnion](orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewTaxJarConfiguration{
				TaxExempt:   orb.F(true),
				TaxProvider: orb.F(orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewTaxJarConfigurationTaxProviderTaxjar),
				// category option not available in TaxJar config
			}),
		})
		if err != nil {
			return err
		}
	case noneTaxProvider:
		o.logger.Named("billing").Warn("no tax provider is set, cannot mark customer tax exempt", zap.String("customer_id", customerID))
	default:
		o.logger.Error("unsupported tax provider", zap.String("tax_provider", o.taxProvider))
	}

	return nil
}

func (o *Orb) UnmarkCustomerTaxExempt(ctx context.Context, customerID string) error {
	switch o.taxProvider {
	case avalaraTaxProvider:
		_, err := o.client.Customers.UpdateByExternalID(ctx, customerID, orb.CustomerUpdateByExternalIDParams{
			TaxConfiguration: orb.F[orb.CustomerUpdateByExternalIDParamsTaxConfigurationUnion](orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewAvalaraTaxConfiguration{
				TaxExempt:   orb.F(false),
				TaxProvider: orb.F(orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewAvalaraTaxConfigurationTaxProviderAvalara),
			}),
		})
		if err != nil {
			return err
		}
	case taxJarTaxProvider:
		_, err := o.client.Customers.UpdateByExternalID(ctx, customerID, orb.CustomerUpdateByExternalIDParams{
			TaxConfiguration: orb.F[orb.CustomerUpdateByExternalIDParamsTaxConfigurationUnion](orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewTaxJarConfiguration{
				TaxExempt:   orb.F(false),
				TaxProvider: orb.F(orb.CustomerUpdateByExternalIDParamsTaxConfigurationNewTaxJarConfigurationTaxProviderTaxjar),
			}),
		})
		if err != nil {
			return err
		}
	case noneTaxProvider:
		o.logger.Named("billing").Warn("no tax provider is set, cannot unmark customer tax exempt", zap.String("customer_id", customerID))
	default:
		o.logger.Error("unsupported tax provider", zap.String("tax_provider", o.taxProvider))
	}

	return nil
}

func (o *Orb) ReportUsage(ctx context.Context, usage []*Usage) error {
	var orbUsage []orb.EventIngestParamsEvent
	// sync max 500 events at a time
	for _, u := range usage {
		eventName := u.MetricName + "_" + string(u.ReportingGrain)
		// use end time minus 1 second to make sure the event is attributed to the current time bucket
		eventTime := u.EndTime.Add(-1 * time.Second)
		// generate idempotency key using customer id, timestamp, event name and metadata
		idempotencyKey := fmt.Sprintf("%s_%d_%s_%v", u.CustomerID, eventTime.UnixMilli(), eventName, u.Metadata)

		props := make(map[string]interface{}, len(u.Metadata)+1)
		for k, v := range u.Metadata {
			props[k] = v
		}
		props["amount"] = u.Value

		orbUsage = append(orbUsage, orb.EventIngestParamsEvent{
			ExternalCustomerID: orb.String(u.CustomerID),
			EventName:          orb.String(eventName),
			IdempotencyKey:     orb.String(idempotencyKey),
			Timestamp:          orb.F(eventTime),
			Properties:         orb.F[any](props),
		})

		if len(orbUsage) == eventIngestBatchSize {
			err := o.pushUsage(ctx, &orbUsage)
			if err != nil {
				return err
			}
			orbUsage = nil
		}
	}

	if len(orbUsage) > 0 {
		err := o.pushUsage(ctx, &orbUsage)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Orb) GetReportingGranularity() UsageReportingGranularity {
	return UsageReportingGranularityHour
}

func (o *Orb) GetReportingWorkerCron() string {
	// run every hour at around end of the hour
	return "55 * * * *"
}

func (o *Orb) WebhookHandlerFunc(ctx context.Context, jc jobs.Client) httputil.Handler {
	if o.webhookSecret == "" {
		return nil
	}
	ow := &orbWebhook{orb: o, jobs: jc}
	return ow.handleWebhook
}

func (o *Orb) createSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error) {
	sub, err := o.client.Subscriptions.New(ctx, orb.SubscriptionNewParams{
		ExternalCustomerID: orb.String(customerID),
		PlanID:             orb.String(plan.ID),
	})
	if err != nil {
		return nil, err
	}

	return &Subscription{
		ID:                           sub.ID,
		Customer:                     getBillingCustomerFromOrbCustomer(&sub.Customer),
		Plan:                         plan,
		StartDate:                    sub.StartDate,
		EndDate:                      sub.EndDate,
		CurrentBillingCycleStartDate: sub.CurrentBillingPeriodStartDate,
		CurrentBillingCycleEndDate:   sub.CurrentBillingPeriodEndDate,
		TrialEndDate:                 sub.TrialInfo.EndDate,
		Metadata:                     sub.Metadata,
	}, nil
}

func (o *Orb) getSubscriptions(ctx context.Context, customerID string, status orb.SubscriptionListParamsStatus) ([]*Subscription, error) {
	if customerID == "" { // weird behaviour but empty external customer id returns all active subscriptions
		return nil, ErrCustomerIDRequired
	}

	sub, err := o.client.Subscriptions.List(ctx, orb.SubscriptionListParams{
		ExternalCustomerID: orb.String(customerID),
		Status:             orb.F(status),
	})
	if err != nil {
		return nil, err
	}

	var subscriptions []*Subscription
	for i := 0; i < len(sub.Data); i++ {
		s := sub.Data[i]
		billingSub, err := o.getBillingSubscriptionFromOrbSubscription(ctx, &s)
		if err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, billingSub)
	}
	return subscriptions, nil
}

func (o *Orb) getActiveSubscriptions(ctx context.Context, customerID string) ([]*Subscription, error) {
	subs, err := o.getSubscriptions(ctx, customerID, orb.SubscriptionListParamsStatusActive)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (o *Orb) getUpcomingSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error) {
	subs, err := o.getSubscriptions(ctx, customerID, orb.SubscriptionListParamsStatusUpcoming)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (o *Orb) getAllPlans(ctx context.Context) ([]*Plan, error) {
	plans, err := o.client.Plans.List(ctx, orb.PlanListParams{
		Limit:  orb.Int(paginationLimit), // TODO handle pagination, for now don't expect more than 100 plans
		Status: orb.F(orb.PlanListParamsStatusActive),
	})
	if err != nil {
		return nil, err
	}

	var billingPlans []*Plan
	for i := 0; i < len(plans.Data); i++ {
		billingPlan, err := o.getBillingPlanFromOrbPlan(ctx, &plans.Data[i])
		if err != nil {
			return nil, err
		}
		billingPlans = append(billingPlans, billingPlan)
	}
	return billingPlans, nil
}

func (o *Orb) pushUsage(ctx context.Context, usage *[]orb.EventIngestParamsEvent) error {
	re := retrier.New(retrier.ExponentialBackoff(5, 500*time.Millisecond), retryErrClassifier{})
	err := re.RunCtx(ctx, func(ctx context.Context) error {
		resp, err := o.client.Events.Ingest(ctx, orb.EventIngestParams{
			Events: orb.F(*usage),
		})
		if err != nil {
			return err
		}
		if len(resp.ValidationFailed) > 0 {
			errMsg := fmt.Sprintf("validation failure for %d events, showing first few:", len(resp.ValidationFailed))
			for i := 0; i < 5 && i < len(resp.ValidationFailed); i++ {
				errMsg += fmt.Sprintf("\n%s: %s", resp.ValidationFailed[i].IdempotencyKey, resp.ValidationFailed[i].ValidationErrors)
			}
			return errors.New(errMsg)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *Orb) getBillingPlanFromOrbPlan(ctx context.Context, p *orb.Plan) (*Plan, error) {
	if p.BasePlanID != "" {
		// fetch base plan metadata, child plans are auto-created by Orb in case of overrides so will be only one level deep
		basePlan, err := o.client.Plans.Fetch(ctx, p.BasePlanID)
		if err != nil {
			return nil, err
		}
		p.Metadata = basePlan.Metadata
	}

	metadata := &planMetadata{}
	err := mapstructure.WeakDecode(p.Metadata, metadata)
	if err != nil {
		return nil, err
	}

	q := &Quotas{
		StorageLimitBytesPerDeployment: metadata.StorageLimitBytesPerDeployment,
		NumProjects:                    metadata.NumProjects,
		NumDeployments:                 metadata.NumDeployments,
		NumSlotsTotal:                  metadata.NumSlotsTotal,
		NumSlotsPerDeployment:          metadata.NumSlotsPerDeployment,
		NumOutstandingInvites:          metadata.NumOutstandingInvites,
	}

	trialPeriodDays := 0
	if p.TrialConfig.TrialPeriodUnit == orb.PlanTrialConfigTrialPeriodUnitDays {
		trialPeriodDays = int(p.TrialConfig.TrialPeriod)
	}

	billingPlan := &Plan{
		ID:              p.ID,
		Name:            p.ExternalPlanID,
		PlanType:        getPlanType(p.ExternalPlanID),
		DisplayName:     getPlanDisplayName(p.ExternalPlanID),
		Description:     p.Description,
		TrialPeriodDays: trialPeriodDays,
		Default:         metadata.Default,
		Public:          metadata.Public,
		Quotas:          *q,
		Metadata:        p.Metadata,
	}
	return billingPlan, nil
}

func (o *Orb) getBillingSubscriptionFromOrbSubscription(ctx context.Context, s *orb.Subscription) (*Subscription, error) {
	plan, err := o.getBillingPlanFromOrbPlan(ctx, &s.Plan)
	if err != nil {
		return nil, err
	}
	return &Subscription{
		ID:                           s.ID,
		Customer:                     getBillingCustomerFromOrbCustomer(&s.Customer),
		Plan:                         plan,
		StartDate:                    s.StartDate,
		EndDate:                      s.EndDate,
		CurrentBillingCycleStartDate: s.CurrentBillingPeriodStartDate,
		CurrentBillingCycleEndDate:   s.CurrentBillingPeriodEndDate,
		TrialEndDate:                 s.TrialInfo.EndDate,
		Metadata:                     s.Metadata,
	}, nil
}

func getBillingCustomerFromOrbCustomer(c *orb.Customer) *Customer {
	return &Customer{
		ID:                c.ExternalCustomerID,
		Email:             c.Email,
		Name:              c.Name,
		PaymentProviderID: c.PaymentProviderID,
		PortalURL:         c.PortalURL,
	}
}

func getBillingInvoiceFromOrbInvoice(i *orb.Invoice) *Invoice {
	return &Invoice{
		ID:             i.ID,
		Status:         string(i.Status),
		CustomerID:     i.Customer.ExternalCustomerID,
		Amount:         i.AmountDue,
		Currency:       i.Currency,
		DueDate:        i.DueDate,
		CreatedAt:      i.CreatedAt,
		SubscriptionID: i.Subscription.ID,
		Metadata:       map[string]interface{}{"issued_at": i.IssuedAt, "voided_at": i.VoidedAt, "paid_at": i.PaidAt, "payment_failed_at": i.PaymentFailedAt},
	}
}

func getPlanType(externalID string) PlanType {
	switch externalID {
	case "free_trial":
		return TrailPlanType
	case "team":
		return TeamPlanType
	case "managed":
		return ManagedPlanType
	default:
		return EnterprisePlanType
	}
}

func getPlanDisplayName(externalID string) string {
	switch externalID {
	case "free_trial":
		return "Free trial"
	case "team":
		return "Team"
	case "managed":
		return "Managed"
	default:
		return "Enterprise"
	}
}

// retryErrClassifier classifies 429 and 500 errors as retryable and all other errors as non retryable
type retryErrClassifier struct{}

func (retryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	var orbErr *orb.Error
	if errors.As(err, &orbErr) {
		if orbErr.Status == orb.ErrorStatus500 || orbErr.Status == orb.ErrorStatus429 {
			return retrier.Retry
		}
	} else {
		return retrier.Fail
	}

	return retrier.Fail
}
