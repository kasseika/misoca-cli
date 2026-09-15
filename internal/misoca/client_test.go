// client_test.go は HTTP クライアント基盤（Authorizationヘッダ付与・
// エラーレスポンスの変換・Linkヘッダの伝搬・タイムアウト）を検証します。
package misoca

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// staticTokenSource は常に固定のアクセストークンを返すテスト用の TokenSource です。
type staticTokenSource struct {
	token string
	err   error
}

func (s staticTokenSource) Token(_ context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.token, nil
}

// TestClient_SetsAuthorizationHeader は、リクエスト送出時に TokenSource から
// 取得したアクセストークンが Authorization: Bearer ヘッダに設定されることを検証します。
func TestClient_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"email":"yamada@example.co.jp"}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "アクセストークンABC"}, WithBaseURL(server.URL))

	var user User
	_, err := client.doJSON(context.Background(), http.MethodGet, "/user/me", nil, nil, &user)
	if err != nil {
		t.Fatalf("doJSON が失敗しました: %v", err)
	}

	if gotAuth != "Bearer アクセストークンABC" {
		t.Errorf("Authorizationヘッダ = %q, 期待値 = %q", gotAuth, "Bearer アクセストークンABC")
	}
	if user.Email == nil || *user.Email != "yamada@example.co.jp" {
		t.Errorf("レスポンスのデコード結果が期待と異なります: %+v", user)
	}
}

// TestClient_ErrorResponse は、APIがエラーステータスを返した場合に
// ApiEntity_Error（reasons配列）が APIError として返されることを検証します。
func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"reasons":["取引先名を入力してください","郵便番号の形式が不正です"]}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	_, err := client.doJSON(context.Background(), http.MethodPost, "/contact", nil, nil, nil)
	if err == nil {
		t.Fatal("エラーが返されることを期待しましたが nil でした")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("APIError型のエラーを期待しましたが: %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, 期待値 = %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
	wantReasons := []string{"取引先名を入力してください", "郵便番号の形式が不正です"}
	if len(apiErr.Reasons) != len(wantReasons) {
		t.Fatalf("Reasons = %v, 期待値 = %v", apiErr.Reasons, wantReasons)
	}
	for i, r := range wantReasons {
		if apiErr.Reasons[i] != r {
			t.Errorf("Reasons[%d] = %q, 期待値 = %q", i, apiErr.Reasons[i], r)
		}
	}
}

// TestClient_PropagatesLinkHeader は、一覧系レスポンスの Link ヘッダが
// 呼び出し元に伝搬されることを検証します（ページネーション追従のため）。
func TestClient_PropagatesLinkHeader(t *testing.T) {
	wantLink := `<https://app.misoca.jp/api/v3/invoices?page=2>; rel="next"`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", wantLink)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	var invoices []Invoice
	linkHeader, err := client.doJSON(context.Background(), http.MethodGet, "/invoices", nil, nil, &invoices)
	if err != nil {
		t.Fatalf("doJSON が失敗しました: %v", err)
	}
	if linkHeader != wantLink {
		t.Errorf("Linkヘッダ = %q, 期待値 = %q", linkHeader, wantLink)
	}
}

// TestClient_TokenSourceError は、TokenSource がエラーを返した場合に
// HTTPリクエストを送出せずエラーを返すことを検証します。
func TestClient_TokenSourceError(t *testing.T) {
	requested := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{err: errors.New("トークン取得失敗")}, WithBaseURL(server.URL))

	_, err := client.doJSON(context.Background(), http.MethodGet, "/user/me", nil, nil, nil)
	if err == nil {
		t.Fatal("エラーが返されることを期待しましたが nil でした")
	}
	if requested {
		t.Error("TokenSourceがエラーの場合、HTTPリクエストは送出されないべきです")
	}
}

// TestClient_ContextTimeout は、コンテキストのタイムアウトによって
// リクエストが中断されることを検証します。
func TestClient_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(staticTokenSource{token: "t"}, WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.doJSON(ctx, http.MethodGet, "/user/me", nil, nil, nil)
	if err == nil {
		t.Fatal("タイムアウトによるエラーが返されることを期待しましたが nil でした")
	}
}
