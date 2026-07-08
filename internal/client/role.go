package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// roleCollectionPage is one page from GET /user/roles.
type roleCollectionPage struct {
	Roles []Role `json:"roles"`
	Next  string `json:"next,omitempty"`
}

// GetRole fetches a single global role by its name (e.g. "ROLE_ALARM_ADMIN").
// Returns ErrNotFound if the role does not exist.
func (c *Client) GetRole(ctx context.Context, name string) (*Role, error) {
	path := fmt.Sprintf("/user/roles/%s", url.PathEscape(name))
	var result Role
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting role %q: %w", name, err)
	}
	return &result, nil
}

// ListRoles fetches all global roles, following pagination automatically.
func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	var all []Role
	path := "/user/roles?pageSize=100"

	for path != "" {
		var page roleCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, fmt.Errorf("listing roles: %w", err)
		}
		all = append(all, page.Roles...)

		// Follow the next-page link if present; strip the base URL prefix.
		if page.Next != "" {
			parsed, err := url.Parse(page.Next)
			if err != nil {
				break
			}
			path = parsed.RequestURI() // e.g. "/user/roles?pageSize=100&currentPage=2"
		} else {
			path = ""
		}
	}
	return all, nil
}
