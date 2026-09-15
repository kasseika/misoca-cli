// contact.go は送り先（Contact）に関する Misoca API v3 のクライアントメソッドを実装します。
package misoca

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListContactsParams は GET /contacts のクエリパラメータです。
type ListContactsParams struct {
	// Trashed に true を指定すると非表示にした送り先のみを取得します。
	Trashed *bool
	// ContactGroupID を指定すると、特定の取引先に紐付いた送り先のみを取得します。
	ContactGroupID *int
}

func (p ListContactsParams) toQuery() url.Values {
	q := url.Values{}
	setBoolPtr(q, "trashed", p.Trashed)
	setIntPtr(q, "contact_group_id", p.ContactGroupID)
	return q
}

// ListContacts は送り先一覧を取得します。(GET /contacts)
func (c *Client) ListContacts(ctx context.Context, params ListContactsParams) ([]Contact, error) {
	var contacts []Contact
	if _, err := c.doJSON(ctx, http.MethodGet, "/contacts", params.toQuery(), nil, &contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}

// GetContact は送り先を取得します。(GET /contact/{id})
func (c *Client) GetContact(ctx context.Context, id int) (*Contact, error) {
	var contact Contact
	path := fmt.Sprintf("/contact/%d", id)
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &contact); err != nil {
		return nil, err
	}
	return &contact, nil
}

// CreateContact は送り先を作成します。(POST /contact)
func (c *Client) CreateContact(ctx context.Context, req CreateContactRequest) (*Contact, error) {
	var created Contact
	if _, err := c.doJSON(ctx, http.MethodPost, "/contact", nil, req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// TrashContact は送り先を非表示にします。(PUT /contact/{id}/trashed)
func (c *Client) TrashContact(ctx context.Context, id int) (*Contact, error) {
	var contact Contact
	path := fmt.Sprintf("/contact/%d/trashed", id)
	if _, err := c.doJSON(ctx, http.MethodPut, path, nil, nil, &contact); err != nil {
		return nil, err
	}
	return &contact, nil
}

// UntrashContact は非表示にした送り先を表示に戻します。(DELETE /contact/{id}/trashed)
func (c *Client) UntrashContact(ctx context.Context, id int) (*Contact, error) {
	var contact Contact
	path := fmt.Sprintf("/contact/%d/trashed", id)
	if _, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &contact); err != nil {
		return nil, err
	}
	return &contact, nil
}
