package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// RetentionRule represents a Cumulocity retention rule.
type RetentionRule struct {
	ID           string `json:"id,omitempty"`
	DataType     string `json:"dataType,omitempty"`
	FragmentType string `json:"fragmentType,omitempty"`
	MaximumAge   int64  `json:"maximumAge,omitempty"`
	Source       string `json:"source,omitempty"`
	Type         string `json:"type,omitempty"`
	Editable     bool   `json:"editable"`
	Self         string `json:"self,omitempty"`
}

// RetentionRuleRequest is used for create and update payloads.
type RetentionRuleRequest struct {
	DataType     string `json:"dataType,omitempty"`
	FragmentType string `json:"fragmentType,omitempty"`
	MaximumAge   int64  `json:"maximumAge"`
	Source       string `json:"source,omitempty"`
	Type         string `json:"type,omitempty"`
}

// CreateRetentionRule creates a new retention rule.
func (c *Client) CreateRetentionRule(ctx context.Context, req RetentionRuleRequest) (*RetentionRule, error) {
	var result RetentionRule
	if _, err := c.doJSON(ctx, http.MethodPost, "/retention/retentions", req, &result); err != nil {
		return nil, fmt.Errorf("creating retention rule: %w", err)
	}
	return &result, nil
}

// GetRetentionRule retrieves a retention rule by ID.
func (c *Client) GetRetentionRule(ctx context.Context, id string) (*RetentionRule, error) {
	path := fmt.Sprintf("/retention/retentions/%s", url.PathEscape(id))
	var result RetentionRule
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting retention rule %q: %w", id, err)
	}
	return &result, nil
}

// UpdateRetentionRule replaces a retention rule by ID.
func (c *Client) UpdateRetentionRule(ctx context.Context, id string, req RetentionRuleRequest) (*RetentionRule, error) {
	path := fmt.Sprintf("/retention/retentions/%s", url.PathEscape(id))
	var result RetentionRule
	if _, err := c.doJSON(ctx, http.MethodPut, path, req, &result); err != nil {
		return nil, fmt.Errorf("updating retention rule %q: %w", id, err)
	}
	return &result, nil
}

// DeleteRetentionRule removes a retention rule by ID.
func (c *Client) DeleteRetentionRule(ctx context.Context, id string) error {
	path := fmt.Sprintf("/retention/retentions/%s", url.PathEscape(id))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting retention rule %q: %w", id, err)
	}
	return nil
}
