// main.go は misoca CLI のエントリポイントです。
//
// コマンドライン引数の解釈と実行はすべて internal/cli.App に委譲し、
// ここでは戻り値の終了コードで os.Exit するだけに責務を限定しています。
package main

import (
	"os"

	"github.com/kasseika/misoca-cli/internal/cli"
)

func main() {
	app := cli.NewApp()
	os.Exit(app.Execute(os.Args[1:]))
}
