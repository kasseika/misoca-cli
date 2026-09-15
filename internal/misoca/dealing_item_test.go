// dealing_item_test.go は品目（DealingItem）APIのクライアントメソッドを検証します。
package misoca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_ListDealingItems は page/per_page クエリパラメータの付与を検証します。
func TestClient_ListDealingItems(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"name":"システム開発費"}]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	page, perPage := 2, 50
	items, _, err := client.ListDealingItems(context.Background(), ListDealingItemsParams{Page: &page, PerPage: &perPage})
	if err != nil {
		t.Fatalf("ListDealingItems が失敗しました: %v", err)
	}
	if gotQuery != "page=2&per_page=50" {
		t.Errorf("クエリ = %q, 期待値 = %q", gotQuery, "page=2&per_page=50")
	}
	if len(items) != 1 || items[0].Name == nil || *items[0].Name != "システム開発費" {
		t.Errorf("デコード結果が期待と異なります: %+v", items)
	}
}

// TestClient_GetDealingItem は取得リクエストのパスを検証します。
func TestClient_GetDealingItem(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	if _, err := client.GetDealingItem(context.Background(), 3); err != nil {
		t.Fatalf("GetDealingItem が失敗しました: %v", err)
	}
	if gotPath != "/dealing_item/3" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/dealing_item/3")
	}
}

// TestClient_CreateDealingItem はPOSTリクエストとボディを検証します。
func TestClient_CreateDealingItem(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"name":"システム開発費"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	created, err := client.CreateDealingItem(context.Background(), CreateDealingItemRequest{Name: "システム開発費"})
	if err != nil {
		t.Fatalf("CreateDealingItem が失敗しました: %v", err)
	}
	if gotBody["name"] != "システム開発費" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
	if created.Name == nil || *created.Name != "システム開発費" {
		t.Errorf("レスポンス = %+v", created)
	}
}
