// invoice.go は請求書（Invoice）に関する Misoca API v3 のクライアントメソッドを実装します。
package misoca

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListInvoicesParams は GET /invoices のクエリパラメータです。
//
// PaymentStatus は "paid" または "unpaid"、InvoiceStatus は "submitted" または
// "unsubmitted"、Type は "active" / "archived" / "trashed" / "untrashed" を
// 指定します。Order は "asc" / "desc"、OrderBy は
// "created_at" / "updated_at" / "issue_date" / "payment_due_on" です。
// 日付系のパラメータはすべて "YYYY/MM/DD" 形式の文字列で指定します。
type ListInvoicesParams struct {
	Type           string
	From           string
	To             string
	DueDateFrom    string
	DueDateTo      string
	UpdatedAtFrom  string
	UpdatedAtTo    string
	PaymentStatus  string
	InvoiceStatus  string
	ContactGroupID *int
	Condition      string
	Order          string
	OrderBy        string
	Page           *int
	PerPage        *int
}

func (p ListInvoicesParams) toQuery() url.Values {
	q := url.Values{}
	setString(q, "type", p.Type)
	setString(q, "from", p.From)
	setString(q, "to", p.To)
	setString(q, "due_date_from", p.DueDateFrom)
	setString(q, "due_date_to", p.DueDateTo)
	setString(q, "updated_at_from", p.UpdatedAtFrom)
	setString(q, "updated_at_to", p.UpdatedAtTo)
	setString(q, "payment_status", p.PaymentStatus)
	setString(q, "invoice_status", p.InvoiceStatus)
	setIntPtr(q, "contact_group_id", p.ContactGroupID)
	setString(q, "condition", p.Condition)
	setString(q, "order", p.Order)
	setString(q, "order_by", p.OrderBy)
	setIntPtr(q, "page", p.Page)
	setIntPtr(q, "per_page", p.PerPage)
	return q
}

// ListInvoices は請求書一覧を取得します。(GET /invoices)
//
// 戻り値の linkHeader は RFC5988 の Link ヘッダです。ParseLinkHeader でパースし、
// 次ページの取得に利用できます。
func (c *Client) ListInvoices(ctx context.Context, params ListInvoicesParams) (invoices []Invoice, linkHeader string, err error) {
	linkHeader, err = c.doJSON(ctx, http.MethodGet, "/invoices", params.toQuery(), nil, &invoices)
	if err != nil {
		return nil, "", err
	}
	return invoices, linkHeader, nil
}

// GetInvoice は請求書を取得します。(GET /invoice/{id})
func (c *Client) GetInvoice(ctx context.Context, id int) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d", id)
	if _, err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// CreateInvoice は請求書を作成します。(POST /invoice)
func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	var created Invoice
	if _, err := c.doJSON(ctx, http.MethodPost, "/invoice", nil, req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetInvoicePDF は請求書のPDFをバイナリのまま取得します。(GET /invoice/{id}/pdf)
func (c *Client) GetInvoicePDF(ctx context.Context, id int) ([]byte, error) {
	path := fmt.Sprintf("/invoice/%d/pdf", id)
	return c.doRaw(ctx, http.MethodGet, path, nil)
}

// SubmitInvoice は請求書を請求済にします。(PUT /invoice/{id}/submitted)
func (c *Client) SubmitInvoice(ctx context.Context, id int) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d/submitted", id)
	if _, err := c.doJSON(ctx, http.MethodPut, path, nil, nil, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// UnsubmitInvoice は請求書を未請求に戻します。(DELETE /invoice/{id}/submitted)
func (c *Client) UnsubmitInvoice(ctx context.Context, id int) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d/submitted", id)
	if _, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// PayInvoice は請求書を入金済にします。(PUT /invoice/{id}/paid)
func (c *Client) PayInvoice(ctx context.Context, id int, req MarkInvoicePaidRequest) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d/paid", id)
	if _, err := c.doJSON(ctx, http.MethodPut, path, nil, req, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// UnpayInvoice は請求書を未入金に戻します。(DELETE /invoice/{id}/paid)
func (c *Client) UnpayInvoice(ctx context.Context, id int) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d/paid", id)
	if _, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// TrashInvoice は請求書をごみ箱に移動します。(PUT /invoice/{id}/trashed)
func (c *Client) TrashInvoice(ctx context.Context, id int) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d/trashed", id)
	if _, err := c.doJSON(ctx, http.MethodPut, path, nil, nil, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// UntrashInvoice はごみ箱の請求書を復元します。(DELETE /invoice/{id}/trashed)
func (c *Client) UntrashInvoice(ctx context.Context, id int) (*Invoice, error) {
	var invoice Invoice
	path := fmt.Sprintf("/invoice/%d/trashed", id)
	if _, err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// SendInvoiceByPostalMail は請求書の郵送を指示します。(POST /invoice/{id}/send_by_postal_mail)
//
// この操作はMisoca側で課金が発生する可能性があるため、呼び出し確認は
// CLI層（internal/cli）の責務とします。
func (c *Client) SendInvoiceByPostalMail(ctx context.Context, id int) (*InvoicePostalMail, error) {
	var result InvoicePostalMail
	path := fmt.Sprintf("/invoice/%d/send_by_postal_mail", id)
	if _, err := c.doJSON(ctx, http.MethodPost, path, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
