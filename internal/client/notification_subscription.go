package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// NotificationSubscription represents a Cumulocity Notification 2.0 subscription.
type NotificationSubscription struct {
	ID                 string                          `json:"id,omitempty"`
	Self               string                          `json:"self,omitempty"`
	Context            string                          `json:"context,omitempty"`
	Subscription       string                          `json:"subscription,omitempty"`
	Source             *NotificationSubscriptionSource `json:"source,omitempty"`
	SubscriptionFilter *NotificationSubscriptionFilter `json:"subscriptionFilter,omitempty"`
	FragmentsToCopy    []string                        `json:"fragmentsToCopy,omitempty"`
	NonPersistent      bool                            `json:"nonPersistent"`
}

// NotificationSubscriptionSource is the managed object reference in a subscription.
type NotificationSubscriptionSource struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// NotificationSubscriptionFilter holds the filter criteria for a subscription.
type NotificationSubscriptionFilter struct {
	APIs       []string `json:"apis,omitempty"`
	TypeFilter string   `json:"typeFilter,omitempty"`
}

// NotificationSubscriptionRequest is the payload for creating a subscription.
type NotificationSubscriptionRequest struct {
	Context            string                          `json:"context"`
	Subscription       string                          `json:"subscription"`
	Source             *NotificationSubscriptionSource `json:"source,omitempty"`
	SubscriptionFilter *NotificationSubscriptionFilter `json:"subscriptionFilter,omitempty"`
	FragmentsToCopy    []string                        `json:"fragmentsToCopy,omitempty"`
	NonPersistent      bool                            `json:"nonPersistent"`
}

// CreateNotificationSubscription creates a new Notification 2.0 subscription.
func (c *Client) CreateNotificationSubscription(ctx context.Context, req NotificationSubscriptionRequest) (*NotificationSubscription, error) {
	var result NotificationSubscription
	if _, err := c.doJSON(ctx, http.MethodPost, "/notification2/subscriptions", req, &result); err != nil {
		return nil, fmt.Errorf("creating notification subscription: %w", err)
	}
	return &result, nil
}

// GetNotificationSubscription retrieves a subscription by ID.
func (c *Client) GetNotificationSubscription(ctx context.Context, id string) (*NotificationSubscription, error) {
	path := fmt.Sprintf("/notification2/subscriptions/%s", url.PathEscape(id))
	var result NotificationSubscription
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting notification subscription %q: %w", id, err)
	}
	return &result, nil
}

// DeleteNotificationSubscription removes a subscription by ID.
func (c *Client) DeleteNotificationSubscription(ctx context.Context, id string) error {
	path := fmt.Sprintf("/notification2/subscriptions/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting notification subscription %q: %w", id, err)
	}
	return nil
}
