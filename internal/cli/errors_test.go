// errors_test.go は終了コード決定ロジック（exitCodeFor）を検証します。
package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mtane0412/misoca-cli/internal/auth"
	"github.com/mtane0412/misoca-cli/internal/misoca"
)

func TestExitCodeFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nilはゼロ", nil, 0},
		{"一般的なエラーは1", errors.New("何らかの失敗"), 1},
		{"usageErrorfは2", usageErrorf("不正な値です: %s", "abc"), 2},
		{"ErrProfileNotFoundは3", auth.ErrProfileNotFound, 3},
		{"ErrProfileNotFoundをラップしたエラーも3", fmt.Errorf("認証情報の取得に失敗: %w", auth.ErrProfileNotFound), 3},
		{"APIErrorは4", &misoca.APIError{StatusCode: 422, Reasons: []string{"取引先名を入力してください"}}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCodeFor(tt.err); got != tt.want {
				t.Errorf("exitCodeFor(%v) = %d, 期待値 = %d", tt.err, got, tt.want)
			}
		})
	}
}
