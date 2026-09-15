// contact_test.go は `misoca contact` コマンド群の振る舞いを検証します。
package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestContactList_JSON は送り先一覧がJSON出力されることを検証します。
func TestContactList_JSON(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contacts" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"recipient_name":"山田太郎"}]`))
	})

	code := app.Execute([]string{"contact", "list"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "山田太郎") {
		t.Errorf("出力に名前が含まれていません: %s", stdout.String())
	}
}

// TestContactTrash はごみ箱移動コマンドのリクエストを検証します。
func TestContactTrash(t *testing.T) {
	var gotMethod, gotPath string
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	})

	code := app.Execute([]string{"contact", "trash", "1"})
	if code != 0 {
		t.Fatalf("終了コード = %d", code)
	}
	if gotMethod != http.MethodPut || gotPath != "/contact/1/trashed" {
		t.Errorf("リクエスト = %s %s", gotMethod, gotPath)
	}
}
