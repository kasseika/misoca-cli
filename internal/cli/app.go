// app.go は misoca CLI のエントリポイントとなる App 構造体を定義します。
//
// App は依存（標準入出力・認証情報の保存先・APIクライアントの生成方法・
// ブラウザ起動処理）を注入可能にすることで、実際のネットワーク通信や
// ファイルシステムを使わずにコマンドの振る舞いをテストできるようにしています。
package cli

import (
	"context"
	"io"
	"os"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/mtane0412/misoca-cli/internal/auth"
	"github.com/mtane0412/misoca-cli/internal/misoca"
)

// App は misoca CLI の実行に必要な依存をまとめた構造体です。
type App struct {
	// Stdout / Stderr はコマンドの標準出力・標準エラー出力です。
	Stdout, Stderr io.Writer
	// Stdin はコマンドの標準入力です（`auth login` の対話入力・`--file -` に使用）。
	Stdin io.Reader
	// CredentialsPath は認証情報ファイルのパスです。空文字列の場合は
	// auth.DefaultCredentialsPath() の既定パスを使用します（テスト用に上書き可能）。
	CredentialsPath string
	// NewClient はプロファイル名からAPIクライアントを生成する関数です。
	// テストではhttptestサーバを指す偽クライアントに差し替えます。
	NewClient func(profile string) (*misoca.Client, error)
	// OpenBrowser は認可URLをブラウザで開く処理です。
	OpenBrowser func(url string) error
	// MisocaBaseURL はAPIクライアントのベースURLです。空文字列の場合は
	// misoca.DefaultBaseURL を使用します（テスト用にhttptestサーバへ差し替え可能）。
	MisocaBaseURL string
	// SpecCheckURL は `misoca spec check` が参照するライブSwagger仕様書のURLです。
	// 空文字列の場合は defaultSpecCheckURL を使用します（テスト用に差し替え可能）。
	SpecCheckURL string
}

// NewApp は標準入出力・実際の認証情報ストア・実際のAPIクライアントを使う
// 既定のAppを生成します。
func NewApp() *App {
	a := &App{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		OpenBrowser: browser.OpenURL,
	}
	a.NewClient = a.defaultNewClient
	return a
}

// credentialsStore は CredentialsPath（未設定なら既定パス）を使う Store を返します。
func (a *App) credentialsStore() (*auth.Store, error) {
	path := a.CredentialsPath
	if path == "" {
		var err error
		path, err = auth.DefaultCredentialsPath()
		if err != nil {
			return nil, err
		}
	}
	return auth.NewStore(path), nil
}

// defaultNewClient は保存済みの認証情報から自動リフレッシュ対応のAPIクライアントを生成します。
func (a *App) defaultNewClient(profile string) (*misoca.Client, error) {
	store, err := a.credentialsStore()
	if err != nil {
		return nil, err
	}
	// 認証情報が無ければここで ErrProfileNotFound を返す（終了コード3）。
	creds, err := store.Load(profile)
	if err != nil {
		return nil, err
	}
	config := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  auth.AuthURL,
			TokenURL: auth.TokenURL,
		},
	}
	tokenSource := auth.NewTokenSource(store, profile, config)
	opts := []misoca.Option{}
	if a.MisocaBaseURL != "" {
		opts = append(opts, misoca.WithBaseURL(a.MisocaBaseURL))
	}
	return misoca.NewClient(tokenSource, opts...), nil
}

// Execute はコマンドライン引数を解釈して実行し、終了コードを返します。
func (a *App) Execute(args []string) int {
	root := a.buildRootCommand()
	root.SetArgs(args)
	root.SetOut(a.Stdout)
	root.SetErr(a.Stderr)

	err := root.ExecuteContext(context.Background())
	if err != nil {
		_, _ = io.WriteString(a.Stderr, "エラー: "+err.Error()+"\n")
	}
	return exitCodeFor(err)
}

// globalFlags はすべてのサブコマンドで共有するグローバルフラグです。
type globalFlags struct {
	format  string
	profile string
	timeout int // 秒単位
}

func (a *App) buildRootCommand() *cobra.Command {
	flags := &globalFlags{}

	root := &cobra.Command{
		Use:           "misoca",
		Short:         "Misoca API v3 を操作するCLI",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&flags.format, "format", "json", "出力形式 (json/table/csv)")
	root.PersistentFlags().StringVar(&flags.profile, "profile", "default", "使用するプロファイル名")
	root.PersistentFlags().IntVar(&flags.timeout, "timeout", 30, "APIリクエストのタイムアウト（秒）")

	root.AddCommand(a.newAuthCmd(flags))
	root.AddCommand(a.newUserCmd(flags))
	root.AddCommand(a.newInvoiceCmd(flags))
	root.AddCommand(a.newEstimateCmd(flags))
	root.AddCommand(a.newDeliverySlipCmd(flags))
	root.AddCommand(a.newContactCmd(flags))
	root.AddCommand(a.newContactGroupCmd(flags))
	root.AddCommand(a.newItemCmd(flags))
	root.AddCommand(a.newSpecCmd(flags))

	return root
}
