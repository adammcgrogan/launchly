package payment

import (
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v81"
	stripesession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
)

type Client struct {
	webhookSecret    string
	starterProductID string
	proProductID     string
}

func New(secretKey, webhookSecret, starterProductID, proProductID string) *Client {
	stripe.Key = secretKey
	return &Client{
		webhookSecret:    webhookSecret,
		starterProductID: starterProductID,
		proProductID:     proProductID,
	}
}

func (c *Client) productIDForPlan(plan string) (string, bool) {
	switch plan {
	case "starter":
		return c.starterProductID, true
	case "pro":
		return c.proProductID, true
	}
	return "", false
}

// CreateCheckoutSession creates a Stripe Checkout session for the given plan and returns
// the session ID and the hosted checkout URL.
func (c *Client) CreateCheckoutSession(plan, customerEmail, successURL, cancelURL string) (sessionID, checkoutURL string, err error) {
	productID, ok := c.productIDForPlan(plan)
	if !ok {
		return "", "", fmt.Errorf("unknown plan: %s", plan)
	}

	priceID, err := c.priceForProduct(productID)
	if err != nil {
		return "", "", err
	}

	params := &stripe.CheckoutSessionParams{
		CustomerEmail: stripe.String(customerEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}

	sess, err := stripesession.New(params)
	if err != nil {
		return "", "", fmt.Errorf("create checkout session: %w", err)
	}
	return sess.ID, sess.URL, nil
}

func (c *Client) priceForProduct(productID string) (string, error) {
	params := &stripe.PriceListParams{
		Product: stripe.String(productID),
		Active:  stripe.Bool(true),
	}
	iter := price.List(params)
	for iter.Next() {
		return iter.Price().ID, nil
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("list prices: %w", err)
	}
	return "", fmt.Errorf("no active price found for product %s", productID)
}

// CancelSubscription immediately cancels a Stripe subscription.
// If the subscription no longer exists in Stripe, it is treated as already cancelled.
func (c *Client) CancelSubscription(subscriptionID string) error {
	_, err := subscription.Cancel(subscriptionID, &stripe.SubscriptionCancelParams{})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil
		}
		return fmt.Errorf("cancel subscription: %w", err)
	}
	return nil
}

// WebhookEvent is a parsed Stripe webhook event.
type WebhookEvent struct {
	ID             string // Stripe event ID — used for idempotency
	Type           string
	SessionID      string // populated for checkout.session.completed
	SubscriptionID string // populated for checkout.session.completed, customer.subscription.deleted, invoice.payment_failed
	CustomerEmail  string // populated for invoice.payment_failed
}

// ParseWebhook verifies the Stripe webhook signature and returns a parsed event.
func (c *Client) ParseWebhook(payload []byte, sigHeader string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, c.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return nil, fmt.Errorf("webhook signature: %w", err)
	}
	we := &WebhookEvent{ID: event.ID, Type: string(event.Type)}
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return nil, fmt.Errorf("unmarshal session: %w", err)
		}
		we.SessionID = sess.ID
		if sess.Subscription != nil {
			we.SubscriptionID = sess.Subscription.ID
		}
	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return nil, fmt.Errorf("unmarshal subscription: %w", err)
		}
		we.SubscriptionID = sub.ID
	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			return nil, fmt.Errorf("unmarshal invoice: %w", err)
		}
		if inv.Subscription != nil {
			we.SubscriptionID = inv.Subscription.ID
		}
		we.CustomerEmail = inv.CustomerEmail
	}
	return we, nil
}
