// contact_test.go は送り先（Contact）APIのクライアントメソッドを検証します。
package misoca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_ListContacts は、trashed・contact_group_idの2つのクエリパラメータが
// 正しく付与されることを検証します。
func TestClient_ListContacts(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"recipient_name":"山田太郎"}]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	groupID := 5
	contacts, err := client.ListContacts(context.Background(), ListContactsParams{ContactGroupID: &groupID})
	if err != nil {
		t.Fatalf("ListContacts が失敗しました: %v", err)
	}
	if gotQuery != "contact_group_id=5" {
		t.Errorf("クエリ = %q, 期待値 = %q", gotQuery, "contact_group_id=5")
	}
	if len(contacts) != 1 || contacts[0].RecipientName == nil || *contacts[0].RecipientName != "山田太郎" {
		t.Errorf("デコード結果が期待と異なります: %+v", contacts)
	}
}

// TestClient_GetContact は取得リクエストのパスを検証します。
func TestClient_GetContact(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	if _, err := client.GetContact(context.Background(), 7); err != nil {
		t.Fatalf("GetContact が失敗しました: %v", err)
	}
	if gotPath != "/contact/7" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/contact/7")
	}
}

// TestClient_CreateContact はPOSTリクエストとボディを検証します。
func TestClient_CreateContact(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = decodeJSONBody(r, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"recipient_name":"山田太郎"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	created, err := client.CreateContact(context.Background(), CreateContactRequest{RecipientName: "山田太郎"})
	if err != nil {
		t.Fatalf("CreateContact が失敗しました: %v", err)
	}
	if gotBody["recipient_name"] != "山田太郎" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
	if created.RecipientName == nil || *created.RecipientName != "山田太郎" {
		t.Errorf("レスポンス = %+v", created)
	}
}

// TestClient_TrashUntrashContact は非表示化・復元のリクエストを検証します。
func TestClient_TrashUntrashContact(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	if _, err := client.TrashContact(context.Background(), 1); err != nil {
		t.Fatalf("TrashContact が失敗しました: %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/contact/1/trashed" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}

	if _, err := client.UntrashContact(context.Background(), 1); err != nil {
		t.Fatalf("UntrashContact が失敗しました: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/contact/1/trashed" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
}
