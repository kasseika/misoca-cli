// user_test.go はユーザー情報APIのクライアントメソッドを検証します。
package misoca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_GetMe は GET /user/me が呼ばれ、レスポンスがデコードされることを検証します。
func TestClient_GetMe(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"email":"shacho@example.co.jp"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))
	user, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe が失敗しました: %v", err)
	}
	if gotPath != "/user/me" {
		t.Errorf("パス = %q, 期待値 = %q", gotPath, "/user/me")
	}
	if user.Email == nil || *user.Email != "shacho@example.co.jp" {
		t.Errorf("デコード結果が期待と異なります: %+v", user)
	}
}
