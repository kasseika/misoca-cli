// auth.go は `misoca auth` コマンド群（login/status/logout）を実装します。
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mtane0412/misoca-cli/internal/auth"
	"github.com/mtane0412/misoca-cli/internal/misoca"
	"github.com/mtane0412/misoca-cli/internal/output"
)

func (a *App) newAuthCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "認証の管理",
	}
	cmd.AddCommand(a.newAuthLoginCmd(flags))
	cmd.AddCommand(a.newAuthStatusCmd(flags))
	cmd.AddCommand(a.newAuthLogoutCmd(flags))
	return cmd
}

func (a *App) newAuthLoginCmd(flags *globalFlags) *cobra.Command {
	var clientID, clientSecret, scope, authURL, tokenURL string
	var port int

	cmd := &cobra.Command{
		Use:   "login",
		Short: "OAuth2認可コードフローでログインします",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				clientID = os.Getenv("MISOCA_CLIENT_ID")
			}
			if clientSecret == "" {
				clientSecret = os.Getenv("MISOCA_CLIENT_SECRET")
			}
			if clientID == "" || clientSecret == "" {
				clientID, clientSecret = a.promptClientCredentials(clientID, clientSecret)
			}
			if clientID == "" || clientSecret == "" {
				return usageErrorf(
					"クライアントIDとクライアントシークレットが必要です。" +
						"https://app.misoca.jp/oauth2/applications でアプリケーションを登録し、" +
						"コールバックURLに http://localhost:<port>/callback を指定してください。")
			}

			token, err := auth.Login(cmd.Context(), auth.LoginOptions{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Scope:        scope,
				Port:         port,
				AuthURL:      authURL,
				TokenURL:     tokenURL,
				OpenBrowser:  a.OpenBrowser,
			})
			if err != nil {
				return err
			}

			creds := &auth.Credentials{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				Expiry:       token.Expiry,
			}

			// ユーザー情報を取得し、`auth status` 表示用にメールアドレスを記録する。
			// 取得に失敗してもログイン自体は成功とする（致命的ではないため）。
			opts := []misoca.Option{}
			if a.MisocaBaseURL != "" {
				opts = append(opts, misoca.WithBaseURL(a.MisocaBaseURL))
			}
			client := misoca.NewClient(fixedTokenSource(token.AccessToken), opts...)
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			if user, err := client.GetMe(ctx); err == nil && user.Email != nil {
				creds.Email = *user.Email
			}

			store, err := a.credentialsStore()
			if err != nil {
				return err
			}
			if err := store.Save(flags.profile, creds); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(a.Stdout, "ログインしました（プロファイル: %s", flags.profile)
			if creds.Email != "" {
				_, _ = fmt.Fprintf(a.Stdout, ", ユーザー: %s", creds.Email)
			}
			_, _ = fmt.Fprintln(a.Stdout, ")")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "MisocaアプリケーションのクライアントID（環境変数 MISOCA_CLIENT_ID でも指定可）")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Misocaアプリケーションのクライアントシークレット（環境変数 MISOCA_CLIENT_SECRET でも指定可）")
	cmd.Flags().StringVar(&scope, "scope", "write", "要求するスコープ (read/write)")
	cmd.Flags().IntVar(&port, "port", auth.DefaultLoginPort, "ローカルループバックサーバのポート")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "テスト用: 認可エンドポイントの上書き")
	cmd.Flags().StringVar(&tokenURL, "token-url", "", "テスト用: トークンエンドポイントの上書き")
	_ = cmd.Flags().MarkHidden("auth-url")
	_ = cmd.Flags().MarkHidden("token-url")

	return cmd
}

// promptClientCredentials は標準入力からクライアントID・シークレットを対話的に取得します。
func (a *App) promptClientCredentials(clientID, clientSecret string) (string, string) {
	reader := bufio.NewReader(a.Stdin)
	if clientID == "" {
		_, _ = fmt.Fprintln(a.Stderr, "Misocaアプリケーションのクライアント情報が未設定です。")
		_, _ = fmt.Fprintln(a.Stderr, "https://app.misoca.jp/oauth2/applications でアプリケーションを登録してください。")
		_, _ = fmt.Fprint(a.Stderr, "クライアントID: ")
		line, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(line)
	}
	if clientSecret == "" {
		_, _ = fmt.Fprint(a.Stderr, "クライアントシークレット: ")
		line, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(line)
	}
	return clientID, clientSecret
}

func (a *App) newAuthStatusCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "ログイン状態を確認します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			store, err := a.credentialsStore()
			if err != nil {
				return err
			}
			creds, err := store.Load(flags.profile)
			if err != nil {
				return err
			}

			raw := map[string]any{
				"profile": flags.profile,
				"email":   creds.Email,
				"expiry":  creds.Expiry.Format(time.RFC3339),
			}
			table := &output.Table{
				Header: []string{"プロファイル", "ユーザー", "有効期限"},
				Rows:   [][]string{{flags.profile, creds.Email, creds.Expiry.Format(time.RFC3339)}},
			}
			return output.Write(a.Stdout, format, raw, table)
		},
	}
}

func (a *App) newAuthLogoutCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "ログアウトします（保存済み認証情報を削除します）",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := a.credentialsStore()
			if err != nil {
				return err
			}
			if err := store.Delete(flags.profile); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(a.Stdout, "ログアウトしました（プロファイル: %s）\n", flags.profile)
			return nil
		},
	}
}

// fixedTokenSource はログイン直後、認証情報を保存する前の一時的なAPI呼び出し
// （ユーザー情報取得）にのみ使用する固定トークンのTokenSourceです。
type fixedTokenSource string

func (t fixedTokenSource) Token(_ context.Context) (string, error) { return string(t), nil }
