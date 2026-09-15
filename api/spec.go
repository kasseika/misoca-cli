// spec.go は Misoca API v3 の Swagger 仕様書スナップショット（swagger.json）を
// バイナリに埋め込みます。`misoca spec check` が、実行時に取得したライブの
// 仕様書とこのスナップショットを比較し、Misoca側の仕様変更を検知するために使用します。
package api

import _ "embed"

// SwaggerJSON はリポジトリに同梱された Swagger 仕様書スナップショットの内容です。
//
//go:embed swagger.json
var SwaggerJSON []byte
