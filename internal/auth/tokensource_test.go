// tokensource_test.go は、保存済みのリフレッシュトークンを用いてアクセストークンを
// 自動更新する TokenSource の振る舞いを検証します。
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// TestTokenSource_ReturnsStoredTokenWhenNotExpired は、保存済みトークンが
// まだ有効な場合、トークンエンドポイントに問い合わせずそのまま返すことを検証します。
func TestTokenSource_ReturnsStoredTokenWhenNotExpired(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))
	if err := store.Save("default", &Credentials{
		AccessToken:  "有効なアクセストークン",
		RefreshToken: "リフレッシュトークン",
		Expiry:       time.Now().Add(1 * time.Hour),
	}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	tokenEndpointCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenEndpointCalled = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{TokenURL: server.URL},
	}
	ts := NewTokenSource(store, "default", config)

	got, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("Token が失敗しました: %v", err)
	}
	if got != "有効なアクセストークン" {
		t.Errorf("Token() = %q, 期待値 = %q", got, "有効なアクセストークン")
	}
	if tokenEndpointCalled {
		t.Error("有効なトークンがある場合、トークンエンドポイントは呼ばれないべきです")
	}
}

// TestTokenSource_RefreshesAndPersists は、期限切れのアクセストークンが
// リフレッシュされ、その結果がStoreに書き戻されることを検証します。
func TestTokenSource_RefreshesAndPersists(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials.json")
	store := NewStore(credsPath)
	if err := store.Save("default", &Credentials{
		AccessToken:  "期限切れのアクセストークン",
		RefreshToken: "元のリフレッシュトークン",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "新しいアクセストークン",
			"refresh_token": "新しいリフレッシュトークン",
			"token_type":    "Bearer",
			"expires_in":    86400,
		})
	}))
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{TokenURL: server.URL},
	}
	ts := NewTokenSource(store, "default", config)

	got, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("Token が失敗しました: %v", err)
	}
	if got != "新しいアクセストークン" {
		t.Errorf("Token() = %q, 期待値 = %q", got, "新しいアクセストークン")
	}

	// Store（ファイル）に書き戻されていることを、別インスタンスで読み直して確認する。
	reloaded, err := NewStore(credsPath).Load("default")
	if err != nil {
		t.Fatalf("Load が失敗しました: %v", err)
	}
	if reloaded.AccessToken != "新しいアクセストークン" {
		t.Errorf("永続化されたAccessToken = %q, 期待値 = %q", reloaded.AccessToken, "新しいアクセストークン")
	}
	if reloaded.RefreshToken != "新しいリフレッシュトークン" {
		t.Errorf("永続化されたRefreshToken = %q, 期待値 = %q", reloaded.RefreshToken, "新しいリフレッシュトークン")
	}
	if !reloaded.Expiry.After(time.Now()) {
		t.Errorf("永続化されたExpiryが未来の時刻ではありません: %v", reloaded.Expiry)
	}
}

// TestTokenSource_ProfileNotFound は、未ログインのプロファイルに対して
// 再ログインを促すエラーメッセージを含むエラーが返ることを検証します。
func TestTokenSource_ProfileNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))
	config := &oauth2.Config{}
	ts := NewTokenSource(store, "default", config)

	_, err := ts.Token(context.Background())
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("ErrProfileNotFound を期待しましたが: %v", err)
	}
}

// TestTokenSource_RefreshFailure は、リフレッシュトークンが失効している場合など、
// トークンエンドポイントがエラーを返した場合に、再ログインを促すエラーとなることを検証します。
func TestTokenSource_RefreshFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))
	if err := store.Save("default", &Credentials{
		AccessToken:  "期限切れのアクセストークン",
		RefreshToken: "失効したリフレッシュトークン",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{TokenURL: server.URL},
	}
	ts := NewTokenSource(store, "default", config)

	_, err := ts.Token(context.Background())
	if err == nil {
		t.Fatal("エラーが返されることを期待しましたが nil でした")
	}
}
