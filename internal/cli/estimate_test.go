// estimate_test.go は `misoca estimate` コマンド群の振る舞いを検証します。
package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestEstimateList_JSON は一覧取得がJSON出力されることを検証します。
func TestEstimateList_JSON(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/estimates" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"subject":"新規サイト構築費用"}]`))
	})

	code := app.Execute([]string{"estimate", "list"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "新規サイト構築費用") {
		t.Errorf("出力に件名が含まれていません: %s", stdout.String())
	}
}

// TestEstimateDistribute はメール送信コマンドのボディを検証します。
func TestEstimateDistribute(t *testing.T) {
	var gotPath string
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	})

	code := app.Execute([]string{"estimate", "distribute", "1", "--mail-subject", "お見積りのご案内"})
	if code != 0 {
		t.Fatalf("終了コード = %d", code)
	}
	if gotPath != "/estimate/1/distribute" {
		t.Errorf("パス = %q", gotPath)
	}
}
