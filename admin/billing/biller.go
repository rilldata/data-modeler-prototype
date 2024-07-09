package billing

import (
	"context"
	"errors"
	"time"

	"github.com/rilldata/rill/admin/database"
)

const (
	SupportEmail    = "support@rilldata.com"
	DefaultTimeZone = "UTC"
)

var ErrNotFound = errors.New("not found")

type Biller interface {
	Name() string
	GetDefaultPlan(ctx context.Context) (*Plan, error)
	GetPlans(ctx context.Context) ([]*Plan, error)
	// GetPublicPlans for listing purposes
	GetPublicPlans(ctx context.Context) ([]*Plan, error)
	// GetPlan returns the plan with the given biller plan ID.
	GetPlan(ctx context.Context, id string) (*Plan, error)
	// GetPlanByName returns the plan with the given Rill plan name.
	GetPlanByName(ctx context.Context, name string) (*Plan, error)

	// CreateCustomer creates a customer for the given organization in the billing system and returns the external customer ID.
	CreateCustomer(ctx context.Context, organization *database.Organization) (string, error)

	// CreateSubscription creates a subscription for the given organization.
	// The subscription starts immediately.
	CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, cancelOption SubscriptionCancellationOption) error
	GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error)
	ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan) (*Subscription, error)
	// CancelSubscriptionsForCustomer deletes the subscription for the given organization.
	// cancellationDate only applicable if option is SubscriptionCancellationOptionRequestedDate
	CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) error

	ReportUsage(ctx context.Context, usage []*Usage) error

	GetReportingGranularity() UsageReportingGranularity
	GetReportingWorkerCron() string
}

type Plan struct {
	ID              string // ID of the plan in the external billing system
	Name            string // Unique name of the plan in Rill, can be empty if biller does not support it
	DisplayName     string
	Description     string
	TrialPeriodDays int
	Default         bool
	Public          bool
	Quotas          Quotas
	Metadata        map[string]string
}

type Quotas struct {
	StorageLimitBytesPerDeployment *int64

	// Existing quotas
	NumProjects           *int
	NumDeployments        *int
	NumSlotsTotal         *int
	NumSlotsPerDeployment *int
	NumOutstandingInvites *int
}

type planMetadata struct {
	Default                        bool   `mapstructure:"default"`
	Public                         bool   `mapstructure:"public"`
	StorageLimitBytesPerDeployment *int64 `mapstructure:"storage_limit_bytes_per_deployment"`
	NumProjects                    *int   `mapstructure:"num_projects"`
	NumDeployments                 *int   `mapstructure:"num_deployments"`
	NumSlotsTotal                  *int   `mapstructure:"num_slots_total"`
	NumSlotsPerDeployment          *int   `mapstructure:"num_slots_per_deployment"`
	NumOutstandingInvites          *int   `mapstructure:"num_outstanding_invites"`
}

type Subscription struct {
	ID                           string
	CustomerID                   string
	Plan                         *Plan
	StartDate                    time.Time
	EndDate                      time.Time
	CurrentBillingCycleStartDate time.Time
	CurrentBillingCycleEndDate   time.Time
	TrialEndDate                 time.Time
	Metadata                     map[string]string
}

type Usage struct {
	CustomerID     string
	MetricName     string
	Value          float64
	ReportingGrain UsageReportingGranularity
	StartTime      time.Time // Start time of the usage period
	EndTime        time.Time // End time of the usage period
	Metadata       map[string]interface{}
}

type UsageReportingGranularity string

const (
	UsageReportingGranularityNone UsageReportingGranularity = ""
	UsageReportingGranularityHour UsageReportingGranularity = "hour"
)

type SubscriptionCancellationOption int

const (
	SubscriptionCancellationOptionEndOfSubscriptionTerm SubscriptionCancellationOption = iota
	SubscriptionCancellationOptionImmediate
)
