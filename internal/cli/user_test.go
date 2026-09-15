// user_test.go は `misoca user me` コマンドの振る舞いを検証します。
package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtane0412/misoca-cli/internal/misoca"
)

// newTestApp は httptest サーバを指す偽クライアントを使う、テスト用のAppを生成します。
func newTestApp(t *testing.T, handler http.HandlerFunc) (*App, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app := &App{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  bytes.NewReader(nil),
		NewClient: func(profile string) (*misoca.Client, error) {
			return misoca.NewClient(fakeTokenSource{}, misoca.WithBaseURL(server.URL)), nil
		},
	}
	return app, stdout, stderr
}

// fakeTokenSource はCLIテストで使う固定トークンのTokenSourceです。
type fakeTokenSource struct{}

func (fakeTokenSource) Token(context.Context) (string, error) { return "test-token", nil }

// TestUserMe_JSON は `misoca user me` がJSON形式でユーザー情報を出力することを検証します。
func TestUserMe_JSON(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"email":"shacho@example.co.jp"}`))
	})

	code := app.Execute([]string{"user", "me"})
	if code != 0 {
		t.Fatalf("終了コード = %d, 期待値 = 0, stdout=%s", code, stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("shacho@example.co.jp")) {
		t.Errorf("出力にメールアドレスが含まれていません: %s", stdout.String())
	}
}

// TestUserMe_InvalidFormat は不正な --format 指定が終了コード2で失敗することを検証します。
func TestUserMe_InvalidFormat(t *testing.T) {
	app, _, stderr := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("不正なフォーマットの場合、APIは呼ばれるべきではありません")
	})

	code := app.Execute([]string{"--format", "xml", "user", "me"})
	if code != 2 {
		t.Fatalf("終了コード = %d, 期待値 = 2, stderr=%s", code, stderr.String())
	}
}

// TestUserMe_APIError は、APIがエラーを返した場合に終了コード4になることを検証します。
func TestUserMe_APIError(t *testing.T) {
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"reasons":["トークンが無効です"]}`))
	})

	code := app.Execute([]string{"user", "me"})
	if code != 4 {
		t.Fatalf("終了コード = %d, 期待値 = 4", code)
	}
}
