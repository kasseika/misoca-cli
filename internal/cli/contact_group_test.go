// contact_group_test.go は `misoca contact-group` コマンド群の振る舞いを検証します。
package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestContactGroupList_Table は取引先一覧が表形式で出力されることを検証します。
func TestContactGroupList_Table(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contact_groups" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"recipient_name":"株式会社サンプル商事"}]`))
	})

	code := app.Execute([]string{"contact-group", "list", "--format", "table"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "株式会社サンプル商事") {
		t.Errorf("出力に取引先名が含まれていません: %s", stdout.String())
	}
}
