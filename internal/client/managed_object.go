package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// emptyFragment represents a Cumulocity marker fragment whose value is an empty JSON object {}.
// Using a pointer (*emptyFragment) lets us distinguish "absent" (nil) from "present" (&emptyFragment{}).
type emptyFragment struct{}

// ManagedObject is the API object for a Cumulocity managed object (device, group, etc.).
type ManagedObject struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	Owner        string `json:"owner,omitempty"`
	Self         string `json:"self,omitempty"`
	CreationTime string `json:"creationTime,omitempty"`
	LastUpdated  string `json:"lastUpdated,omitempty"`

	// Standard marker fragments — present as {} when set, absent otherwise.
	C8yIsDevice      *emptyFragment `json:"c8y_IsDevice,omitempty"`
	C8yIsDeviceGroup *emptyFragment `json:"c8y_IsDeviceGroup,omitempty"`
}

// managedObjectRequest is the body used for both POST (create) and PUT (update).
type managedObjectRequest struct {
	Name             string         `json:"name,omitempty"`
	Type             string         `json:"type,omitempty"`
	C8yIsDevice      *emptyFragment `json:"c8y_IsDevice,omitempty"`
	C8yIsDeviceGroup *emptyFragment `json:"c8y_IsDeviceGroup,omitempty"`
}

// CreateManagedObject creates a new managed object in the Cumulocity inventory.
func (c *Client) CreateManagedObject(ctx context.Context, name, moType string, isDevice, isDeviceGroup bool) (*ManagedObject, error) {
	body := managedObjectRequest{
		Name: name,
		Type: moType,
	}
	if isDevice {
		body.C8yIsDevice = &emptyFragment{}
	}
	if isDeviceGroup {
		body.C8yIsDeviceGroup = &emptyFragment{}
	}

	var result ManagedObject
	if _, err := c.doJSON(ctx, http.MethodPost, "/inventory/managedObjects", body, &result); err != nil {
		return nil, fmt.Errorf("creating managed object %q: %w", name, err)
	}
	return &result, nil
}

// GetManagedObject retrieves a managed object by its ID.
func (c *Client) GetManagedObject(ctx context.Context, id string) (*ManagedObject, error) {
	path := fmt.Sprintf("/inventory/managedObjects/%s", url.PathEscape(id))
	var result ManagedObject
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting managed object %q: %w", id, err)
	}
	return &result, nil
}

// UpdateManagedObject updates an existing managed object.
func (c *Client) UpdateManagedObject(ctx context.Context, id, name, moType string, isDevice, isDeviceGroup bool) (*ManagedObject, error) {
	path := fmt.Sprintf("/inventory/managedObjects/%s", url.PathEscape(id))
	body := managedObjectRequest{
		Name: name,
		Type: moType,
	}
	if isDevice {
		body.C8yIsDevice = &emptyFragment{}
	}
	if isDeviceGroup {
		body.C8yIsDeviceGroup = &emptyFragment{}
	}

	var result ManagedObject
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating managed object %q: %w", id, err)
	}
	return &result, nil
}

type managedObjectCollectionPage struct {
	ManagedObjects []ManagedObject `json:"managedObjects"`
	Next           string          `json:"next,omitempty"`
}

// ListManagedObjects retrieves managed objects with optional filters. All pages are followed.
func (c *Client) ListManagedObjects(ctx context.Context, moType, fragmentType, query, text, owner string) ([]ManagedObject, error) {
	base := "/inventory/managedObjects?pageSize=100"
	if moType != "" {
		base += "&type=" + url.QueryEscape(moType)
	}
	if fragmentType != "" {
		base += "&fragmentType=" + url.QueryEscape(fragmentType)
	}
	if query != "" {
		base += "&query=" + url.QueryEscape(query)
	}
	if text != "" {
		base += "&text=" + url.QueryEscape(text)
	}
	if owner != "" {
		base += "&owner=" + url.QueryEscape(owner)
	}

	var all []ManagedObject
	path := base
	for path != "" {
		var page managedObjectCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing managed objects: %w", err)
		}
		all = append(all, page.ManagedObjects...)
		path = ""
		if page.Next != "" {
			u, err := url.Parse(page.Next)
			if err == nil {
				path = u.RequestURI()
			}
		}
	}
	return all, nil
}

// DeleteManagedObject removes a managed object. Returns nil if already gone.
func (c *Client) DeleteManagedObject(ctx context.Context, id string) error {
	path := fmt.Sprintf("/inventory/managedObjects/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting managed object %q: %w", id, err)
	}
	return nil
}
