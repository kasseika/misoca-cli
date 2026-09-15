// output_test.go は JSON / table / CSV の出力層を検証します。
package output

import (
	"bytes"
	"strings"
	"testing"
)

// TestParseFormat は --format フラグの値検証を確認します。
// 未知の値をJSONへ暗黙にフォールバックせず、エラーにすることが重要です。
func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"table", FormatTable, false},
		{"csv", FormatCSV, false},
		{"xml", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseFormat(%q) はエラーを期待しましたが nil でした", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseFormat(%q) が失敗しました: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %q, 期待値 = %q", tt.input, got, tt.want)
		}
	}
}

// TestWrite_JSON は raw データがそのままインデント付きJSONとして出力されることを検証します。
func TestWrite_JSON(t *testing.T) {
	var buf bytes.Buffer
	type sample struct {
		Name string `json:"name"`
	}
	err := Write(&buf, FormatJSON, sample{Name: "株式会社サンプル商事"}, nil)
	if err != nil {
		t.Fatalf("Write が失敗しました: %v", err)
	}
	want := "{\n  \"name\": \"株式会社サンプル商事\"\n}\n"
	if buf.String() != want {
		t.Errorf("出力 = %q, 期待値 = %q", buf.String(), want)
	}
}

// TestWrite_Table は Table データがタブ区切りの表として整形されることを検証します。
func TestWrite_Table(t *testing.T) {
	var buf bytes.Buffer
	table := &Table{
		Header: []string{"ID", "件名", "金額"},
		Rows: [][]string{
			{"1", "9月分システム開発費", "110,000"},
		},
	}
	if err := Write(&buf, FormatTable, nil, table); err != nil {
		t.Fatalf("Write が失敗しました: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "件名") || !strings.Contains(out, "金額") {
		t.Errorf("ヘッダが出力に含まれていません: %q", out)
	}
	if !strings.Contains(out, "9月分システム開発費") || !strings.Contains(out, "110,000") {
		t.Errorf("データ行が出力に含まれていません: %q", out)
	}
}

// TestWrite_CSV は Table データが正しいCSV形式で出力されることを検証します。
func TestWrite_CSV(t *testing.T) {
	var buf bytes.Buffer
	table := &Table{
		Header: []string{"ID", "件名"},
		Rows: [][]string{
			{"1", "9月分システム開発費"},
			{"2", "カンマ,を含む件名"},
		},
	}
	if err := Write(&buf, FormatCSV, nil, table); err != nil {
		t.Fatalf("Write が失敗しました: %v", err)
	}
	want := "ID,件名\n1,9月分システム開発費\n2,\"カンマ,を含む件名\"\n"
	if buf.String() != want {
		t.Errorf("出力 = %q, 期待値 = %q", buf.String(), want)
	}
}

// TestWrite_TableWithoutTableData は、tableデータを持たないコマンドで
// table/csv形式が指定された場合、暗黙にJSONへフォールバックせずエラーになることを検証します。
func TestWrite_TableWithoutTableData(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, FormatTable, "raw", nil)
	if err == nil {
		t.Fatal("エラーを期待しましたが nil でした")
	}
}
