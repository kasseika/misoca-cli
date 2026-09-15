// linkheader_test.go は RFC5988 の Link ヘッダをパースする ParseLinkHeader の
// 振る舞いを検証します。Misoca API v3 の一覧系エンドポイントは、次ページ・
// 最終ページのURLをこの形式のヘッダで返します。
package misoca

import (
	"reflect"
	"testing"
)

func TestParseLinkHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   map[string]string
	}{
		{
			name:   "空文字列の場合は空のmapを返す",
			header: "",
			want:   map[string]string{},
		},
		{
			name:   "nextとlastの2つのrelを含む場合",
			header: `<https://app.misoca.jp/api/v3/invoices?page=2&per_page=25>; rel="next", <https://app.misoca.jp/api/v3/invoices?page=4&per_page=25>; rel="last"`,
			want: map[string]string{
				"next": "https://app.misoca.jp/api/v3/invoices?page=2&per_page=25",
				"last": "https://app.misoca.jp/api/v3/invoices?page=4&per_page=25",
			},
		},
		{
			name:   "1つのrelのみの場合",
			header: `<https://app.misoca.jp/api/v3/invoices?page=1>; rel="first"`,
			want: map[string]string{
				"first": "https://app.misoca.jp/api/v3/invoices?page=1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLinkHeader(tt.header)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLinkHeader(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}
