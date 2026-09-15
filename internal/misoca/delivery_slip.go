// delivery_slip.go は納品書（DeliverySlip）に関する Misoca API v3 の
// クライアントメソッドを実装します。
package misoca

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListDeliverySlipsParams は GET /delivery_slips のクエリパラメータです。
// Type は "active" / "archived" / "trashed" / "untrashed" を指定します。
type ListDeliverySlipsParams struct {
	Type           string
	ContactGroupID *int
	Page           *int
	PerPage        *int
}

func (p ListDeliverySlipsParams) toQuery() url.Values {
	q := url.Values{}
	setString(q, "type", p.Type)
	setIntPtr(q, "contact_group_id", p.ContactGroupID)
	setIntPtr(q, "page", p.Page)
	setIntPtr(q, "per_page", p.PerPage)
	return q
}

// ListDeliverySlips は納品書一覧を取得します。(GET /delivery_slips)
func (c *Client) ListDeliverySlips(ctx context.Context, params ListDeliverySlipsParams) (slips []DeliverySlip, linkHeader string, err error) {
	linkHeader, err = c.doJSON(ctx, http.MethodGet, "/delivery_slips", params.toQuery(), nil, &slips)
	if err != nil {
		return nil, "", err
	}
	return slips, linkHeader, nil
}

// GetDeliverySlip は納品書を取得します。(GET /delivery_slip/{id})
func (c *Client) GetDeliverySlip(ctx context.Context, id int) (*DeliverySlip, error) {
	var slip DeliverySlip
	path := fmt.Sprintf("/delivery_slip/%d", id)
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &slip); err != nil {
		return nil, err
	}
	return &slip, nil
}

// CreateDeliverySlip は納品書を作成します。(POST /delivery_slip)
func (c *Client) CreateDeliverySlip(ctx context.Context, req CreateDeliverySlipRequest) (*DeliverySlip, error) {
	var created DeliverySlip
	if _, err := c.doJSON(ctx, http.MethodPost, "/delivery_slip", nil, req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetDeliverySlipPDF は納品書のPDFを取得します。(GET /delivery_slip/{id}/pdf)
func (c *Client) GetDeliverySlipPDF(ctx context.Context, id int) ([]byte, error) {
	path := fmt.Sprintf("/delivery_slip/%d/pdf", id)
	return c.doRaw(ctx, http.MethodGet, path, nil)
}
