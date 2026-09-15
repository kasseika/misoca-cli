// auth_test.go は `misoca auth` コマンド群の振る舞いを検証します。
package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mtane0412/misoca-cli/internal/auth"
)

// newAuthTestApp は認証情報ファイルを一時ディレクトリに作るAppを生成します。
func newAuthTestApp(t *testing.T) (*App, string) {
	t.Helper()
	dir := t.TempDir()
	credsPath := dir + "/credentials.json"
	app := &App{
		Stdout:          &bytes.Buffer{},
		Stderr:          &bytes.Buffer{},
		Stdin:           strings.NewReader(""),
		CredentialsPath: credsPath,
	}
	app.NewClient = app.defaultNewClient
	return app, credsPath
}

// TestAuthStatus_NotLoggedIn は未ログイン状態で終了コード3になることを検証します。
func TestAuthStatus_NotLoggedIn(t *testing.T) {
	app, _ := newAuthTestApp(t)
	code := app.Execute([]string{"auth", "status"})
	if code != 3 {
		t.Fatalf("終了コード = %d, 期待値 = 3, stderr=%s", code, app.Stderr.(*bytes.Buffer).String())
	}
}

// TestAuthStatus_LoggedIn は、保存済み認証情報がある場合にプロファイル名と
// メールアドレスが出力されることを検証します。
func TestAuthStatus_LoggedIn(t *testing.T) {
	app, credsPath := newAuthTestApp(t)
	store := auth.NewStore(credsPath)
	if err := store.Save("default", &auth.Credentials{
		AccessToken: "t", RefreshToken: "r", Expiry: time.Now().Add(time.Hour), Email: "shacho@example.co.jp",
	}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	code := app.Execute([]string{"auth", "status"})
	stdout := app.Stdout.(*bytes.Buffer).String()
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout)
	}
	if !strings.Contains(stdout, "shacho@example.co.jp") {
		t.Errorf("出力にメールアドレスが含まれていません: %s", stdout)
	}
}

// TestAuthLogout は、ログアウト後に auth status が未ログイン扱いになることを検証します。
func TestAuthLogout(t *testing.T) {
	app, credsPath := newAuthTestApp(t)
	store := auth.NewStore(credsPath)
	if err := store.Save("default", &auth.Credentials{AccessToken: "t"}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	code := app.Execute([]string{"auth", "logout"})
	if code != 0 {
		t.Fatalf("logoutの終了コード = %d", code)
	}

	code = app.Execute([]string{"auth", "status"})
	if code != 3 {
		t.Fatalf("logout後のstatusの終了コード = %d, 期待値 = 3", code)
	}
}

// TestAuthLogin_Success は、ループバックのコールバックを模擬したログインフローが
// 成功し、認証情報が保存されることを検証します。
func TestAuthLogin_Success(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"新しいアクセストークン","refresh_token":"新しいリフレッシュトークン","token_type":"Bearer","expires_in":86400}`))
	}))
	defer tokenServer.Close()
	userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"email":"shacho@example.co.jp"}`))
	}))
	defer userServer.Close()

	app, credsPath := newAuthTestApp(t)
	app.MisocaBaseURL = userServer.URL

	var capturedURL string
	app.OpenBrowser = func(u string) error {
		capturedURL = u
		go func() {
			// authURLからstateとredirect_uriを取り出し、コールバックを模擬する。
			parsed := mustParseURL(t, u)
			state := parsed.Query().Get("state")
			redirectURI := parsed.Query().Get("redirect_uri")
			_, _ = http.Get(redirectURI + "?code=dummy-code&state=" + state)
		}()
		return nil
	}

	code := app.Execute([]string{
		"auth", "login",
		"--client-id", "client-id",
		"--client-secret", "client-secret",
		"--port", "0",
		"--auth-url", "https://example.com/oauth2/authorize",
		"--token-url", tokenServer.URL,
	})
	if code != 0 {
		t.Fatalf("終了コード = %d, stderr=%s", code, app.Stderr.(*bytes.Buffer).String())
	}
	if capturedURL == "" {
		t.Fatal("認可URLがOpenBrowserに渡されていません")
	}

	store := auth.NewStore(credsPath)
	creds, err := store.Load("default")
	if err != nil {
		t.Fatalf("Load が失敗しました: %v", err)
	}
	if creds.AccessToken != "新しいアクセストークン" {
		t.Errorf("AccessToken = %q", creds.AccessToken)
	}
}
