// user.go はログイン中のユーザー情報に関する Misoca API v3 のクライアントメソッドを実装します。
package misoca

import (
	"context"
	"net/http"
)

// GetMe は認証中のユーザー情報を取得します。(GET /user/me)
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	var user User
	if _, err := c.doJSON(ctx, http.MethodGet, "/user/me", nil, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
