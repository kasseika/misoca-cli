// invoice_test.go は請求書（Invoice）APIのクライアントメソッドを検証します。
package misoca

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestClient_ListInvoices は請求書一覧の検索クエリパラメータが正しく
// 組み立てられ、Linkヘッダが呼び出し元に伝搬されることを検証します。
func TestClient_ListInvoices(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Link", `<https://app.misoca.jp/api/v3/invoices?page=2>; rel="next"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"subject":"9月分システム開発費"}]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	page := 1
	invoices, linkHeader, err := client.ListInvoices(context.Background(), ListInvoicesParams{
		PaymentStatus: "unpaid",
		InvoiceStatus: "submitted",
		Page:          &page,
	})
	if err != nil {
		t.Fatalf("ListInvoices が失敗しました: %v", err)
	}
	if gotQuery.Get("payment_status") != "unpaid" {
		t.Errorf("payment_status = %q, 期待値 = %q", gotQuery.Get("payment_status"), "unpaid")
	}
	if gotQuery.Get("invoice_status") != "submitted" {
		t.Errorf("invoice_status = %q, 期待値 = %q", gotQuery.Get("invoice_status"), "submitted")
	}
	if gotQuery.Get("page") != "1" {
		t.Errorf("page = %q, 期待値 = %q", gotQuery.Get("page"), "1")
	}
	if len(invoices) != 1 || invoices[0].Subject == nil || *invoices[0].Subject != "9月分システム開発費" {
		t.Errorf("デコード結果が期待と異なります: %+v", invoices)
	}
	if linkHeader == "" {
		t.Error("Linkヘッダが伝搬されていません")
	}
}

// TestClient_GetInvoice は取得リクエストのパスを検証します。
func TestClient_GetInvoice(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":10}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	if _, err := client.GetInvoice(context.Background(), 10); err != nil {
		t.Fatalf("GetInvoice が失敗しました: %v", err)
	}
	if gotPath != "/invoice/10" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/invoice/10")
	}
}

// TestClient_CreateInvoice はPOSTリクエストのボディを検証します。
func TestClient_CreateInvoice(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"subject":"9月分システム開発費"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	created, err := client.CreateInvoice(context.Background(), CreateInvoiceRequest{
		ContactID: 5,
		Subject:   "9月分システム開発費",
	})
	if err != nil {
		t.Fatalf("CreateInvoice が失敗しました: %v", err)
	}
	if gotBody["subject"] != "9月分システム開発費" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
	if created.Subject == nil || *created.Subject != "9月分システム開発費" {
		t.Errorf("レスポンス = %+v", created)
	}
}

// TestClient_GetInvoicePDF は、レスポンスのバイナリボディがそのまま
// 返されることを検証します。
func TestClient_GetInvoicePDF(t *testing.T) {
	pdfBytes := []byte("%PDF-1.4 fake content")
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdfBytes)
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	got, err := client.GetInvoicePDF(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetInvoicePDF が失敗しました: %v", err)
	}
	if gotPath != "/invoice/10/pdf" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/invoice/10/pdf")
	}
	if !bytes.Equal(got, pdfBytes) {
		t.Errorf("PDFバイナリが一致しません: got=%v want=%v", got, pdfBytes)
	}
}

// TestClient_SubmitUnsubmitInvoice は請求ステータスの更新リクエストを検証します。
func TestClient_SubmitUnsubmitInvoice(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	if _, err := client.SubmitInvoice(context.Background(), 1); err != nil {
		t.Fatalf("SubmitInvoice が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/invoice/1/submitted" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}

	if _, err := client.UnsubmitInvoice(context.Background(), 1); err != nil {
		t.Fatalf("UnsubmitInvoice が失敗しました: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/invoice/1/submitted" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
}

// TestClient_PayUnpayInvoice は入金ステータスの更新リクエストを検証します。
// PayInvoiceはリクエストボディに入金日を含みます。
func TestClient_PayUnpayInvoice(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if r.Method == http.MethodPut {
			_ = decodeJSONBody(r, &gotBody)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	if _, err := client.PayInvoice(context.Background(), 1, MarkInvoicePaidRequest{PaidOn: "2026/09/15"}); err != nil {
		t.Fatalf("PayInvoice が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/invoice/1/paid" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
	if gotBody["paid_on"] != "2026/09/15" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}

	if _, err := client.UnpayInvoice(context.Background(), 1); err != nil {
		t.Fatalf("UnpayInvoice が失敗しました: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/invoice/1/paid" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
}

// TestClient_TrashUntrashInvoice はごみ箱への移動・復元のリクエストを検証します。
func TestClient_TrashUntrashInvoice(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	if _, err := client.TrashInvoice(context.Background(), 1); err != nil {
		t.Fatalf("TrashInvoice が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/invoice/1/trashed" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}

	if _, err := client.UntrashInvoice(context.Background(), 1); err != nil {
		t.Fatalf("UntrashInvoice が失敗しました: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/invoice/1/trashed" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
}

// TestClient_SendInvoiceByPostalMail は郵送指示リクエストのパスとメソッドを検証します。
// 課金が発生しうる操作のため、呼び出し確認はCLI層の責務としますが、
// クライアント自体はリクエストの組み立てのみに責務を限定します。
func TestClient_SendInvoiceByPostalMail(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoice":{"id":1}}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	if _, err := client.SendInvoiceByPostalMail(context.Background(), 1); err != nil {
		t.Fatalf("SendInvoiceByPostalMail が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/invoice/1/send_by_postal_mail" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
}
