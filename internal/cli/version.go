// version.go はビルド時に ldflags 経由で埋め込まれるバージョン文字列を保持します。
package cli

// Version は misoca CLI のバージョン番号です。
// goreleaser のビルド時に `-X github.com/kasseika/misoca-cli/internal/cli.Version={{.Version}}`
// で上書きされます。ローカルビルド（go run/go build）では "dev" のままです。
var Version = "dev"
