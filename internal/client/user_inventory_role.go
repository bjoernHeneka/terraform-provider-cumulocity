package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// InventoryAssignment is the API object returned by the inventory role assignment endpoints.
type InventoryAssignment struct {
	ID            int64           `json:"id,omitempty"`
	ManagedObject string          `json:"managedObject,omitempty"`
	Roles         []InventoryRole `json:"roles,omitempty"`
	Self          string          `json:"self,omitempty"`
}

// InventoryAssignmentCollection is returned by GET .../roles/inventory.
type InventoryAssignmentCollection struct {
	InventoryAssignments []InventoryAssignment `json:"inventoryAssignments,omitempty"`
}

// createInventoryAssignmentRequest is the POST body.
// The API accepts role names on create.
type createInventoryAssignmentRequest struct {
	ManagedObject string                 `json:"managedObject"`
	Roles         []inventoryRoleNameRef `json:"roles"`
}

type inventoryRoleNameRef struct {
	Name string `json:"name"`
}

// updateInventoryAssignmentRequest is the PUT body.
// The API requires role IDs on update.
type updateInventoryAssignmentRequest struct {
	Roles []inventoryRoleIDRef `json:"roles"`
}

type inventoryRoleIDRef struct {
	ID int64 `json:"id"`
}

// CreateUserInventoryRoleAssignment creates a new inventory role assignment for a user on a managed object.
// roleNames are the human-readable names, e.g. "Operations: Restart Device".
func (c *Client) CreateUserInventoryRoleAssignment(ctx context.Context, tenantID, userID, managedObjectID string, roleNames []string) (*InventoryAssignment, error) {
	path := fmt.Sprintf("/user/%s/users/%s/roles/inventory", url.PathEscape(tenantID), url.PathEscape(userID))

	roles := make([]inventoryRoleNameRef, len(roleNames))
	for i, n := range roleNames {
		roles[i] = inventoryRoleNameRef{Name: n}
	}

	body := createInventoryAssignmentRequest{
		ManagedObject: managedObjectID,
		Roles:         roles,
	}

	var result InventoryAssignment
	if _, err := c.doJSON(ctx, http.MethodPost, path, body, &result); err != nil {
		return nil, fmt.Errorf("creating inventory role assignment for user %q on managed object %q: %w", userID, managedObjectID, err)
	}
	return &result, nil
}

// GetUserInventoryRoleAssignment retrieves a specific inventory role assignment by its numeric ID.
func (c *Client) GetUserInventoryRoleAssignment(ctx context.Context, tenantID, userID string, assignmentID int64) (*InventoryAssignment, error) {
	path := fmt.Sprintf("/user/%s/users/%s/roles/inventory/%d", url.PathEscape(tenantID), url.PathEscape(userID), assignmentID)
	var result InventoryAssignment
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting inventory role assignment %d for user %q: %w", assignmentID, userID, err)
	}
	return &result, nil
}

// UpdateUserInventoryRoleAssignment replaces the roles on an existing assignment.
// roleIDs are the numeric IDs of the inventory roles to assign.
func (c *Client) UpdateUserInventoryRoleAssignment(ctx context.Context, tenantID, userID string, assignmentID int64, roleIDs []int64) (*InventoryAssignment, error) {
	path := fmt.Sprintf("/user/%s/users/%s/roles/inventory/%d", url.PathEscape(tenantID), url.PathEscape(userID), assignmentID)

	refs := make([]inventoryRoleIDRef, len(roleIDs))
	for i, id := range roleIDs {
		refs[i] = inventoryRoleIDRef{ID: id}
	}

	var result InventoryAssignment
	if _, err := c.doJSON(ctx, http.MethodPut, path, updateInventoryAssignmentRequest{Roles: refs}, &result); err != nil {
		return nil, fmt.Errorf("updating inventory role assignment %d for user %q: %w", assignmentID, userID, err)
	}
	return &result, nil
}

// DeleteUserInventoryRoleAssignment removes an inventory role assignment. Returns nil if already gone.
func (c *Client) DeleteUserInventoryRoleAssignment(ctx context.Context, tenantID, userID string, assignmentID int64) error {
	path := fmt.Sprintf("/user/%s/users/%s/roles/inventory/%d", url.PathEscape(tenantID), url.PathEscape(userID), assignmentID)
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting inventory role assignment %d for user %q: %w", assignmentID, userID, err)
	}
	return nil
}
