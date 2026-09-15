// client.go は Misoca API v3 に対する HTTP リクエストを実行する薄いクライアントです。
//
// CLI 固有の関心事（出力フォーマット・フラグ解析）を一切持たず、Misoca API との
// 通信のみに責務を限定しています。認証（アクセストークンの発行・自動更新）は
// TokenSource インタフェースの実装（internal/auth）に委譲します。
package misoca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DefaultBaseURL は Misoca API v3 の既定のベースURLです。
const DefaultBaseURL = "https://app.misoca.jp/api/v3"

// TokenSource は HTTP リクエストに付与するアクセストークンを提供します。
// 実装（internal/auth）は、トークンが期限切れの場合に必要に応じて
// リフレッシュを行った上で有効なトークンを返す責務を持ちます。
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// Client は Misoca API v3 の HTTP クライアントです。
type Client struct {
	baseURL    string
	httpClient *http.Client
	tokens     TokenSource
}

// Option は Client の生成時オプションです。
type Option func(*Client)

// WithBaseURL はベースURLを既定値から変更します。主にテストで
// httptest.Server の URL を指定するために使用します。
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient は内部で使用する *http.Client を差し替えます。
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient は Misoca API v3 用のクライアントを生成します。
func NewClient(tokens TokenSource, opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: http.DefaultClient,
		tokens:     tokens,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// doJSON は JSON ボディを受け渡しする API 呼び出しを実行します。
//
// reqBody が非nilの場合は JSON エンコードしてリクエストボディに設定します。
// out が非nilの場合はレスポンスボディを JSON デコードして書き込みます。
// 戻り値の Link ヘッダは、一覧系エンドポイントのページネーション追従に使用します
// （呼び出し元が不要であれば無視して構いません）。
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, reqBody, out any) (linkHeader string, err error) {
	var bodyReader io.Reader
	if reqBody != nil {
		encoded, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("リクエストボディのエンコードに失敗しました: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	resp, respBody, err := c.do(ctx, method, path, query, bodyReader, "application/json")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", newAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return "", fmt.Errorf("レスポンスボディのデコードに失敗しました: %w", err)
		}
	}

	return resp.Header.Get("Link"), nil
}

// doRaw は PDF・画像など JSON 以外のバイナリレスポンスを受け取る API 呼び出しを実行します。
func (c *Client) doRaw(ctx context.Context, method, path string, query url.Values) ([]byte, error) {
	resp, respBody, err := c.do(ctx, method, path, query, nil, "")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, newAPIError(resp.StatusCode, respBody)
	}
	return respBody, nil
}

// do は共通のリクエスト構築・送出・ボディ読み込みを行います。
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType string) (*http.Response, []byte, error) {
	token, err := c.tokens.Token(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("アクセストークンの取得に失敗しました: %w", err)
	}

	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("リクエストの構築に失敗しました: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("リクエストの送出に失敗しました: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("レスポンスボディの読み込みに失敗しました: %w", err)
	}
	// 呼び出し元が resp.Body.Close() を呼べるよう、読み込んだ内容で作り直す。
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	return resp, respBody, nil
}

// newAPIError はエラーレスポンスのボディ（ApiEntity_Error）を APIError に変換します。
// ボディが期待する形式でない場合も、可能な限り情報を保持したエラーを返します。
func newAPIError(statusCode int, body []byte) *APIError {
	var parsed struct {
		Reasons []string `json:"reasons"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return &APIError{StatusCode: statusCode, Reasons: []string{string(body)}}
	}
	return &APIError{StatusCode: statusCode, Reasons: parsed.Reasons}
}
