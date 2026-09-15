// estimate.go は見積書（Estimate）に関する Misoca API v3 のクライアントメソッドを実装します。
package misoca

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListEstimatesParams は GET /estimates のクエリパラメータです。
// Type は "active" / "archived" / "trashed" / "untrashed" を指定します。
type ListEstimatesParams struct {
	Type           string
	ContactGroupID *int
	Page           *int
	PerPage        *int
}

func (p ListEstimatesParams) toQuery() url.Values {
	q := url.Values{}
	setString(q, "type", p.Type)
	setIntPtr(q, "contact_group_id", p.ContactGroupID)
	setIntPtr(q, "page", p.Page)
	setIntPtr(q, "per_page", p.PerPage)
	return q
}

// ListEstimates は見積書一覧を取得します。(GET /estimates)
func (c *Client) ListEstimates(ctx context.Context, params ListEstimatesParams) (estimates []Estimate, linkHeader string, err error) {
	linkHeader, err = c.doJSON(ctx, http.MethodGet, "/estimates", params.toQuery(), nil, &estimates)
	if err != nil {
		return nil, "", err
	}
	return estimates, linkHeader, nil
}

// GetEstimate は見積書を取得します。(GET /estimate/{id})
func (c *Client) GetEstimate(ctx context.Context, id int) (*Estimate, error) {
	var estimate Estimate
	path := fmt.Sprintf("/estimate/%d", id)
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &estimate); err != nil {
		return nil, err
	}
	return &estimate, nil
}

// CreateEstimate は見積書を作成します。(POST /estimate)
func (c *Client) CreateEstimate(ctx context.Context, req CreateEstimateRequest) (*Estimate, error) {
	var created Estimate
	if _, err := c.doJSON(ctx, http.MethodPost, "/estimate", nil, req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetEstimatePDF は見積書のPDFを取得します。(GET /estimate/{id}/pdf)
func (c *Client) GetEstimatePDF(ctx context.Context, id int) ([]byte, error) {
	path := fmt.Sprintf("/estimate/%d/pdf", id)
	return c.doRaw(ctx, http.MethodGet, path, nil)
}

// GetEstimateLogo は見積書に設定されたロゴ画像を取得します。(GET /estimate/{id}/logo)
func (c *Client) GetEstimateLogo(ctx context.Context, id int) ([]byte, error) {
	path := fmt.Sprintf("/estimate/%d/logo", id)
	return c.doRaw(ctx, http.MethodGet, path, nil)
}

// GetEstimateStamp は見積書に設定された印影画像を取得します。(GET /estimate/{id}/stamp)
func (c *Client) GetEstimateStamp(ctx context.Context, id int) ([]byte, error) {
	path := fmt.Sprintf("/estimate/%d/stamp", id)
	return c.doRaw(ctx, http.MethodGet, path, nil)
}

// DistributeEstimate は見積書をメール送信します。(POST /estimate/{id}/distribute)
func (c *Client) DistributeEstimate(ctx context.Context, id int, req DistributeEstimateRequest) (*Distribution, error) {
	var dist Distribution
	path := fmt.Sprintf("/estimate/%d/distribute", id)
	if _, err := c.doJSON(ctx, http.MethodPost, path, nil, req, &dist); err != nil {
		return nil, err
	}
	return &dist, nil
}
