// testhelpers_test.go は internal/cli の各テストファイルで共有する
// ファイル入出力ヘルパーです。
package cli

import (
	"net/url"
	"os"
	"testing"
)

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func readTestFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("URLのパースに失敗しました: %v", err)
	}
	return u
}
