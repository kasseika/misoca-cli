// store.go は OAuth2 認証情報をローカルファイルに永続化する Store を提供します。
//
// 複数の Misoca アカウント（個人・法人など）を切り替えられるよう、
// プロファイル名をキーとした JSON ファイルとして保存します。
// ファイルには機密情報（アクセストークン・リフレッシュトークン・
// クライアントシークレット）を含むため、パーミッションを 0600（ディレクトリは
// 0700）に制限し、他ユーザーから読み取れないようにします。
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials は1プロファイル分のOAuth2認証情報です。
type Credentials struct {
	// ClientID は Misoca アプリケーション管理ページで発行されるアプリケーションIDです。
	ClientID string `json:"client_id"`
	// ClientSecret は同ページで発行されるシークレットです。
	ClientSecret string `json:"client_secret"`
	// AccessToken は現在有効な（または直近に取得した）アクセストークンです。
	AccessToken string `json:"access_token"`
	// RefreshToken はアクセストークンの再発行に使用するリフレッシュトークンです。
	RefreshToken string `json:"refresh_token"`
	// Expiry はアクセストークンの有効期限です。
	Expiry time.Time `json:"expiry"`
	// Email はログイン中のMisocaユーザーのメールアドレスです（`misoca auth status`表示用、任意）。
	Email string `json:"email,omitempty"`
}

// ErrProfileNotFound は指定したプロファイルの認証情報が保存されていない場合のエラーです。
var ErrProfileNotFound = errors.New("認証情報が見つかりません。`misoca auth login` を実行してください")

// Store はプロファイル名をキーとした認証情報をファイルに永続化します。
type Store struct {
	path string
}

// NewStore は path のファイルを使用する Store を生成します。
// ファイルが存在しない場合は、初回の Save 時に作成されます。
func NewStore(path string) *Store {
	return &Store{path: path}
}

// DefaultCredentialsPath は既定の認証情報ファイルパス
// （$HOME/.config/misoca/credentials.json）を返します。
func DefaultCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}
	return filepath.Join(home, ".config", "misoca", "credentials.json"), nil
}

// profileMap はファイル上のデータ形式（プロファイル名 -> 認証情報）です。
type profileMap map[string]Credentials

func (s *Store) readAll() (profileMap, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return profileMap{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("認証情報ファイルの読み込みに失敗しました: %w", err)
	}
	all := profileMap{}
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, fmt.Errorf("認証情報ファイルの解析に失敗しました: %w", err)
	}
	return all, nil
}

func (s *Store) writeAll(all profileMap) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("認証情報ディレクトリの作成に失敗しました: %w", err)
	}
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("認証情報のエンコードに失敗しました: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("認証情報ファイルの書き込みに失敗しました: %w", err)
	}
	return nil
}

// Load はプロファイルの認証情報を読み込みます。
// 保存されていない場合は ErrProfileNotFound を返します。
func (s *Store) Load(profile string) (*Credentials, error) {
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}
	creds, ok := all[profile]
	if !ok {
		return nil, fmt.Errorf("%w (profile=%s)", ErrProfileNotFound, profile)
	}
	return &creds, nil
}

// Save はプロファイルの認証情報を保存します。既存の他プロファイルは保持されます。
func (s *Store) Save(profile string, creds *Credentials) error {
	all, err := s.readAll()
	if err != nil {
		return err
	}
	all[profile] = *creds
	return s.writeAll(all)
}

// Delete はプロファイルの認証情報を削除します。
// 保存されていない場合は ErrProfileNotFound を返します。
func (s *Store) Delete(profile string) error {
	all, err := s.readAll()
	if err != nil {
		return err
	}
	if _, ok := all[profile]; !ok {
		return fmt.Errorf("%w (profile=%s)", ErrProfileNotFound, profile)
	}
	delete(all, profile)
	return s.writeAll(all)
}
