package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrNotFound is returned when the API responds with 404.
// Resources use this to remove themselves from state rather than error.
var ErrNotFound = errors.New("resource not found")

const (
	// defaultTimeout bounds every HTTP request so a hung server cannot block a
	// Terraform run indefinitely.
	defaultTimeout = 60 * time.Second
	// maxErrorBodyLen caps how much of an error response body is echoed into an
	// error message. API error bodies can contain arbitrary (potentially
	// sensitive) content, so they are truncated.
	maxErrorBodyLen = 2048
	// defaultUserAgent is used until the provider supplies a version-stamped one.
	defaultUserAgent = "terraform-provider-cumulocity"
)

// Client is a typed HTTP client for the Cumulocity REST API.
//
// Cumulocity Basic auth format:
//
//	Authorization: Basic base64(<tenantID>/<username>:<password>)
//
// The BaseURL is constructed from TenantDomain: https://<TenantDomain>
type Client struct {
	BaseURL      string
	TenantID     string
	Username     string
	Password     string
	httpClient   *http.Client
	encodedCreds string
	userAgent    string
}

// NewClient constructs and validates a Cumulocity API client.
func NewClient(tenantDomain, tenantID, username, password string) (*Client, error) {
	if tenantDomain == "" {
		return nil, fmt.Errorf("tenant_domain is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// tenant_domain must be a bare hostname (optionally host:port). Reject a
	// scheme, path, or whitespace so it cannot be used to redirect requests to
	// an unexpected URL.
	if strings.Contains(tenantDomain, "://") {
		return nil, fmt.Errorf("tenant_domain must not include a URL scheme (got %q); use a bare hostname such as \"mytenant.cumulocity.com\"", tenantDomain)
	}
	if strings.ContainsAny(tenantDomain, "/\\ \t\r\n") {
		return nil, fmt.Errorf("tenant_domain must be a bare hostname without slashes or whitespace (got %q)", tenantDomain)
	}

	baseURL := fmt.Sprintf("https://%s", tenantDomain)

	// Build the Basic auth credential: base64(<tenantID>/<username>:<password>)
	// If tenantID is empty, fall back to username:password (some setups omit tenant prefix).
	var raw string
	if tenantID != "" {
		raw = fmt.Sprintf("%s/%s:%s", tenantID, username, password)
	} else {
		raw = fmt.Sprintf("%s:%s", username, password)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	httpClient := &http.Client{
		Timeout: defaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}

	return &Client{
		BaseURL:      baseURL,
		TenantID:     tenantID,
		Username:     username,
		Password:     password,
		httpClient:   httpClient,
		encodedCreds: encoded,
		userAgent:    defaultUserAgent,
	}, nil
}

// SetUserAgent overrides the User-Agent sent with every request, allowing the
// provider to stamp its version (e.g. "terraform-provider-cumulocity/1.2.3").
func (c *Client) SetUserAgent(ua string) {
	if ua != "" {
		c.userAgent = ua
	}
}

// setAuth applies Cumulocity Basic auth and standard headers to every request.
func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Basic "+c.encodedCreds)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.setUserAgent(req)
}

// setUserAgent sets the User-Agent header. Exposed separately so request
// builders that set their own Content-Type (e.g. multipart uploads) can reuse it.
func (c *Client) setUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
}

// truncateBody bounds the length of an API error body echoed into an error
// message.
func truncateBody(body []byte) string {
	if len(body) > maxErrorBodyLen {
		return string(body[:maxErrorBodyLen]) + "... (truncated)"
	}
	return string(body)
}

// doRequest executes an HTTP request, returns (body, statusCode, error).
func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	c.setAuth(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}
	return body, resp.StatusCode, nil
}

// doJSON is a convenience wrapper: marshals input, calls doRequest, unmarshals into out.
// Pass nil for out to skip unmarshalling (e.g. DELETE responses).
func (c *Client) doJSON(ctx context.Context, method, path string, input, out interface{}) (int, error) {
	var bodyReader io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return 0, fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return status, err
	}

	if status == http.StatusNotFound {
		return status, ErrNotFound
	}
	if status >= 300 {
		return status, fmt.Errorf("API error %d: %s", status, truncateBody(body))
	}

	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return status, fmt.Errorf("unmarshalling response: %w", err)
		}
	}
	return status, nil
}

// CurrentUser represents the authenticated user info returned by /user/currentUser.
type CurrentUser struct {
	ID       string `json:"id"`
	Self     string `json:"self"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
}

// GetCurrentUser retrieves the currently authenticated user.
// Useful for validating credentials during provider configuration.
func (c *Client) GetCurrentUser(ctx context.Context) (*CurrentUser, error) {
	var result CurrentUser
	_, err := c.doJSON(ctx, http.MethodGet, "/user/currentUser", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}
	return &result, nil
}
