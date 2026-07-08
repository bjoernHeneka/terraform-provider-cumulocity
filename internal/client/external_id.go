package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// ExternalID represents a Cumulocity identity external ID linking a managed object
// to an identifier in an external system.
type ExternalID struct {
	ExternalId    string                   `json:"externalId,omitempty"`
	Type          string                   `json:"type,omitempty"`
	Self          string                   `json:"self,omitempty"`
	ManagedObject *ExternalIDManagedObject `json:"managedObject,omitempty"`
}

// ExternalIDManagedObject holds the back-reference to the managed object.
type ExternalIDManagedObject struct {
	ID   string `json:"id,omitempty"`
	Self string `json:"self,omitempty"`
}

// CreateExternalID creates a new external ID for an existing managed object.
func (c *Client) CreateExternalID(ctx context.Context, managedObjectID string, externalID ExternalID) (*ExternalID, error) {
	path := fmt.Sprintf("/identity/globalIds/%s/externalIds", url.PathEscape(managedObjectID))
	var result ExternalID
	if _, err := c.doJSON(ctx, http.MethodPost, path, externalID, &result); err != nil {
		return nil, fmt.Errorf("creating external ID: %w", err)
	}
	return &result, nil
}

// GetExternalID retrieves an external ID by type and value.
func (c *Client) GetExternalID(ctx context.Context, idType, externalID string) (*ExternalID, error) {
	path := fmt.Sprintf("/identity/externalIds/%s/%s", url.PathEscape(idType), url.PathEscape(externalID))
	var result ExternalID
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting external ID %q/%q: %w", idType, externalID, err)
	}
	return &result, nil
}

// DeleteExternalID removes an external ID. Returns nil if already gone.
func (c *Client) DeleteExternalID(ctx context.Context, idType, externalID string) error {
	path := fmt.Sprintf("/identity/externalIds/%s/%s", url.PathEscape(idType), url.PathEscape(externalID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting external ID %q/%q: %w", idType, externalID, err)
	}
	return nil
}
