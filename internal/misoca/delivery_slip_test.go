// delivery_slip_test.go は納品書（DeliverySlip）APIのクライアントメソッドを検証します。
package misoca

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestClient_ListDeliverySlips はクエリパラメータの組み立てを検証します。
func TestClient_ListDeliverySlips(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"subject":"部材一式納品"}]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	slips, _, err := client.ListDeliverySlips(context.Background(), ListDeliverySlipsParams{Type: "active"})
	if err != nil {
		t.Fatalf("ListDeliverySlips が失敗しました: %v", err)
	}
	if gotQuery.Get("type") != "active" {
		t.Errorf("type = %q, 期待値 = %q", gotQuery.Get("type"), "active")
	}
	if len(slips) != 1 || slips[0].Subject == nil || *slips[0].Subject != "部材一式納品" {
		t.Errorf("デコード結果が期待と異なります: %+v", slips)
	}
}

// TestClient_GetDeliverySlip は取得リクエストのパスを検証します。
func TestClient_GetDeliverySlip(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":8}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	if _, err := client.GetDeliverySlip(context.Background(), 8); err != nil {
		t.Fatalf("GetDeliverySlip が失敗しました: %v", err)
	}
	if gotPath != "/delivery_slip/8" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/delivery_slip/8")
	}
}

// TestClient_CreateDeliverySlip はPOSTリクエストのボディを検証します。
func TestClient_CreateDeliverySlip(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"subject":"部材一式納品"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	created, err := client.CreateDeliverySlip(context.Background(), CreateDeliverySlipRequest{
		ContactID: 2,
		Subject:   "部材一式納品",
	})
	if err != nil {
		t.Fatalf("CreateDeliverySlip が失敗しました: %v", err)
	}
	if gotBody["subject"] != "部材一式納品" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
	if created.Subject == nil || *created.Subject != "部材一式納品" {
		t.Errorf("レスポンス = %+v", created)
	}
}

// TestClient_GetDeliverySlipPDF はPDFバイナリの取得を検証します。
func TestClient_GetDeliverySlipPDF(t *testing.T) {
	content := []byte("%PDF-1.4 fake")
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	got, err := client.GetDeliverySlipPDF(context.Background(), 8)
	if err != nil {
		t.Fatalf("GetDeliverySlipPDF が失敗しました: %v", err)
	}
	if gotPath != "/delivery_slip/8/pdf" || !bytes.Equal(got, content) {
		t.Errorf("パス=%q 内容不一致", gotPath)
	}
}
