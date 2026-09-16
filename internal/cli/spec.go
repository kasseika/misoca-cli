// spec.go は `misoca spec check` コマンドを実装します。
//
// Misoca API は「現状有姿・予告なく変更される可能性がある」と明言されているため、
// リポジトリに同梱した Swagger 仕様書スナップショット（api/swagger.json）と
// ライブの仕様書を比較し、差異があれば早期に検知できるようにします。
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/kasseika/misoca-cli/api"
)

// defaultSpecCheckURL はMisocaのライブSwagger仕様書のURLです。
const defaultSpecCheckURL = "https://app.misoca.jp/api/v3/swagger_doc"

func (a *App) newSpecCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "spec", Short: "API仕様の確認"}
	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "リポジトリ同梱の仕様書スナップショットとMisocaのライブ仕様書を比較します",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runSpecCheck(cmd)
		},
	})
	return cmd
}

func (a *App) runSpecCheck(cmd *cobra.Command) error {
	url := a.SpecCheckURL
	if url == "" {
		url = defaultSpecCheckURL
	}

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("リクエストの構築に失敗しました: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ライブ仕様書の取得に失敗しました: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	liveBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ライブ仕様書の読み込みに失敗しました: %w", err)
	}

	var live, snapshot any
	if err := json.Unmarshal(liveBytes, &live); err != nil {
		return fmt.Errorf("ライブ仕様書のJSON解析に失敗しました: %w", err)
	}
	if err := json.Unmarshal(api.SwaggerJSON, &snapshot); err != nil {
		return fmt.Errorf("同梱の仕様書スナップショットのJSON解析に失敗しました: %w", err)
	}

	if !reflect.DeepEqual(live, snapshot) {
		return fmt.Errorf(
			"API仕様（Misoca API v3）がリポジトリ同梱のスナップショット（api/swagger.json）と異なります。" +
				"内容を確認し、スナップショットを更新してください")
	}

	_, _ = fmt.Fprintln(a.Stdout, "差分はありません（api/swagger.json はライブ仕様と一致しています）")
	return nil
}
