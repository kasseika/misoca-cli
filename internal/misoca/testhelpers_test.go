// testhelpers_test.go は internal/misoca の各テストファイルで共有する
// 小さなヘルパーです。
package misoca

import (
	"encoding/json"
	"net/http"
)

// decodeJSONBody はリクエストボディをJSONデコードします。
// テストサーバがクライアントの送出したリクエストボディを検証するために使用します。
func decodeJSONBody(r *http.Request, out any) error {
	defer func() { _ = r.Body.Close() }()
	return json.NewDecoder(r.Body).Decode(out)
}
