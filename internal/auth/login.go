// login.go は、ローカルループバックサーバを用いたOAuth2認可コードフローを実装します。
//
// Misoca は Web アプリケーション用の OAuth2 のみをサポートしており、CLI のような
// ネイティブアプリケーション向けの特別なフローは提供していません。そこで、
// 一時的に 127.0.0.1 上でコールバック受信用のHTTPサーバを起動し、ブラウザでの
// 認可完了後にそのサーバへリダイレクトさせることで、CLI 単体で認可コードを
// 受け取れるようにしています。
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

const (
	// AuthURL は Misoca の OAuth2 認可エンドポイントです。
	AuthURL = "https://app.misoca.jp/oauth2/authorize"
	// TokenURL は Misoca の OAuth2 トークンエンドポイントです。
	TokenURL = "https://app.misoca.jp/oauth2/token"

	// DefaultLoginPort はループバックサーバの既定のポート番号です。
	// misoca auth login 実行前に、このポートをコールバックURLとして
	// Misocaのアプリケーション管理ページに登録しておく必要があります。
	DefaultLoginPort = 8765

	defaultLoginTimeout = 5 * time.Minute
)

// LoginOptions は Login の実行パラメータです。
type LoginOptions struct {
	// ClientID / ClientSecret は Misoca アプリケーション管理ページで発行された値です。
	ClientID     string
	ClientSecret string
	// Scope は要求するOAuth2スコープ（"read" または "write"）です。
	Scope string
	// Port はループバックサーバのポート番号です。0を指定するとOSが空きポートを選択します
	// （テスト用途）。実運用では、事前にMisocaへ登録したコールバックURLのポートと
	// 一致させる必要があるため、既定値は DefaultLoginPort です。
	Port int
	// AuthURL / TokenURL はテスト時に差し替え可能なエンドポイントです。
	// 空文字列の場合はMisocaの実エンドポイント（AuthURL / TokenURL定数）を使用します。
	AuthURL  string
	TokenURL string
	// OpenBrowser は認可URLを開く処理です。nilの場合はブラウザを開きません
	// （呼び出し元が標準エラー出力のURLを見て手動で開くことを想定します）。
	OpenBrowser func(authURL string) error
	// Timeout はコールバック待機の最大時間です。0の場合は既定値（5分）を使用します。
	Timeout time.Duration
}

// loginCallbackResult はコールバックハンドラからLogin本体へ結果を伝える内部型です。
type loginCallbackResult struct {
	code string
	err  error
}

// Login はループバックサーバを用いたOAuth2認可コードフローを実行し、
// 取得したトークンを返します。
func Login(ctx context.Context, opts LoginOptions) (*oauth2.Token, error) {
	if opts.ClientID == "" || opts.ClientSecret == "" {
		return nil, fmt.Errorf("クライアントIDとクライアントシークレットが必要です。" +
			"https://app.misoca.jp/oauth2/applications でアプリケーションを登録してください")
	}

	authURL := opts.AuthURL
	if authURL == "" {
		authURL = AuthURL
	}
	tokenURL := opts.TokenURL
	if tokenURL == "" {
		tokenURL = TokenURL
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = defaultLoginTimeout
	}

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("stateの生成に失敗しました: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.Port))
	if err != nil {
		return nil, fmt.Errorf(
			"ローカルサーバの起動に失敗しました（ポート%dが使用中の可能性があります）: %w", opts.Port, err)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port

	config := &oauth2.Config{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		Scopes:       []string{opts.Scope},
		RedirectURL:  fmt.Sprintf("http://localhost:%d/callback", actualPort),
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	resultCh := make(chan loginCallbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, state, resultCh)
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close() }()

	authCodeURL := config.AuthCodeURL(state)
	// ブラウザが自動で開けない環境でも認証できるよう、URLを必ず標準エラー出力に表示する。
	fmt.Fprintf(os.Stderr, "ブラウザで以下のURLを開いて認証してください:\n%s\n", authCodeURL)
	if opts.OpenBrowser != nil {
		_ = opts.OpenBrowser(authCodeURL)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case res := <-resultCh:
		if res.err != nil {
			return nil, res.err
		}
		token, err := config.Exchange(timeoutCtx, res.code)
		if err != nil {
			return nil, fmt.Errorf("トークンの取得に失敗しました: %w", err)
		}
		return token, nil
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("認証がタイムアウトしました。もう一度 `misoca auth login` を実行してください: %w", timeoutCtx.Err())
	}
}

// handleCallback はループバックサーバのコールバックエンドポイントの処理本体です。
func handleCallback(w http.ResponseWriter, r *http.Request, expectedState string, resultCh chan<- loginCallbackResult) {
	q := r.URL.Query()

	if errParam := q.Get("error"); errParam != "" {
		http.Error(w, fmt.Sprintf("認可が拒否されました: %s", errParam), http.StatusBadRequest)
		resultCh <- loginCallbackResult{err: fmt.Errorf("認可が拒否されました: %s", errParam)}
		return
	}
	if q.Get("state") != expectedState {
		http.Error(w, "stateが一致しません。認証を中断しました。", http.StatusBadRequest)
		resultCh <- loginCallbackResult{err: fmt.Errorf("コールバックのstateが一致しません（CSRF対策により中断しました）")}
		return
	}

	code := q.Get("code")
	if code == "" {
		http.Error(w, "認可コードが含まれていません。", http.StatusBadRequest)
		resultCh <- loginCallbackResult{err: fmt.Errorf("コールバックに認可コードが含まれていません")}
		return
	}

	_, _ = fmt.Fprintln(w, "認証が完了しました。このタブを閉じて構いません。")
	resultCh <- loginCallbackResult{code: code}
}

// randomState はCSRF対策用のランダムなstate値を生成します。
func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
