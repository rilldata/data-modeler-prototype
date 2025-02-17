package payment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.uber.org/zap"
)

const maxBodyBytes = int64(65536)

type stripeWebhook struct {
	stripe *Stripe
	jobs   jobs.Client
}

// handleWebhook handles incoming webhook events from Stripe
func (s *stripeWebhook) handleWebhook(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return httputil.Errorf(http.StatusServiceUnavailable, "error reading request body: %w", err)
	}

	endpointSecret := s.stripe.webhookSecret
	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "error verifying webhook signature: %w", err)
	}

	// Handle the event based on its type
	switch event.Type {
	case "payment_method.attached":
		var paymentMethod stripe.PaymentMethod
		if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "error parsing payment method data: %w", err)
		}
		if paymentMethod.Customer == nil {
			// just log warn and send http ok as we can't do anything without customer id
			s.stripe.logger.Warn("no customer info sent for payment_method.attached event", zap.String("event_id", event.ID), zap.Time("event_time", time.UnixMilli(event.Created*1000)))
		} else {
			err = s.handlePaymentMethodAdded(r.Context(), event.ID, &paymentMethod)
			if err != nil {
				return httputil.Errorf(http.StatusInternalServerError, "error handling payment_method.attached event: %w", err)
			}
		}
	case "payment_method.detached":
		var paymentMethod stripe.PaymentMethod
		if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "error parsing payment method data: %w", err)
		}
		if cust, ok := event.Data.PreviousAttributes["customer"]; ok && cust != nil {
			err = s.handlePaymentMethodRemoved(r.Context(), event.ID, cust.(string), &paymentMethod)
			if err != nil {
				return httputil.Errorf(http.StatusInternalServerError, "error handling payment_method.detached event: %w", err)
			}
		} else {
			// just log warn and send http ok as we can't do anything without customer id
			s.stripe.logger.Warn("no customer info sent for payment method detached event", zap.String("event_id", event.ID), zap.Time("event_time", time.UnixMilli(event.Created*1000)))
		}
	case "customer.updated":
		var customer stripe.Customer
		if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "error parsing customer data: %w", err)
		}
		if customer.ID == "" {
			// just log warn and send http ok as we can't do anything without customer id
			s.stripe.logger.Warn("no customer info sent for customer.updated event", zap.String("event_id", event.ID), zap.Time("event_time", time.UnixMilli(event.Created*1000)))
		} else {
			// we just care about address update, so check if address was update and now customer has a valid address
			if _, ok := event.Data.PreviousAttributes["address"]; ok && customer.Address != nil {
				err = s.handleCustomerAddressUpdated(r.Context(), event.ID, time.UnixMilli(event.Created*1000), &customer)
				if err != nil {
					return httputil.Errorf(http.StatusInternalServerError, "error handling customer.updated event: %w", err)
				}
			}
		}
	default:
		s.stripe.logger.Warn("unhandled stripe event type", zap.String("type", string(event.Type)))
	}

	// Acknowledge receipt of the event
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *stripeWebhook) handlePaymentMethodAdded(ctx context.Context, eventID string, method *stripe.PaymentMethod) error {
	res, err := s.jobs.PaymentMethodAdded(ctx, method.ID, method.Customer.ID, string(method.Type), time.UnixMilli(method.Created*1000))
	if err != nil {
		s.stripe.logger.Error("failed to add payment method added job", zap.String("payment_customer_id", method.Customer.ID), zap.Error(err), observability.ZapCtx(ctx))
		return err
	}
	if res.Duplicate {
		s.stripe.logger.Debug("duplicate payment_method.attached event", zap.String("event_id", eventID))
		return nil
	}
	return nil
}

func (s *stripeWebhook) handlePaymentMethodRemoved(ctx context.Context, eventID, customerID string, method *stripe.PaymentMethod) error {
	res, err := s.jobs.PaymentMethodRemoved(ctx, method.ID, customerID, time.UnixMilli(method.Created*1000))
	if err != nil {
		s.stripe.logger.Error("failed to add payment method removed job", zap.String("payment_customer_id", customerID), zap.Error(err), observability.ZapCtx(ctx))
		return err
	}
	if res.Duplicate {
		s.stripe.logger.Debug("duplicate payment_method.detached event", zap.String("event_id", eventID))
		return nil
	}
	return nil
}

func (s *stripeWebhook) handleCustomerAddressUpdated(ctx context.Context, eventID string, eventTime time.Time, customer *stripe.Customer) error {
	res, err := s.jobs.CustomerAddressUpdated(ctx, customer.ID, eventTime)
	if err != nil {
		s.stripe.logger.Error("failed to add customer updated job", zap.String("payment_customer_id", customer.ID), zap.Error(err), observability.ZapCtx(ctx))
		return err
	}
	if res.Duplicate {
		s.stripe.logger.Debug("duplicate customer.updated event", zap.String("event_d", eventID))
		return nil
	}
	return nil
}
