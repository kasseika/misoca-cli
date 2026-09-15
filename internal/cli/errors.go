// errors.go は misoca CLI の終了コードの決定ロジックを実装します。
//
// 終了コードの意味:
//
//	0: 成功
//	1: 一般的な失敗（ネットワーク・ファイルI/Oなど）
//	2: 使い方の誤り（不正なフラグ値など）
//	3: 未認証・トークンリフレッシュ失敗
//	4: APIがエラーを返した
package cli

import (
	"errors"
	"fmt"

	"github.com/mtane0412/misoca-cli/internal/auth"
	"github.com/mtane0412/misoca-cli/internal/misoca"
)

const (
	exitCodeSuccess         = 0
	exitCodeGeneral         = 1
	exitCodeUsageError      = 2
	exitCodeUnauthenticated = 3
	exitCodeAPIError        = 4
)

// exitError は明示的な終了コードを持つエラーです。
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

// usageErrorf は使い方の誤り（終了コード2）を表すエラーを生成します。
func usageErrorf(format string, args ...any) error {
	return &exitError{code: exitCodeUsageError, err: fmt.Errorf(format, args...)}
}

// exitCodeFor はエラーの種類から終了コードを決定します。
func exitCodeFor(err error) int {
	if err == nil {
		return exitCodeSuccess
	}

	var ee *exitError
	if errors.As(err, &ee) {
		return ee.code
	}

	var apiErr *misoca.APIError
	if errors.As(err, &apiErr) {
		return exitCodeAPIError
	}

	if errors.Is(err, auth.ErrProfileNotFound) {
		return exitCodeUnauthenticated
	}

	return exitCodeGeneral
}
