// output.go は misoca CLI の出力形式（JSON / table / CSV）を実装します。
//
// 既定の出力形式は常にJSONです。Claude Codeのようなエージェントから
// 安定して機械的に処理できるようにするため、TTYかどうかによる自動判定は
// 行いません。人間が読みやすい表示が必要な場合は明示的に --format table を
// 指定します。
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Format は出力形式です。
type Format string

const (
	// FormatJSON はインデント付きJSONで出力します（既定値）。
	FormatJSON Format = "json"
	// FormatTable はタブ区切りの表として出力します。
	FormatTable Format = "table"
	// FormatCSV はCSV形式で出力します。
	FormatCSV Format = "csv"
)

// ParseFormat は --format フラグの値を検証します。
// json/table/csv 以外の値は、暗黙のフォールバックをせずエラーとします。
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case FormatJSON, FormatTable, FormatCSV:
		return Format(s), nil
	default:
		return "", fmt.Errorf("不明な出力形式です: %q（json/table/csv のいずれかを指定してください）", s)
	}
}

// Table は table/csv 形式で描画するためのヘッダと行データです。
type Table struct {
	Header []string
	Rows   [][]string
}

// Write は raw（JSON出力用の元データ）と table（table/csv出力用のデータ）から、
// format に応じた出力を w に書き込みます。
//
// format が table または csv であるにもかかわらず table が nil の場合、
// そのコマンドが表形式出力に未対応であることを意味するため、暗黙にJSONへ
// フォールバックせずエラーを返します。
func Write(w io.Writer, format Format, raw any, table *Table) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, raw)
	case FormatTable:
		if table == nil {
			return fmt.Errorf("このコマンドは表形式（table）出力に対応していません")
		}
		return writeTable(w, *table)
	case FormatCSV:
		if table == nil {
			return fmt.Errorf("このコマンドはCSV出力に対応していません")
		}
		return writeCSV(w, *table)
	default:
		return fmt.Errorf("不明な出力形式です: %q", format)
	}
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func writeTable(w io.Writer, t Table) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.Join(t.Header, "\t")); err != nil {
		return err
	}
	for _, row := range t.Rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writeCSV(w io.Writer, t Table) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(t.Header); err != nil {
		return err
	}
	for _, row := range t.Rows {
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
