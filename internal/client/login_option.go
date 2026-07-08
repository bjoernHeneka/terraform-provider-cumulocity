package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// LoginOption represents a Cumulocity authentication configuration (authConfig).
//
// Only the broadly-useful scalar fields are typed explicitly. The full server
// payload — including type-specific nested structures like tokenRequest,
// authorizationRequest, onNewUser.dynamicMapping, signatureVerificationConfig,
// etc. — is preserved verbatim in Raw so callers can expose it as raw JSON.
type LoginOption struct {
	ID                   string `json:"id,omitempty"`
	Self                 string `json:"self,omitempty"`
	Type                 string `json:"type,omitempty"`
	ProviderName         string `json:"providerName,omitempty"`
	GrantType            string `json:"grantType,omitempty"`
	UserManagementSource string `json:"userManagementSource,omitempty"`
	VisibleOnLoginPage   bool   `json:"visibleOnLoginPage,omitempty"`

	// Common OAuth2/SSO scalar fields.
	Template           string `json:"template,omitempty"`
	ButtonName         string `json:"buttonName,omitempty"`
	Issuer             string `json:"issuer,omitempty"`
	ClientID           string `json:"clientId,omitempty"`
	Audience           string `json:"audience,omitempty"`
	RedirectToPlatform string `json:"redirectToPlatform,omitempty"`
	UseIDToken         bool   `json:"useIdToken,omitempty"`

	// Raw holds the exact JSON object returned by the API (verbatim bytes),
	// including all nested/type-specific fields not modelled above. It is
	// populated on unmarshal and never sent back on create/update.
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON decodes the known scalar fields and, in addition, captures the
// complete raw JSON object into Raw for lossless exposure to callers.
func (o *LoginOption) UnmarshalJSON(data []byte) error {
	// alias avoids recursion: the defined type does not carry this method.
	type alias LoginOption
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*o = LoginOption(a)
	o.Raw = append(json.RawMessage(nil), data...)
	return nil
}

// CreateLoginOption creates a new login option.
func (c *Client) CreateLoginOption(ctx context.Context, opt LoginOption) (*LoginOption, error) {
	var result LoginOption
	if _, err := c.doJSON(ctx, http.MethodPost, "/tenant/loginOptions", opt, &result); err != nil {
		return nil, fmt.Errorf("creating login option: %w", err)
	}
	return &result, nil
}

// loginOptionCollectionPage models the GET /tenant/loginOptions collection response.
type loginOptionCollectionPage struct {
	LoginOptions []LoginOption `json:"loginOptions"`
}

// ListLoginOptions retrieves all login options configured on the tenant.
func (c *Client) ListLoginOptions(ctx context.Context) ([]LoginOption, error) {
	var page loginOptionCollectionPage
	if _, err := c.doJSON(ctx, http.MethodGet, "/tenant/loginOptions", nil, &page); err != nil {
		return nil, fmt.Errorf("listing login options: %w", err)
	}
	return page.LoginOptions, nil
}

// CreateLoginOptionRaw creates a login option from a verbatim JSON body,
// allowing the full authConfig (including nested/type-specific fields) to be
// sent as-is. The parsed response (id, known scalars, Raw) is returned.
func (c *Client) CreateLoginOptionRaw(ctx context.Context, body []byte) (*LoginOption, error) {
	var result LoginOption
	if _, err := c.doJSON(ctx, http.MethodPost, "/tenant/loginOptions", json.RawMessage(body), &result); err != nil {
		return nil, fmt.Errorf("creating login option: %w", err)
	}
	return &result, nil
}

// UpdateLoginOptionRaw updates a login option from a verbatim JSON body.
func (c *Client) UpdateLoginOptionRaw(ctx context.Context, typeOrID string, body []byte) (*LoginOption, error) {
	path := fmt.Sprintf("/tenant/loginOptions/%s", url.PathEscape(typeOrID))
	var result LoginOption
	if _, err := c.doJSON(ctx, http.MethodPut, path, json.RawMessage(body), &result); err != nil {
		return nil, fmt.Errorf("updating login option %q: %w", typeOrID, err)
	}
	return &result, nil
}

// GetLoginOption retrieves a login option by its type or ID.
func (c *Client) GetLoginOption(ctx context.Context, typeOrID string) (*LoginOption, error) {
	path := fmt.Sprintf("/tenant/loginOptions/%s", url.PathEscape(typeOrID))
	var result LoginOption
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting login option %q: %w", typeOrID, err)
	}
	return &result, nil
}

// UpdateLoginOption updates an existing login option.
func (c *Client) UpdateLoginOption(ctx context.Context, typeOrID string, opt LoginOption) (*LoginOption, error) {
	path := fmt.Sprintf("/tenant/loginOptions/%s", url.PathEscape(typeOrID))
	var result LoginOption
	if _, err := c.doJSON(ctx, http.MethodPut, path, opt, &result); err != nil {
		return nil, fmt.Errorf("updating login option %q: %w", typeOrID, err)
	}
	return &result, nil
}

// DeleteLoginOption deletes a login option. Returns nil if already gone.
func (c *Client) DeleteLoginOption(ctx context.Context, typeOrID string) error {
	path := fmt.Sprintf("/tenant/loginOptions/%s", url.PathEscape(typeOrID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting login option %q: %w", typeOrID, err)
	}
	return nil
}
