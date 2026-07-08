package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// TenantOption represents a Cumulocity tenant configuration option (category/key/value tuple).
type TenantOption struct {
	Category string `json:"category,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Self     string `json:"self,omitempty"`
}

type tenantOptionCreateRequest struct {
	Category string `json:"category"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type tenantOptionUpdateRequest struct {
	Value string `json:"value"`
}

type tenantOptionCollectionPage struct {
	Options []TenantOption `json:"options"`
	Next    string         `json:"next,omitempty"`
}

// CreateTenantOption creates a new tenant option.
func (c *Client) CreateTenantOption(ctx context.Context, category, key, value string) (*TenantOption, error) {
	body := tenantOptionCreateRequest{Category: category, Key: key, Value: value}
	var result TenantOption
	if _, err := c.doJSON(ctx, http.MethodPost, "/tenant/options", body, &result); err != nil {
		return nil, fmt.Errorf("creating tenant option %q/%q: %w", category, key, err)
	}
	return &result, nil
}

// GetTenantOption retrieves a specific tenant option by category and key.
func (c *Client) GetTenantOption(ctx context.Context, category, key string) (*TenantOption, error) {
	path := fmt.Sprintf("/tenant/options/%s/%s", url.PathEscape(category), url.PathEscape(key))
	var result TenantOption
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting tenant option %q/%q: %w", category, key, err)
	}
	return &result, nil
}

// UpdateTenantOption updates the value of an existing tenant option.
func (c *Client) UpdateTenantOption(ctx context.Context, category, key, value string) (*TenantOption, error) {
	path := fmt.Sprintf("/tenant/options/%s/%s", url.PathEscape(category), url.PathEscape(key))
	body := tenantOptionUpdateRequest{Value: value}
	var result TenantOption
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating tenant option %q/%q: %w", category, key, err)
	}
	return &result, nil
}

// DeleteTenantOption removes a tenant option. Returns nil if already gone.
func (c *Client) DeleteTenantOption(ctx context.Context, category, key string) error {
	path := fmt.Sprintf("/tenant/options/%s/%s", url.PathEscape(category), url.PathEscape(key))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting tenant option %q/%q: %w", category, key, err)
	}
	return nil
}

// ListTenantOptions retrieves all tenant options, following pagination.
// If category is non-empty, only options from that category are returned.
func (c *Client) ListTenantOptions(ctx context.Context, category string) ([]TenantOption, error) {
	path := "/tenant/options?pageSize=100"
	var all []TenantOption
	for path != "" {
		var page tenantOptionCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing tenant options: %w", err)
		}
		for _, opt := range page.Options {
			if category == "" || opt.Category == category {
				all = append(all, opt)
			}
		}
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
