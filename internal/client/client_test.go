package client

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name         string
		tenantDomain string
		username     string
		password     string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid bare hostname",
			tenantDomain: "mytenant.cumulocity.com",
			username:     "admin",
			password:     "secret",
			wantErr:      false,
		},
		{
			name:         "valid host with port",
			tenantDomain: "mytenant.cumulocity.com:8443",
			username:     "admin",
			password:     "secret",
			wantErr:      false,
		},
		{
			name:         "missing tenant_domain",
			tenantDomain: "",
			username:     "admin",
			password:     "secret",
			wantErr:      true,
			errContains:  "tenant_domain is required",
		},
		{
			name:         "missing username",
			tenantDomain: "mytenant.cumulocity.com",
			username:     "",
			password:     "secret",
			wantErr:      true,
			errContains:  "username is required",
		},
		{
			name:         "missing password",
			tenantDomain: "mytenant.cumulocity.com",
			username:     "admin",
			password:     "",
			wantErr:      true,
			errContains:  "password is required",
		},
		{
			name:         "rejects scheme",
			tenantDomain: "https://mytenant.cumulocity.com",
			username:     "admin",
			password:     "secret",
			wantErr:      true,
			errContains:  "must not include a URL scheme",
		},
		{
			name:         "rejects slash",
			tenantDomain: "mytenant.cumulocity.com/path",
			username:     "admin",
			password:     "secret",
			wantErr:      true,
			errContains:  "without slashes or whitespace",
		},
		{
			name:         "rejects backslash",
			tenantDomain: "mytenant.cumulocity.com\\evil",
			username:     "admin",
			password:     "secret",
			wantErr:      true,
			errContains:  "without slashes or whitespace",
		},
		{
			name:         "rejects whitespace",
			tenantDomain: "my tenant.cumulocity.com",
			username:     "admin",
			password:     "secret",
			wantErr:      true,
			errContains:  "without slashes or whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.tenantDomain, "", tt.username, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.BaseURL != "https://"+tt.tenantDomain {
				t.Fatalf("BaseURL = %q, want %q", c.BaseURL, "https://"+tt.tenantDomain)
			}
			if c.userAgent != defaultUserAgent {
				t.Fatalf("userAgent = %q, want %q", c.userAgent, defaultUserAgent)
			}
			if c.httpClient.Timeout != defaultTimeout {
				t.Fatalf("timeout = %v, want %v", c.httpClient.Timeout, defaultTimeout)
			}
		})
	}
}

func TestNewClientBasicAuthEncoding(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		username string
		password string
		wantRaw  string
	}{
		{
			name:     "with tenantID",
			tenantID: "t0071234",
			username: "admin",
			password: "secret",
			wantRaw:  "t0071234/admin:secret",
		},
		{
			name:     "without tenantID",
			tenantID: "",
			username: "admin",
			password: "secret",
			wantRaw:  "admin:secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient("mytenant.cumulocity.com", tt.tenantID, tt.username, tt.password)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			decoded, err := base64.StdEncoding.DecodeString(c.encodedCreds)
			if err != nil {
				t.Fatalf("encodedCreds is not valid base64: %v", err)
			}
			if string(decoded) != tt.wantRaw {
				t.Fatalf("decoded creds = %q, want %q", string(decoded), tt.wantRaw)
			}
		})
	}
}

func TestSetUserAgent(t *testing.T) {
	c, err := NewClient("mytenant.cumulocity.com", "", "admin", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c.SetUserAgent("terraform-provider-cumulocity/1.2.3")
	if c.userAgent != "terraform-provider-cumulocity/1.2.3" {
		t.Fatalf("userAgent = %q, want stamped value", c.userAgent)
	}
	// An empty value must not clobber the existing agent.
	c.SetUserAgent("")
	if c.userAgent != "terraform-provider-cumulocity/1.2.3" {
		t.Fatalf("empty SetUserAgent overwrote userAgent: %q", c.userAgent)
	}
}

func TestTruncateBody(t *testing.T) {
	t.Run("under limit returned verbatim", func(t *testing.T) {
		body := []byte("short error body")
		if got := truncateBody(body); got != "short error body" {
			t.Fatalf("truncateBody = %q, want verbatim", got)
		}
	})

	t.Run("at limit returned verbatim", func(t *testing.T) {
		body := []byte(strings.Repeat("a", maxErrorBodyLen))
		got := truncateBody(body)
		if len(got) != maxErrorBodyLen || strings.Contains(got, "truncated") {
			t.Fatalf("body exactly at limit should not be truncated, len=%d", len(got))
		}
	})

	t.Run("over limit truncated", func(t *testing.T) {
		body := []byte(strings.Repeat("a", maxErrorBodyLen+500))
		got := truncateBody(body)
		if !strings.HasSuffix(got, "... (truncated)") {
			t.Fatalf("expected truncation suffix, got tail %q", got[len(got)-20:])
		}
		if len(got) != maxErrorBodyLen+len("... (truncated)") {
			t.Fatalf("truncated length = %d, want %d", len(got), maxErrorBodyLen+len("... (truncated)"))
		}
	})
}

// newTestClient builds a Client pointed at a test server. NewClient forces an
// https:// prefix, so the struct is constructed directly for HTTP test servers.
func newTestClient(baseURL string) *Client {
	return &Client{
		BaseURL:      baseURL,
		encodedCreds: base64.StdEncoding.EncodeToString([]byte("t007/admin:secret")),
		httpClient:   &http.Client{},
		userAgent:    "terraform-provider-cumulocity/test",
	}
}

func TestDoJSONSuccess(t *testing.T) {
	type payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var gotAuth, gotUA, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"42","name":"device"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	var out payload
	status, err := c.doJSON(context.Background(), http.MethodGet, "/inventory/managedObjects/42", nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if out.ID != "42" || out.Name != "device" {
		t.Fatalf("unmarshalled payload = %+v, want id=42 name=device", out)
	}
	if gotAuth != "Basic "+c.encodedCreds {
		t.Fatalf("Authorization header = %q, want Basic creds", gotAuth)
	}
	if gotUA != "terraform-provider-cumulocity/test" {
		t.Fatalf("User-Agent header = %q, want stamped agent", gotUA)
	}
	if gotAccept != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", gotAccept)
	}
}

func TestDoJSONNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	status, err := c.doJSON(context.Background(), http.MethodGet, "/missing", nil, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
}

func TestDoJSONServerErrorTruncatesBody(t *testing.T) {
	bigBody := strings.Repeat("x", maxErrorBodyLen+1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(bigBody))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	status, err := c.doJSON(context.Background(), http.MethodPost, "/broken", map[string]string{"k": "v"}, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if !strings.Contains(err.Error(), "API error 500") {
		t.Fatalf("error %q missing status prefix", err.Error())
	}
	if !strings.Contains(err.Error(), "... (truncated)") {
		t.Fatalf("error body was not truncated: %q", err.Error())
	}
	if len(err.Error()) >= len(bigBody) {
		t.Fatalf("error length %d suggests full body was echoed", len(err.Error()))
	}
}
