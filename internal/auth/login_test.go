// login_test.go は、ローカルループバックサーバを用いたOAuth2認可コードフロー
// （Login関数）の振る舞いを検証します。
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// capturingBrowser は Login が開こうとした認可URLを記録するテスト用のダブルです。
// 実際にブラウザを起動する代わりに、URLへ疑似的なコールバックリクエストを
// 送出するために使用します。
type capturingBrowser struct {
	mu  sync.Mutex
	url string
}

func (c *capturingBrowser) Open(authURL string) error {
	c.mu.Lock()
	c.url = authURL
	c.mu.Unlock()
	return nil
}

func (c *capturingBrowser) URL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.url
}

// waitForBrowserURL は、Login がループバックサーバの起動後に非同期で
// authURLをOpenBrowserへ渡すまで待機します。
func waitForBrowserURL(t *testing.T, browser *capturingBrowser) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if u := browser.URL(); u != "" {
			return u
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("authURLの取得がタイムアウトしました")
	return ""
}

// TestLogin_Success は、ブラウザでの認可 → コールバック受信 → トークン交換 という
// 一連のフローが成功することを検証します。
func TestLogin_Success(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "発行されたアクセストークン",
			"refresh_token": "発行されたリフレッシュトークン",
			"token_type":    "Bearer",
			"expires_in":    86400,
		})
	}))
	defer tokenServer.Close()

	browser := &capturingBrowser{}
	resultCh := make(chan loginResult, 1)
	go func() {
		token, err := Login(context.Background(), LoginOptions{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Scope:        "write",
			Port:         0, // OSに空きポートを選ばせる（テストの並列実行に安全）
			AuthURL:      "https://example.com/oauth2/authorize",
			TokenURL:     tokenServer.URL,
			OpenBrowser:  browser.Open,
		})
		resultCh <- loginResult{token: token, err: err}
	}()

	authURL := waitForBrowserURL(t, browser)
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("authURLのパースに失敗しました: %v", err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatal("authURLにstateが含まれていません")
	}
	redirectURI := parsed.Query().Get("redirect_uri")
	if redirectURI == "" {
		t.Fatal("authURLにredirect_uriが含まれていません")
	}

	callbackURL := fmt.Sprintf("%s?code=認可コード&state=%s", redirectURI, state)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("コールバックリクエストに失敗しました: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("コールバックのステータス = %d, 期待値 = %d", resp.StatusCode, http.StatusOK)
	}

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("Login が失敗しました: %v", res.err)
		}
		if res.token.AccessToken != "発行されたアクセストークン" {
			t.Errorf("AccessToken = %q, 期待値 = %q", res.token.AccessToken, "発行されたアクセストークン")
		}
		if res.token.RefreshToken != "発行されたリフレッシュトークン" {
			t.Errorf("RefreshToken = %q, 期待値 = %q", res.token.RefreshToken, "発行されたリフレッシュトークン")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Loginの完了待機がタイムアウトしました")
	}
}

// TestLogin_StateMismatch は、コールバックのstateが一致しない場合に
// トークン交換を行わずエラーを返すことを検証します（CSRF対策の検証）。
func TestLogin_StateMismatch(t *testing.T) {
	browser := &capturingBrowser{}
	resultCh := make(chan loginResult, 1)
	go func() {
		token, err := Login(context.Background(), LoginOptions{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Scope:        "write",
			Port:         0,
			AuthURL:      "https://example.com/oauth2/authorize",
			TokenURL:     "https://example.com/oauth2/token",
			OpenBrowser:  browser.Open,
		})
		resultCh <- loginResult{token: token, err: err}
	}()

	authURL := waitForBrowserURL(t, browser)
	parsed, _ := url.Parse(authURL)
	redirectURI := parsed.Query().Get("redirect_uri")

	callbackURL := fmt.Sprintf("%s?code=認可コード&state=不正なstate", redirectURI)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("コールバックリクエストに失敗しました: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("コールバックのステータス = %d, 期待値 = %d", resp.StatusCode, http.StatusBadRequest)
	}

	select {
	case res := <-resultCh:
		if res.err == nil {
			t.Fatal("stateが不一致の場合はエラーが返ることを期待しましたが nil でした")
		}
		if !strings.Contains(res.err.Error(), "state") {
			t.Errorf("エラーメッセージにstateへの言及がありません: %v", res.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Loginの完了待機がタイムアウトしました")
	}
}

// TestLogin_Timeout は、コールバックが一定時間内に届かない場合、
// Login がタイムアウトエラーを返すことを検証します。
func TestLogin_Timeout(t *testing.T) {
	browser := &capturingBrowser{}
	_, err := Login(context.Background(), LoginOptions{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scope:        "write",
		Port:         0,
		AuthURL:      "https://example.com/oauth2/authorize",
		TokenURL:     "https://example.com/oauth2/token",
		OpenBrowser:  browser.Open,
		Timeout:      50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("タイムアウトによるエラーを期待しましたが nil でした")
	}
}

// TestLogin_RequiresClientCredentials は、クライアントIDまたはシークレットが
// 空の場合、ループバックサーバを起動せず即座にエラーを返すことを検証します。
func TestLogin_RequiresClientCredentials(t *testing.T) {
	_, err := Login(context.Background(), LoginOptions{
		ClientID:     "",
		ClientSecret: "",
	})
	if err == nil {
		t.Fatal("クライアント認証情報が空の場合はエラーを期待しましたが nil でした")
	}
}

// loginResult は goroutine 内で実行した Login の結果をテスト本体に
// 送り返すためのヘルパー型です。
type loginResult struct {
	token *oauth2.Token
	err   error
}
