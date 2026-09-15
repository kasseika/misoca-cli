// dealing_item.go は品目（DealingItem）に関する Misoca API v3 のクライアントメソッドを実装します。
package misoca

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListDealingItemsParams は GET /dealing_items のクエリパラメータです。
type ListDealingItemsParams struct {
	Page    *int
	PerPage *int
}

func (p ListDealingItemsParams) toQuery() url.Values {
	q := url.Values{}
	setIntPtr(q, "page", p.Page)
	setIntPtr(q, "per_page", p.PerPage)
	return q
}

// ListDealingItems は品目一覧を取得します。(GET /dealing_items)
//
// 戻り値の linkHeader は RFC5988 の Link ヘッダ（次ページ・最終ページのURL）です。
// ParseLinkHeader でパースし、次ページの取得に利用できます。
func (c *Client) ListDealingItems(ctx context.Context, params ListDealingItemsParams) (items []DealingItem, linkHeader string, err error) {
	linkHeader, err = c.doJSON(ctx, http.MethodGet, "/dealing_items", params.toQuery(), nil, &items)
	if err != nil {
		return nil, "", err
	}
	return items, linkHeader, nil
}

// GetDealingItem は品目を取得します。(GET /dealing_item/{id})
func (c *Client) GetDealingItem(ctx context.Context, id int) (*DealingItem, error) {
	var item DealingItem
	path := fmt.Sprintf("/dealing_item/%d", id)
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// CreateDealingItem は品目を作成します。(POST /dealing_item)
func (c *Client) CreateDealingItem(ctx context.Context, req CreateDealingItemRequest) (*DealingItem, error) {
	var created DealingItem
	if _, err := c.doJSON(ctx, http.MethodPost, "/dealing_item", nil, req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}
