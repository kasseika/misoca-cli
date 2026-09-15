// number.go は Number 型を定義します。
//
// Misoca API v3 は Swagger 仕様書上では number 型と定義されているフィールド
// （total_amount, tax, unit_price など金額・数量系のフィールド）を、実際の
// レスポンスでは文字列（例: "12345.0"）で返すことがあります。Misoca API は
// 「現状有姿・予告なく変更される可能性がある」ものとして提供されているため、
// この乖離を暗黙に無視せず、数値・文字列どちらの表現も受け付けた上で、
// 送信時には常に数値としてエンコードします。
package misoca

import (
	"bytes"
	"fmt"
	"strconv"
)

// Number はJSON上で数値または文字列のどちらでも表現されうる金額・数量です。
type Number float64

// Float64 は Number を float64 として返します。
func (n Number) Float64() float64 { return float64(n) }

// UnmarshalJSON は数値リテラル・文字列リテラル・null のいずれからもデコードします。
func (n *Number) UnmarshalJSON(data []byte) error {
	s := bytes.TrimSpace(data)
	if string(s) == "null" {
		*n = 0
		return nil
	}
	s = bytes.Trim(s, `"`)

	f, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return fmt.Errorf("数値としてデコードできませんでした: %q: %w", data, err)
	}
	*n = Number(f)
	return nil
}

// MarshalJSON は常に数値リテラルとしてエンコードします。
func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(n), 'f', -1, 64)), nil
}
