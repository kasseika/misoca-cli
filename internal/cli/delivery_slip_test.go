// delivery_slip_test.go は `misoca delivery-slip` コマンド群の振る舞いを検証します。
package cli

import (
	"net/http"
	"strings"
	"testing"
)

// TestDeliverySlipList_JSON は一覧取得がJSON出力されることを検証します。
func TestDeliverySlipList_JSON(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/delivery_slips" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"subject":"部材一式納品"}]`))
	})

	code := app.Execute([]string{"delivery-slip", "list"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "部材一式納品") {
		t.Errorf("出力に件名が含まれていません: %s", stdout.String())
	}
}
