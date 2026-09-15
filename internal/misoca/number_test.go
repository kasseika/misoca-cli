// number_test.go は Number 型（数値・文字列どちらのJSON表現も受け付ける
// 金額・数量フィールド用の型）の振る舞いを検証します。
//
// Misoca API v3 は Swagger 仕様書上では number 型と定義されている
// フィールド（total_amount, tax, unit_price 等）を、実際のレスポンスでは
// 文字列（例: "12345.0"）で返すことがあります。本APIは「現状有姿」で
// 提供されており仕様書との乖離がありうるため、両方の表現を受け付ける
// 必要があります。
package misoca

import (
	"encoding/json"
	"testing"
)

// TestNumber_UnmarshalJSON_Number は、JSON上の数値リテラルをデコードできることを検証します。
func TestNumber_UnmarshalJSON_Number(t *testing.T) {
	var n Number
	if err := json.Unmarshal([]byte(`12345.5`), &n); err != nil {
		t.Fatalf("UnmarshalJSON が失敗しました: %v", err)
	}
	if n.Float64() != 12345.5 {
		t.Errorf("Float64() = %v, 期待値 = %v", n.Float64(), 12345.5)
	}
}

// TestNumber_UnmarshalJSON_String は、Misoca APIが実際に返す文字列表現の
// 数値（例: 金額の"12345.0"）をデコードできることを検証します。
func TestNumber_UnmarshalJSON_String(t *testing.T) {
	var n Number
	if err := json.Unmarshal([]byte(`"12345.0"`), &n); err != nil {
		t.Fatalf("UnmarshalJSON が失敗しました: %v", err)
	}
	if n.Float64() != 12345.0 {
		t.Errorf("Float64() = %v, 期待値 = %v", n.Float64(), 12345.0)
	}
}

// TestNumber_UnmarshalJSON_Null は、null が 0 としてデコードされることを検証します。
func TestNumber_UnmarshalJSON_Null(t *testing.T) {
	var n Number
	if err := json.Unmarshal([]byte(`null`), &n); err != nil {
		t.Fatalf("UnmarshalJSON が失敗しました: %v", err)
	}
	if n.Float64() != 0 {
		t.Errorf("Float64() = %v, 期待値 = 0", n.Float64())
	}
}

// TestNumber_UnmarshalJSON_Invalid は、数値として解釈できない文字列が
// エラーとして扱われることを検証します（暗黙のフォールバックをしない）。
func TestNumber_UnmarshalJSON_Invalid(t *testing.T) {
	var n Number
	if err := json.Unmarshal([]byte(`"金額不明"`), &n); err == nil {
		t.Fatal("不正な値に対してエラーが返ることを期待しましたが nil でした")
	}
}

// TestNumber_MarshalJSON は、リクエスト送出時に通常の数値リテラルとして
// エンコードされることを検証します（引用符で囲まれた文字列にはしない）。
func TestNumber_MarshalJSON(t *testing.T) {
	n := Number(500000)
	got, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("MarshalJSON が失敗しました: %v", err)
	}
	if string(got) != "500000" {
		t.Errorf("MarshalJSON結果 = %s, 期待値 = %s", got, "500000")
	}
}

// TestNumber_RoundTripInStruct は、構造体フィールドとして数値・文字列
// どちらのJSONからもデコードでき、エンコード結果は数値になることを検証します。
func TestNumber_RoundTripInStruct(t *testing.T) {
	type sample struct {
		UnitPrice Number `json:"unit_price"`
	}

	var fromNumber sample
	if err := json.Unmarshal([]byte(`{"unit_price": 1000}`), &fromNumber); err != nil {
		t.Fatalf("数値からのデコードに失敗しました: %v", err)
	}

	var fromString sample
	if err := json.Unmarshal([]byte(`{"unit_price": "1000"}`), &fromString); err != nil {
		t.Fatalf("文字列からのデコードに失敗しました: %v", err)
	}

	if fromNumber.UnitPrice != fromString.UnitPrice {
		t.Errorf("数値と文字列でデコード結果が一致しません: %v != %v", fromNumber.UnitPrice, fromString.UnitPrice)
	}

	encoded, err := json.Marshal(fromString)
	if err != nil {
		t.Fatalf("Marshalに失敗しました: %v", err)
	}
	want := `{"unit_price":1000}`
	if string(encoded) != want {
		t.Errorf("エンコード結果 = %s, 期待値 = %s", encoded, want)
	}
}
