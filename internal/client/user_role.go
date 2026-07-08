package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// assignRoleRequest is the payload for POST /user/{tenantId}/users/{userId}/roles.
// The API requires the role's self-link URL.
type assignRoleRequest struct {
	Role struct {
		Self string `json:"self"`
	} `json:"role"`
}

// AssignUserRole assigns a global role to a user.
// roleID is the role name, e.g. "ROLE_ALARM_ADMIN".
// Returns the roleReference self-link of the new assignment.
func (c *Client) AssignUserRole(ctx context.Context, tenantID, userID, roleID string) (string, error) {
	path := fmt.Sprintf("/user/%s/users/%s/roles", url.PathEscape(tenantID), url.PathEscape(userID))

	var body assignRoleRequest
	body.Role.Self = fmt.Sprintf("%s/user/roles/%s", c.BaseURL, url.PathEscape(roleID))

	var result RoleReference
	if _, err := c.doJSON(ctx, http.MethodPost, path, body, &result); err != nil {
		return "", fmt.Errorf("assigning role %q to user %q: %w", roleID, userID, err)
	}
	return result.Self, nil
}

// HasUserRole checks whether a role is currently assigned to the user.
// It fetches the user object (which embeds all assigned roles) and searches the list.
// Returns ErrNotFound if the user itself does not exist.
func (c *Client) HasUserRole(ctx context.Context, tenantID, userID, roleID string) (bool, error) {
	user, err := c.GetUser(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if user.Roles == nil {
		return false, nil
	}
	for _, ref := range user.Roles.References {
		if ref.Role.ID == roleID {
			return true, nil
		}
	}
	return false, nil
}

// UnassignUserRole removes a role assignment from a user.
// Returns nil if the assignment is already gone (404).
func (c *Client) UnassignUserRole(ctx context.Context, tenantID, userID, roleID string) error {
	path := fmt.Sprintf("/user/%s/users/%s/roles/%s", url.PathEscape(tenantID), url.PathEscape(userID), url.PathEscape(roleID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil // already removed
	}
	if err != nil {
		return fmt.Errorf("unassigning role %q from user %q: %w", roleID, userID, err)
	}
	return nil
}
