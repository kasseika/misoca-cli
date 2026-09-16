// helpers.go は各リソースコマンドで共通して使う小さなユーティリティです。
package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kasseika/misoca-cli/internal/misoca"
	"github.com/kasseika/misoca-cli/internal/output"
)

// parseID はコマンドライン引数のID文字列を整数に変換します。
// 数値でない場合は使い方の誤り（終了コード2）とします。
func parseID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, usageErrorf("IDは数値で指定してください: %q", s)
	}
	return id, nil
}

// nextPage は Link ヘッダの rel="next" URL から page クエリパラメータを取り出します。
// 次ページが無い場合は ok=false を返します。
func nextPage(linkHeader string) (page int, ok bool) {
	links := misoca.ParseLinkHeader(linkHeader)
	next, exists := links["next"]
	if !exists {
		return 0, false
	}
	u, err := url.Parse(next)
	if err != nil {
		return 0, false
	}
	page, err = strconv.Atoi(u.Query().Get("page"))
	if err != nil {
		return 0, false
	}
	return page, true
}

// timeoutContext は --timeout フラグの値でタイムアウト付きコンテキストを生成します。
func timeoutContext(parent context.Context, flags *globalFlags) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, time.Duration(flags.timeout)*time.Second)
}

// parseFormat は --format フラグを検証します。不正な値は使い方の誤り（終了コード2）とします。
func parseFormat(flags *globalFlags) (output.Format, error) {
	format, err := output.ParseFormat(flags.format)
	if err != nil {
		return "", usageErrorf("%v", err)
	}
	return format, nil
}

// readJSONBody は --file フラグで指定されたパス（"-" の場合は標準入力）から
// JSONを読み込み、v にデコードします。
func readJSONBody(stdin io.Reader, path string, v any) error {
	if path == "" {
		return usageErrorf("--file を指定してください（\"-\" で標準入力から読み込みます）")
	}

	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return fmt.Errorf("リクエストボディの読み込みに失敗しました: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("リクエストボディのJSON解析に失敗しました: %w", err)
	}
	return nil
}

// confirm は破壊的・課金の可能性がある操作の前に、標準入力からy/nの確認を取ります。
// autoYes が true の場合はプロンプトを出さずに常に true を返します。
func confirm(stdin io.Reader, stderr io.Writer, autoYes bool, message string) bool {
	if autoYes {
		return true
	}
	_, _ = fmt.Fprintf(stderr, "%s [y/N]: ", message)
	line, _ := bufio.NewReader(stdin).ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// writeSingle は単一オブジェクトをJSON形式で出力します。単一オブジェクトの
// 表形式出力は情報量が少なく有用性が低いため、table/csv指定時はエラーとします。
func writeSingle(stdout io.Writer, format output.Format, v any) error {
	if format != output.FormatJSON {
		return usageErrorf("このコマンドはJSON形式のみに対応しています（--format json）")
	}
	return output.Write(stdout, format, v, nil)
}

// savePDF は取得したバイナリを --output で指定された宛先に書き込みます。
// "-" の場合は標準出力に書き込みます。
func saveBinary(stdout io.Writer, outputPath string, data []byte) error {
	if outputPath == "-" {
		_, err := stdout.Write(data)
		return err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}
	return nil
}
