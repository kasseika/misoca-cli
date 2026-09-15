// contact_group_test.go は取引先（ContactGroup）APIのクライアントメソッドを検証します。
package misoca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_ListContactGroups は、trashedクエリパラメータが正しく付与され、
// レスポンスのJSON配列がデコードされることを検証します。
func TestClient_ListContactGroups(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"recipient_name":"株式会社サンプル商事"}]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	trashed := true
	groups, err := client.ListContactGroups(context.Background(), ListContactGroupsParams{Trashed: &trashed})
	if err != nil {
		t.Fatalf("ListContactGroups が失敗しました: %v", err)
	}

	if gotPath != "/contact_groups" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/contact_groups")
	}
	if gotQuery != "trashed=true" {
		t.Errorf("クエリ = %q, 期待値 = %q", gotQuery, "trashed=true")
	}
	if len(groups) != 1 || groups[0].RecipientName == nil || *groups[0].RecipientName != "株式会社サンプル商事" {
		t.Errorf("デコード結果が期待と異なります: %+v", groups)
	}
}

// TestClient_GetContactGroup は、指定したIDでの取得リクエストが正しいパスに
// 送出されることを検証します。
func TestClient_GetContactGroup(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"recipient_name":"有限会社テスト工業"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	group, err := client.GetContactGroup(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetContactGroup が失敗しました: %v", err)
	}
	if gotPath != "/contact_group/42" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/contact_group/42")
	}
	if group.ID == nil || *group.ID != 42 {
		t.Errorf("IDのデコード結果が期待と異なります: %+v", group)
	}
}

// TestClient_CreateContactGroup は、POSTリクエストのボディにリクエストが
// JSONエンコードされて送出されることを検証します。
func TestClient_CreateContactGroup(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"recipient_name":"株式会社サンプル商事"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	created, err := client.CreateContactGroup(context.Background(), CreateContactGroupRequest{
		RecipientName: "株式会社サンプル商事",
	})
	if err != nil {
		t.Fatalf("CreateContactGroup が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/contact_group" {
		t.Errorf("リクエスト = %s %s, 期待値 = POST /contact_group", gotMethod, gotPath)
	}
	if gotBody["recipient_name"] != "株式会社サンプル商事" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
	if created.RecipientName == nil || *created.RecipientName != "株式会社サンプル商事" {
		t.Errorf("レスポンスのデコード結果が期待と異なります: %+v", created)
	}
}

// TestClient_TrashUntrashContactGroup は、非表示化・復元がそれぞれ
// PUT/DELETEの /contact_group/{id}/trashed に送出されることを検証します。
func TestClient_TrashUntrashContactGroup(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	if _, err := client.TrashContactGroup(context.Background(), 1); err != nil {
		t.Fatalf("TrashContactGroup が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/contact_group/1/trashed" {
		t.Errorf("リクエスト = %s %s, 期待値 = PUT /contact_group/1/trashed", gotMethod, gotPath)
	}

	if _, err := client.UntrashContactGroup(context.Background(), 1); err != nil {
		t.Fatalf("UntrashContactGroup が失敗しました: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/contact_group/1/trashed" {
		t.Errorf("リクエスト = %s %s, 期待値 = DELETE /contact_group/1/trashed", gotMethod, gotPath)
	}
}
