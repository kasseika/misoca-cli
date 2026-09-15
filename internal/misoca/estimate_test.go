// estimate_test.go は見積書（Estimate）APIのクライアントメソッドを検証します。
package misoca

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestClient_ListEstimates はクエリパラメータの組み立てとLinkヘッダの伝搬を検証します。
func TestClient_ListEstimates(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Link", `<https://app.misoca.jp/api/v3/estimates?page=2>; rel="next"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"subject":"新規サイト構築費用"}]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	groupID := 3
	estimates, linkHeader, err := client.ListEstimates(context.Background(), ListEstimatesParams{
		Type:           "active",
		ContactGroupID: &groupID,
	})
	if err != nil {
		t.Fatalf("ListEstimates が失敗しました: %v", err)
	}
	if gotQuery.Get("type") != "active" || gotQuery.Get("contact_group_id") != "3" {
		t.Errorf("クエリ = %v", gotQuery)
	}
	if len(estimates) != 1 || estimates[0].Subject == nil || *estimates[0].Subject != "新規サイト構築費用" {
		t.Errorf("デコード結果が期待と異なります: %+v", estimates)
	}
	if linkHeader == "" {
		t.Error("Linkヘッダが伝搬されていません")
	}
}

// TestClient_GetEstimate は取得リクエストのパスを検証します。
func TestClient_GetEstimate(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":5}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	if _, err := client.GetEstimate(context.Background(), 5); err != nil {
		t.Fatalf("GetEstimate が失敗しました: %v", err)
	}
	if gotPath != "/estimate/5" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/estimate/5")
	}
}

// TestClient_CreateEstimate はPOSTリクエストのボディを検証します。
func TestClient_CreateEstimate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"subject":"新規サイト構築費用"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	created, err := client.CreateEstimate(context.Background(), CreateEstimateRequest{
		ContactID: 3,
		Subject:   "新規サイト構築費用",
	})
	if err != nil {
		t.Fatalf("CreateEstimate が失敗しました: %v", err)
	}
	if gotBody["subject"] != "新規サイト構築費用" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
	if created.Subject == nil || *created.Subject != "新規サイト構築費用" {
		t.Errorf("レスポンス = %+v", created)
	}
}

// TestClient_GetEstimatePdfLogoStamp は3つのバイナリ取得系エンドポイントが
// それぞれ正しいパスに送出されることを検証します。
func TestClient_GetEstimatePdfLogoStamp(t *testing.T) {
	content := []byte("バイナリコンテンツ")
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	pdf, err := client.GetEstimatePDF(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEstimatePDF が失敗しました: %v", err)
	}
	if gotPath != "/estimate/1/pdf" || !bytes.Equal(pdf, content) {
		t.Errorf("PDF取得 パス=%q 内容不一致", gotPath)
	}

	logo, err := client.GetEstimateLogo(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEstimateLogo が失敗しました: %v", err)
	}
	if gotPath != "/estimate/1/logo" || !bytes.Equal(logo, content) {
		t.Errorf("ロゴ取得 パス=%q 内容不一致", gotPath)
	}

	stamp, err := client.GetEstimateStamp(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEstimateStamp が失敗しました: %v", err)
	}
	if gotPath != "/estimate/1/stamp" || !bytes.Equal(stamp, content) {
		t.Errorf("印影取得 パス=%q 内容不一致", gotPath)
	}
}

// TestClient_DistributeEstimate はメール送信リクエストのボディを検証します。
func TestClient_DistributeEstimate(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	_, err := client.DistributeEstimate(context.Background(), 1, DistributeEstimateRequest{
		MailSubject: "お見積りのご案内",
	})
	if err != nil {
		t.Fatalf("DistributeEstimate が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/estimate/1/distribute" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
	if gotBody["mail_subject"] != "お見積りのご案内" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
}
