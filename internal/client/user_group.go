package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Group represents a Cumulocity user group.
type Group struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

type groupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type groupCollectionPage struct {
	Groups []Group `json:"groups"`
	Next   string  `json:"next,omitempty"`
}

// CreateGroup creates a new user group for the given tenant.
func (c *Client) CreateGroup(ctx context.Context, tenantID, name, description string) (*Group, error) {
	path := fmt.Sprintf("/user/%s/groups", url.PathEscape(tenantID))
	body := groupRequest{Name: name, Description: description}
	var result Group
	if _, err := c.doJSON(ctx, http.MethodPost, path, body, &result); err != nil {
		return nil, fmt.Errorf("creating group %q: %w", name, err)
	}
	return &result, nil
}

// GetGroup retrieves a user group by its numeric ID.
func (c *Client) GetGroup(ctx context.Context, tenantID string, groupID int64) (*Group, error) {
	path := fmt.Sprintf("/user/%s/groups/%d", url.PathEscape(tenantID), groupID)
	var result Group
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("getting group %d: %w", groupID, err)
	}
	return &result, nil
}

// UpdateGroup updates the name and description of an existing user group.
func (c *Client) UpdateGroup(ctx context.Context, tenantID string, groupID int64, name, description string) (*Group, error) {
	path := fmt.Sprintf("/user/%s/groups/%d", url.PathEscape(tenantID), groupID)
	body := groupRequest{Name: name, Description: description}
	var result Group
	if _, err := c.doJSON(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, fmt.Errorf("updating group %d: %w", groupID, err)
	}
	return &result, nil
}

// DeleteGroup deletes a user group. Returns nil if the group is already gone.
func (c *Client) DeleteGroup(ctx context.Context, tenantID string, groupID int64) error {
	path := fmt.Sprintf("/user/%s/groups/%d", url.PathEscape(tenantID), groupID)
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting group %d: %w", groupID, err)
	}
	return nil
}

// userReferenceItem is used when listing members of a group.
type userReferenceItem struct {
	User struct {
		UserName string `json:"userName"`
	} `json:"user"`
}

type userReferenceCollectionPage struct {
	References []userReferenceItem `json:"references"`
	Next       string              `json:"next,omitempty"`
}

// AddUserToGroup adds a user to a group. The user is identified by their userName.
// The body must contain the user's self-link URL per the Cumulocity API contract.
func (c *Client) AddUserToGroup(ctx context.Context, tenantID string, groupID int64, userID string) error {
	path := fmt.Sprintf("/user/%s/groups/%d/users", url.PathEscape(tenantID), groupID)

	type userRef struct {
		Self string `json:"self"`
	}
	type subscribedUser struct {
		User userRef `json:"user"`
	}

	body := subscribedUser{
		User: userRef{
			Self: fmt.Sprintf("%s/user/%s/users/%s", c.BaseURL, url.PathEscape(tenantID), url.PathEscape(userID)),
		},
	}
	if _, err := c.doJSON(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("adding user %q to group %d: %w", userID, groupID, err)
	}
	return nil
}

// HasUserInGroup returns true if the user is currently a member of the group.
// Returns ErrNotFound if the group itself does not exist.
func (c *Client) HasUserInGroup(ctx context.Context, tenantID string, groupID int64, userID string) (bool, error) {
	path := fmt.Sprintf("/user/%s/groups/%d/users?pageSize=100", url.PathEscape(tenantID), groupID)
	for path != "" {
		var page userReferenceCollectionPage
		if _, err := c.doJSON(ctx, http.MethodGet, path, nil, &page); err != nil {
			return false, err // propagates ErrNotFound when group is gone
		}
		for _, ref := range page.References {
			if ref.User.UserName == userID {
				return true, nil
			}
		}
		path = ""
		if page.Next != "" {
			u, err := url.Parse(page.Next)
			if err == nil {
				path = u.RequestURI()
			}
		}
	}
	return false, nil
}

// RemoveUserFromGroup removes a user from a group. Returns nil if the user is not a member.
func (c *Client) RemoveUserFromGroup(ctx context.Context, tenantID string, groupID int64, userID string) error {
	path := fmt.Sprintf("/user/%s/groups/%d/users/%s", url.PathEscape(tenantID), groupID, url.PathEscape(userID))
	_, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("removing user %q from group %d: %w", userID, groupID, err)
	}
	return nil
}
