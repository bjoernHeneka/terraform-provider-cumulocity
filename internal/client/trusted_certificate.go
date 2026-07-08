package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// TrustedCertificate represents a trusted certificate stored in Cumulocity.
type TrustedCertificate struct {
	AlgorithmName           string `json:"algorithmName,omitempty"`
	AutoRegistrationEnabled bool   `json:"autoRegistrationEnabled,omitempty"`
	CertInPemFormat         string `json:"certInPemFormat,omitempty"`
	Fingerprint             string `json:"fingerprint,omitempty"`
	Issuer                  string `json:"issuer,omitempty"`
	Name                    string `json:"name,omitempty"`
	NotAfter                string `json:"notAfter,omitempty"`
	NotBefore               string `json:"notBefore,omitempty"`
	Self                    string `json:"self,omitempty"`
	Status                  string `json:"status,omitempty"`
}

// TrustedCertificateUpdateRequest holds the mutable fields for a trusted certificate.
type TrustedCertificateUpdateRequest struct {
	AutoRegistrationEnabled bool   `json:"autoRegistrationEnabled"`
	Name                    string `json:"name,omitempty"`
	Status                  string `json:"status,omitempty"`
}

// AddTrustedCertificate uploads a new trusted certificate for the given tenant.
func (c *Client) AddTrustedCertificate(ctx context.Context, tenantID string, cert TrustedCertificate) (*TrustedCertificate, error) {
	path := fmt.Sprintf("/tenant/tenants/%s/trusted-certificates", url.PathEscape(tenantID))
	var result TrustedCertificate
	if _, err := c.doJSON(ctx, http.MethodPost, path, cert, &result); err != nil {
		return nil, fmt.Errorf("adding trusted certificate: %w", err)
	}
	return &result, nil
}

// GetTrustedCertificate retrieves a trusted certificate by fingerprint.
func (c *Client) GetTrustedCertificate(ctx context.Context, tenantID, fingerprint string) (*TrustedCertificate, error) {
	path := fmt.Sprintf("/tenant/tenants/%s/trusted-certificates/%s", url.PathEscape(tenantID), url.PathEscape(fingerprint))
	var result TrustedCertificate
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting trusted certificate %q: %w", fingerprint, err)
	}
	return &result, nil
}

// UpdateTrustedCertificate updates the mutable fields (name, status, autoRegistrationEnabled).
func (c *Client) UpdateTrustedCertificate(ctx context.Context, tenantID, fingerprint string, req TrustedCertificateUpdateRequest) (*TrustedCertificate, error) {
	path := fmt.Sprintf("/tenant/tenants/%s/trusted-certificates/%s", url.PathEscape(tenantID), url.PathEscape(fingerprint))
	var result TrustedCertificate
	if _, err := c.doJSON(ctx, http.MethodPut, path, req, &result); err != nil {
		return nil, fmt.Errorf("updating trusted certificate %q: %w", fingerprint, err)
	}
	return &result, nil
}

// DeleteTrustedCertificate removes a trusted certificate. Returns nil if already gone.
func (c *Client) DeleteTrustedCertificate(ctx context.Context, tenantID, fingerprint string) error {
	path := fmt.Sprintf("/tenant/tenants/%s/trusted-certificates/%s", url.PathEscape(tenantID), url.PathEscape(fingerprint))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting trusted certificate %q: %w", fingerprint, err)
	}
	return nil
}
