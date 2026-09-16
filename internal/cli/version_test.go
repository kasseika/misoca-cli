// version_test.go は `misoca --version` がビルド時に埋め込まれたバージョン文字列を
// 出力することを検証するテストです。
package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestApp_VersionFlag は --version フラグ指定時に Version の値が出力されることを確認します。
func TestApp_VersionFlag(t *testing.T) {
	// 前提条件: Version をテスト用の値に差し替える
	original := Version
	Version = "1.2.3"
	t.Cleanup(func() { Version = original })

	var stdout, stderr bytes.Buffer
	a := &App{Stdout: &stdout, Stderr: &stderr}

	code := a.Execute([]string{"--version"})

	// 検証項目: 終了コードが0で、出力にバージョン番号が含まれること
	if code != 0 {
		t.Fatalf("終了コードが0ではありません: %d (stderr: %s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "1.2.3") {
		t.Fatalf("出力にバージョン番号が含まれていません: %q", stdout.String())
	}
}
