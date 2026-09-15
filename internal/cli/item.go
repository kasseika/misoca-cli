// item.go は `misoca item`（品目）コマンド群を実装します。
package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mtane0412/misoca-cli/internal/misoca"
	"github.com/mtane0412/misoca-cli/internal/output"
)

func dealingItemTable(items []misoca.DealingItem) *output.Table {
	table := &output.Table{Header: []string{"ID", "品番・品名", "単価", "単位"}}
	for _, item := range items {
		id := ""
		if item.ID != nil {
			id = strconv.Itoa(*item.ID)
		}
		unitPrice := ""
		if item.UnitPrice != nil {
			unitPrice = strconv.FormatFloat(*item.UnitPrice, 'f', -1, 64)
		}
		table.Rows = append(table.Rows, []string{id, derefString(item.Name), unitPrice, derefString(item.UnitName)})
	}
	return table
}

func (a *App) newItemCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "item", Short: "品目の操作"}

	var page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "品目一覧を取得します",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := misoca.ListDealingItemsParams{}
			if cmd.Flags().Changed("page") {
				params.Page = &page
			}
			if cmd.Flags().Changed("per-page") {
				params.PerPage = &perPage
			}
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			items, _, err := client.ListDealingItems(ctx, params)
			if err != nil {
				return err
			}
			return output.Write(a.Stdout, format, items, dealingItemTable(items))
		},
	}
	listCmd.Flags().IntVar(&page, "page", 0, "ページ番号")
	listCmd.Flags().IntVar(&perPage, "per-page", 0, "1ページあたりの件数")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <id>",
		Short: "品目を取得します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			item, err := client.GetDealingItem(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, item)
		},
	})

	var filePath, name, unitName string
	var unitPrice float64
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "品目を作成します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			var req misoca.CreateDealingItemRequest
			if filePath != "" {
				if err := readJSONBody(a.Stdin, filePath, &req); err != nil {
					return err
				}
			}
			if name != "" {
				req.Name = name
			}
			if unitName != "" {
				req.UnitName = unitName
			}
			if cmd.Flags().Changed("unit-price") {
				req.UnitPrice = unitPrice
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			created, err := client.CreateDealingItem(ctx, req)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, created)
		},
	}
	createCmd.Flags().StringVar(&filePath, "file", "", "リクエストボディのJSONファイル（\"-\"で標準入力）")
	createCmd.Flags().StringVar(&name, "name", "", "品番・品名（--fileの値を上書きします）")
	createCmd.Flags().StringVar(&unitName, "unit-name", "", "単位（--fileの値を上書きします）")
	createCmd.Flags().Float64Var(&unitPrice, "unit-price", 0, "単価（--fileの値を上書きします）")
	cmd.AddCommand(createCmd)

	return cmd
}
