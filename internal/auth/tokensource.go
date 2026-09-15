// tokensource.go は、Store に保存されたリフレッシュトークンを用いて、
// 期限切れのアクセストークンを自動的に更新する misoca.TokenSource 実装を提供します。
package auth

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/oauth2"
)

// TokenSource は internal/misoca.TokenSource インタフェースを満たす、
// Store 永続化と連動したトークンプロバイダです。
//
// アクセストークンの有効期限が切れている場合、golang.org/x/oauth2 の
// リフレッシュ機構を利用して新しいアクセストークンを取得し、その結果を
// Store に書き戻します。複数ゴルーチンからの同時呼び出しに備え、
// 更新処理全体をミューテックスで保護しています。
type TokenSource struct {
	store   *Store
	profile string
	config  *oauth2.Config

	mu sync.Mutex
}

// NewTokenSource は TokenSource を生成します。
func NewTokenSource(store *Store, profile string, config *oauth2.Config) *TokenSource {
	return &TokenSource{store: store, profile: profile, config: config}
}

// Token は有効なアクセストークンを返します。
func (t *TokenSource) Token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	creds, err := t.store.Load(t.profile)
	if err != nil {
		return "", err
	}

	current := &oauth2.Token{
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		Expiry:       creds.Expiry,
	}

	refreshed, err := t.config.TokenSource(ctx, current).Token()
	if err != nil {
		return "", fmt.Errorf("アクセストークンの更新に失敗しました。再度 `misoca auth login` を実行してください: %w", err)
	}

	if refreshed.AccessToken != current.AccessToken || refreshed.RefreshToken != current.RefreshToken {
		creds.AccessToken = refreshed.AccessToken
		if refreshed.RefreshToken != "" {
			creds.RefreshToken = refreshed.RefreshToken
		}
		creds.Expiry = refreshed.Expiry
		if err := t.store.Save(t.profile, creds); err != nil {
			return "", fmt.Errorf("更新した認証情報の保存に失敗しました: %w", err)
		}
	}

	return refreshed.AccessToken, nil
}
