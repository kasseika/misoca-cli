// contact_group.go は取引先（ContactGroup）に関する Misoca API v3 のクライアント
// メソッドを実装します。
package misoca

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListContactGroupsParams は GET /contact_groups のクエリパラメータです。
type ListContactGroupsParams struct {
	// Trashed に true を指定すると非表示にした取引先のみを取得します。
	Trashed *bool
}

func (p ListContactGroupsParams) toQuery() url.Values {
	q := url.Values{}
	setBoolPtr(q, "trashed", p.Trashed)
	return q
}

// ListContactGroups は取引先一覧を取得します。(GET /contact_groups)
func (c *Client) ListContactGroups(ctx context.Context, params ListContactGroupsParams) ([]ContactGroup, error) {
	var groups []ContactGroup
	if _, err := c.doJSON(ctx, http.MethodGet, "/contact_groups", params.toQuery(), nil, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GetContactGroup は取引先を取得します。(GET /contact_group/{id})
func (c *Client) GetContactGroup(ctx context.Context, id int) (*ContactGroup, error) {
	var group ContactGroup
	path := fmt.Sprintf("/contact_group/%d", id)
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// CreateContactGroup は取引先を作成します。(POST /contact_group)
func (c *Client) CreateContactGroup(ctx context.Context, req CreateContactGroupRequest) (*ContactGroup, error) {
	var created ContactGroup
	if _, err := c.doJSON(ctx, http.MethodPost, "/contact_group", nil, req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// TrashContactGroup は取引先を非表示にします。(PUT /contact_group/{id}/trashed)
func (c *Client) TrashContactGroup(ctx context.Context, id int) (*ContactGroup, error) {
	var group ContactGroup
	path := fmt.Sprintf("/contact_group/%d/trashed", id)
	if _, err := c.doJSON(ctx, http.MethodPut, path, nil, nil, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// UntrashContactGroup は非表示にした取引先を表示に戻します。(DELETE /contact_group/{id}/trashed)
func (c *Client) UntrashContactGroup(ctx context.Context, id int) (*ContactGroup, error) {
	var group ContactGroup
	path := fmt.Sprintf("/contact_group/%d/trashed", id)
	if _, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &group); err != nil {
		return nil, err
	}
	return &group, nil
}
