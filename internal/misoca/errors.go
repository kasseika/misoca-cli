// errors.go は Misoca API v3 が返すエラーレスポンス（ApiEntity_Error）を
// Go のエラー型として扱うための定義です。
package misoca

import "fmt"

// APIError は Misoca API がエラーステータスを返した際のエラーです。
// レスポンスボディの ApiEntity_Error（{"reasons": ["..."]}）をそのまま保持します。
type APIError struct {
	// StatusCode は HTTP レスポンスのステータスコードです。
	StatusCode int
	// Reasons は API が返したエラー理由の一覧です。
	Reasons []string
}

// Error は error インタフェースを満たします。
func (e *APIError) Error() string {
	if len(e.Reasons) == 0 {
		return fmt.Sprintf("misoca api error (status=%d)", e.StatusCode)
	}
	return fmt.Sprintf("misoca api error (status=%d): %s", e.StatusCode, joinReasons(e.Reasons))
}

func joinReasons(reasons []string) string {
	out := reasons[0]
	for _, r := range reasons[1:] {
		out += "; " + r
	}
	return out
}
