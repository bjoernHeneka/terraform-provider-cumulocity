package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Tenant represents a Cumulocity subtenant.
type Tenant struct {
	ID                 string `json:"id,omitempty"`
	Self               string `json:"self,omitempty"`
	AdminEmail         string `json:"adminEmail,omitempty"`
	AdminName          string `json:"adminName,omitempty"`
	Company            string `json:"company,omitempty"`
	ContactName        string `json:"contactName,omitempty"`
	ContactPhone       string `json:"contactPhone,omitempty"`
	Domain             string `json:"domain,omitempty"`
	Parent             string `json:"parent,omitempty"`
	Status             string `json:"status,omitempty"`
	CreationTime       string `json:"creationTime,omitempty"`
	AllowCreateTenants bool   `json:"allowCreateTenants,omitempty"`
}

// TenantCreateRequest holds the fields for creating a subtenant.
type TenantCreateRequest struct {
	AdminEmail   string `json:"adminEmail"`
	AdminName    string `json:"adminName,omitempty"`
	AdminPass    string `json:"adminPass,omitempty"`
	Company      string `json:"company"`
	ContactName  string `json:"contactName,omitempty"`
	ContactPhone string `json:"contactPhone,omitempty"`
	Domain       string `json:"domain"`
}

// TenantUpdateRequest holds the mutable fields for updating a tenant.
type TenantUpdateRequest struct {
	AdminEmail   string `json:"adminEmail,omitempty"`
	AdminName    string `json:"adminName,omitempty"`
	AdminPass    string `json:"adminPass,omitempty"`
	Company      string `json:"company,omitempty"`
	ContactName  string `json:"contactName,omitempty"`
	ContactPhone string `json:"contactPhone,omitempty"`
}

type tenantCollectionPage struct {
	Tenants []Tenant `json:"tenants"`
	Next    string   `json:"next,omitempty"`
}

// ApplicationReference represents a reference to an application.
// Used when retrieving subscribed applications.
type ApplicationReference struct {
	Self        string      `json:"self,omitempty"`
	Application Application `json:"application,omitempty"`
}

// SubscribedApplicationReference is used when subscribing to an application.
// The request body only requires the application self-link.
type SubscribedApplicationReference struct {
	Application ApplicationRef `json:"application"`
}

// ApplicationRef contains the self-link to the application.
type ApplicationRef struct {
	Self string `json:"self"`
}

// CreateTenant creates a new subtenant.
func (c *Client) CreateTenant(ctx context.Context, req TenantCreateRequest) (*Tenant, error) {
	var result Tenant
	if _, err := c.doJSON(ctx, http.MethodPost, "/tenant/tenants", req, &result); err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}
	return &result, nil
}

// GetTenant retrieves a specific tenant by ID.
func (c *Client) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	path := fmt.Sprintf("/tenant/tenants/%s", url.PathEscape(tenantID))
	var result Tenant
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting tenant %q: %w", tenantID, err)
	}
	return &result, nil
}

// UpdateTenant updates a tenant's mutable fields.
func (c *Client) UpdateTenant(ctx context.Context, tenantID string, req TenantUpdateRequest) (*Tenant, error) {
	path := fmt.Sprintf("/tenant/tenants/%s", url.PathEscape(tenantID))
	var result Tenant
	if _, err := c.doJSON(ctx, http.MethodPut, path, req, &result); err != nil {
		return nil, fmt.Errorf("updating tenant %q: %w", tenantID, err)
	}
	return &result, nil
}

// DeleteTenant deletes a tenant by ID. Returns nil if already gone.
func (c *Client) DeleteTenant(ctx context.Context, tenantID string) error {
	path := fmt.Sprintf("/tenant/tenants/%s", url.PathEscape(tenantID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting tenant %q: %w", tenantID, err)
	}
	return nil
}

// SubscribeApplication subscribes a tenant to an application.
func (c *Client) SubscribeApplication(ctx context.Context, tenantID, applicationID string) (*ApplicationReference, error) {
	path := fmt.Sprintf("/tenant/tenants/%s/applications", url.PathEscape(tenantID))

	// Build the application self-link
	appSelfLink := fmt.Sprintf("%s/application/applications/%s", c.BaseURL, url.PathEscape(applicationID))

	req := SubscribedApplicationReference{
		Application: ApplicationRef{
			Self: appSelfLink,
		},
	}

	var result ApplicationReference
	status, err := c.doJSON(ctx, http.MethodPost, path, req, &result)
	if err != nil {
		// 409 means the application is already subscribed - treat this as success
		// and fetch the current subscription
		if status == http.StatusConflict {
			return c.GetSubscribedApplication(ctx, tenantID, applicationID)
		}
		return nil, fmt.Errorf("subscribing tenant %q to application %q: %w", tenantID, applicationID, err)
	}
	return &result, nil
}

// UnsubscribeApplication unsubscribes a tenant from an application.
func (c *Client) UnsubscribeApplication(ctx context.Context, tenantID, applicationID string) error {
	path := fmt.Sprintf("/tenant/tenants/%s/applications/%s", url.PathEscape(tenantID), url.PathEscape(applicationID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unsubscribing tenant %q from application %q: %w", tenantID, applicationID, err)
	}
	return nil
}

// GetSubscribedApplication retrieves subscription details for a specific application.
func (c *Client) GetSubscribedApplication(ctx context.Context, tenantID, applicationID string) (*ApplicationReference, error) {
	// GET /tenant/tenants/{tenantId}/applications returns a paginated collection
	// We need to fetch all pages and find the specific application
	path := fmt.Sprintf("/tenant/tenants/%s/applications?pageSize=100", url.PathEscape(tenantID))

	type applicationReferenceCollection struct {
		References []ApplicationReference `json:"references"`
		Self       string                 `json:"self,omitempty"`
		Next       string                 `json:"next,omitempty"`
	}

	// Follow pagination to find the application
	for path != "" {
		var result applicationReferenceCollection
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
			return nil, fmt.Errorf("getting subscribed applications for tenant %q: %w", tenantID, err)
		}

		// Search in current page
		for _, ref := range result.References {
			if ref.Application.ID == applicationID {
				return &ref, nil
			}
		}

		// Move to next page if available
		path = ""
		if result.Next != "" {
			u, err := url.Parse(result.Next)
			if err == nil {
				path = u.RequestURI()
			}
		}
	}

	return nil, ErrNotFound
}
