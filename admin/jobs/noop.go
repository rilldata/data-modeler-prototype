package jobs

import (
	"context"
	"time"
)

type noop struct{}

// NewNoopClient returns a new noop client
func NewNoopClient() Client {
	return &noop{}
}

func (n *noop) Close(ctx context.Context) error {
	return nil
}

func (n *noop) Work(ctx context.Context) error {
	return nil
}

func (n *noop) CancelJob(ctx context.Context, jobID int64) error {
	return nil
}

func (n *noop) ResetAllDeployments(ctx context.Context) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) PaymentMethodAdded(ctx context.Context, methodID, paymentCustomerID, typ string, eventTime time.Time) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) PaymentMethodRemoved(ctx context.Context, methodID, paymentCustomerID string, eventTime time.Time) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) CustomerAddressUpdated(ctx context.Context, paymentCustomerID string, eventTime time.Time) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) PaymentFailed(ctx context.Context, billingCustomerID, invoiceID, invoiceNumber, invoiceURL, amount, currency string, dueDate, failedAt time.Time) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) PaymentSuccess(ctx context.Context, billingCustomerID, invoiceID string) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) HandlePlanChangeBillingIssues(ctx context.Context, orgID, subID, planID string, subStartDate time.Time) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) InitOrgBilling(ctx context.Context, orgID string) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) RepairOrgBilling(ctx context.Context, orgID string) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) StartOrgTrial(ctx context.Context, orgID string) (*InsertResult, error) {
	return nil, nil
}

func (n *noop) DeleteOrg(ctx context.Context, orgID string) (*InsertResult, error) {
	return nil, nil
}
