// item_test.go は `misoca item` コマンド群の振る舞いを検証します。
package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestItemList_JSON は品目一覧がJSON出力されることを検証します。
func TestItemList_JSON(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dealing_items" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"name":"システム開発費"}]`))
	})

	code := app.Execute([]string{"item", "list"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "システム開発費") {
		t.Errorf("出力に品目名が含まれていません: %s", stdout.String())
	}
}
