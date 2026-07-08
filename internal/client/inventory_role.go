package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// InventoryRole represents a Cumulocity inventory role.
type InventoryRole struct {
	ID          int64                     `json:"id,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Description string                    `json:"description,omitempty"`
	Self        string                    `json:"self,omitempty"`
	Permissions []InventoryRolePermission `json:"permissions,omitempty"`
}

// InventoryRolePermission is one permission entry inside an inventory role.
type InventoryRolePermission struct {
	ID         int64  `json:"id,omitempty"`
	Permission string `json:"permission,omitempty"` // ADMIN | READ | *
	Scope      string `json:"scope,omitempty"`      // ALARM | AUDIT | EVENT | MANAGED_OBJECT | MEASUREMENT | OPERATION | *
	Type       string `json:"type,omitempty"`       // fragment name, e.g. c8y_Restart
}

type inventoryRoleCollectionPage struct {
	Roles []InventoryRole `json:"roles"`
	Next  string          `json:"next,omitempty"`
}

// GetInventoryRole fetches a single inventory role by its numeric ID.
func (c *Client) GetInventoryRole(ctx context.Context, id int64) (*InventoryRole, error) {
	path := fmt.Sprintf("/user/inventoryroles/%d", id)
	var result InventoryRole
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting inventory role %d: %w", id, err)
	}
	return &result, nil
}

// GetInventoryRoleByName lists all inventory roles and returns the first one
// whose name matches exactly. Returns ErrNotFound if no match exists.
func (c *Client) GetInventoryRoleByName(ctx context.Context, name string) (*InventoryRole, error) {
	roles, err := c.ListInventoryRoles(ctx)
	if err != nil {
		return nil, err
	}
	for i, r := range roles {
		if r.Name == name {
			return &roles[i], nil
		}
	}
	return nil, fmt.Errorf("getting inventory role %q: %w", name, ErrNotFound)
}

// ListInventoryRoles fetches all inventory roles, following pagination automatically.
func (c *Client) ListInventoryRoles(ctx context.Context) ([]InventoryRole, error) {
	var all []InventoryRole
	path := "/user/inventoryroles?pageSize=100"

	for path != "" {
		var page inventoryRoleCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing inventory roles: %w", err)
		}
		all = append(all, page.Roles...)

		if page.Next != "" {
			parsed, err := url.Parse(page.Next)
			if err != nil {
				break
			}
			path = parsed.RequestURI()
		} else {
			path = ""
		}
	}
	return all, nil
}
