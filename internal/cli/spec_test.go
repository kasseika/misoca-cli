// spec_test.go は `misoca spec check` コマンド（仕様ドリフト検知）を検証します。
package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtane0412/misoca-cli/api"
)

// TestSpecCheck_NoDrift は、ライブの仕様書が埋め込み済みスナップショットと
// 一致する場合に成功（終了コード0）することを検証します。
func TestSpecCheck_NoDrift(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(api.SwaggerJSON)
	}))
	defer server.Close()

	app := &App{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, SpecCheckURL: server.URL}
	code := app.Execute([]string{"spec", "check"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stderr=%s", code, app.Stderr.(*bytes.Buffer).String())
	}
}

// TestSpecCheck_Drift は、ライブの仕様書がスナップショットと異なる場合に
// 失敗（非ゼロ終了コード）することを検証します。
func TestSpecCheck_Drift(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"swagger":"2.0","definitions":{"NewEntity":{"type":"object"}}}`))
	}))
	defer server.Close()

	app := &App{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, SpecCheckURL: server.URL}
	code := app.Execute([]string{"spec", "check"})
	if code == 0 {
		t.Fatal("差分がある場合は失敗することを期待しましたが成功しました")
	}
}
