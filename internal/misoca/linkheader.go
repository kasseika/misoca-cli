// linkheader.go は RFC5988（Web Linking）形式の Link ヘッダをパースします。
// Misoca API v3 の一覧系エンドポイントは、ページネーションのための次ページ・
// 最終ページURLをこの形式で返します。
package misoca

import "strings"

// ParseLinkHeader は Link ヘッダの値を rel 名をキーとした URL の map に変換します。
// ヘッダが空文字列の場合は空の map を返します。
func ParseLinkHeader(header string) map[string]string {
	result := make(map[string]string)
	if strings.TrimSpace(header) == "" {
		return result
	}

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		segs := strings.SplitN(part, ";", 2)
		if len(segs) != 2 {
			continue
		}
		url := strings.TrimSpace(segs[0])
		url = strings.TrimPrefix(url, "<")
		url = strings.TrimSuffix(url, ">")

		relPart := strings.TrimSpace(segs[1])
		const relPrefix = `rel="`
		idx := strings.Index(relPart, relPrefix)
		if idx == -1 {
			continue
		}
		relPart = relPart[idx+len(relPrefix):]
		endIdx := strings.Index(relPart, `"`)
		if endIdx == -1 {
			continue
		}
		rel := relPart[:endIdx]

		result[rel] = url
	}
	return result
}
