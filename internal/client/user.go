package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// User represents a Cumulocity user as returned by the API.
// Password is writeOnly in the API and is never included in GET responses.
type User struct {
	ID                             string     `json:"id,omitempty"`
	UserName                       string     `json:"userName,omitempty"`
	Email                          string     `json:"email,omitempty"`
	FirstName                      string     `json:"firstName,omitempty"`
	LastName                       string     `json:"lastName,omitempty"`
	DisplayName                    string     `json:"displayName,omitempty"`
	Phone                          string     `json:"phone,omitempty"`
	Enabled                        *bool      `json:"enabled,omitempty"`
	Newsletter                     *bool      `json:"newsletter,omitempty"`
	Self                           string     `json:"self,omitempty"`
	PasswordStrength               string     `json:"passwordStrength,omitempty"`
	ShouldResetPassword            *bool      `json:"shouldResetPassword,omitempty"`
	LastPasswordChange             string     `json:"lastPasswordChange,omitempty"`
	TwoFactorAuthenticationEnabled *bool      `json:"twoFactorAuthenticationEnabled,omitempty"`
	Roles                          *UserRoles `json:"roles,omitempty"`
}

// UserRoles is the embedded roles object returned inside a User GET response.
type UserRoles struct {
	References []RoleReference `json:"references,omitempty"`
}

// RoleReference is a single entry in the roles list (assignment + role details).
type RoleReference struct {
	Self string `json:"self,omitempty"`
	Role Role   `json:"role"`
}

// Role is a Cumulocity global role, e.g. ROLE_ALARM_ADMIN.
type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// CreateUserRequest is the payload for POST /user/{tenantId}/users.
// At least userName and email are required.
// Either password or sendPasswordResetEmail=true must be provided.
type CreateUserRequest struct {
	UserName               string `json:"userName"`
	Email                  string `json:"email"`
	FirstName              string `json:"firstName,omitempty"`
	LastName               string `json:"lastName,omitempty"`
	DisplayName            string `json:"displayName,omitempty"`
	Phone                  string `json:"phone,omitempty"`
	Enabled                *bool  `json:"enabled,omitempty"`
	Newsletter             *bool  `json:"newsletter,omitempty"`
	Password               string `json:"password,omitempty"`
	SendPasswordResetEmail *bool  `json:"sendPasswordResetEmail,omitempty"`
}

// UpdateUserRequest is the payload for PUT /user/{tenantId}/users/{id}.
// userName is immutable and must not be sent; email/password can only be
// updated for the current user.
type UpdateUserRequest struct {
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	Newsletter  *bool  `json:"newsletter,omitempty"`
}

// CreateUser creates a new user under the given tenant.
func (c *Client) CreateUser(ctx context.Context, tenantID string, req CreateUserRequest) (*User, error) {
	path := fmt.Sprintf("/user/%s/users", url.PathEscape(tenantID))
	var result User
	if _, err := c.doJSON(ctx, http.MethodPost, path, req, &result); err != nil {
		return nil, fmt.Errorf("creating user %q: %w", req.UserName, err)
	}
	return &result, nil
}

// GetUser retrieves a user by ID (which equals userName) within the given tenant.
func (c *Client) GetUser(ctx context.Context, tenantID, userID string) (*User, error) {
	path := fmt.Sprintf("/user/%s/users/%s", url.PathEscape(tenantID), url.PathEscape(userID))
	var result User
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting user %q: %w", userID, err)
	}
	return &result, nil
}

// UpdateUser updates a user's mutable fields.
func (c *Client) UpdateUser(ctx context.Context, tenantID, userID string, req UpdateUserRequest) (*User, error) {
	path := fmt.Sprintf("/user/%s/users/%s", url.PathEscape(tenantID), url.PathEscape(userID))
	var result User
	if _, err := c.doJSON(ctx, http.MethodPut, path, req, &result); err != nil {
		return nil, fmt.Errorf("updating user %q: %w", userID, err)
	}
	return &result, nil
}

// DeleteUser removes a user from the tenant. Returns nil if already deleted (404).
func (c *Client) DeleteUser(ctx context.Context, tenantID, userID string) error {
	path := fmt.Sprintf("/user/%s/users/%s", url.PathEscape(tenantID), url.PathEscape(userID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if err == ErrNotFound {
		return nil // already gone
	}
	if err != nil {
		return fmt.Errorf("deleting user %q: %w", userID, err)
	}
	return nil
}
