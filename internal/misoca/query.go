// query.go は各リソースの一覧取得系メソッドで共通して必要となる、
// 未指定（ゼロ値）のクエリパラメータを省略するための小さなヘルパーです。
package misoca

import (
	"net/url"
	"strconv"
)

// setString は value が空文字列でない場合のみクエリパラメータを設定します。
func setString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// setIntPtr は value が nil でない場合のみクエリパラメータを設定します。
func setIntPtr(q url.Values, key string, value *int) {
	if value != nil {
		q.Set(key, strconv.Itoa(*value))
	}
}

// setBoolPtr は value が nil でない場合のみクエリパラメータを設定します。
func setBoolPtr(q url.Values, key string, value *bool) {
	if value != nil {
		q.Set(key, strconv.FormatBool(*value))
	}
}
