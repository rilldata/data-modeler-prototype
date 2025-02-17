package jobs

import (
	"context"
	"time"
)

type Client interface {
	Close(ctx context.Context) error
	Work(ctx context.Context) error
	CancelJob(ctx context.Context, jobID int64) error

	// NOTE: Add new job trigger functions here
	ResetAllDeployments(ctx context.Context) (*InsertResult, error)

	// payment provider related jobs
	PaymentMethodAdded(ctx context.Context, methodID, paymentCustomerID, typ string, eventTime time.Time) (*InsertResult, error)
	PaymentMethodRemoved(ctx context.Context, methodID, paymentCustomerID string, eventTime time.Time) (*InsertResult, error)
	CustomerAddressUpdated(ctx context.Context, paymentCustomerID string, eventTime time.Time) (*InsertResult, error)

	// biller related jobs
	PaymentFailed(ctx context.Context, billingCustomerID, invoiceID, invoiceNumber, invoiceURL, amount, currency string, dueDate, failedAt time.Time) (*InsertResult, error)
	PaymentSuccess(ctx context.Context, billingCustomerID, invoiceID string) (*InsertResult, error)

	// org related jobs
	InitOrgBilling(ctx context.Context, orgID string) (*InsertResult, error)
	RepairOrgBilling(ctx context.Context, orgID string) (*InsertResult, error) // biller is just used for deduplication
	StartOrgTrial(ctx context.Context, orgID string) (*InsertResult, error)
	DeleteOrg(ctx context.Context, orgID string) (*InsertResult, error)

	PlanChanged(ctx context.Context, billingCustomerID string) (*InsertResult, error)
}

type InsertResult struct {
	ID        int64
	Duplicate bool
}
