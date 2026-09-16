// contact_group.go は `misoca contact-group`（取引先）コマンド群を実装します。
package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kasseika/misoca-cli/internal/misoca"
	"github.com/kasseika/misoca-cli/internal/output"
)

func contactGroupTable(groups []misoca.ContactGroup) *output.Table {
	table := &output.Table{Header: []string{"ID", "取引先名"}}
	for _, g := range groups {
		id := ""
		if g.ID != nil {
			id = strconv.Itoa(*g.ID)
		}
		table.Rows = append(table.Rows, []string{id, derefString(g.RecipientName)})
	}
	return table
}

// contactGroupIDAction はIDのみを引数に取る取引先操作（trash/untrash）の型です。
type contactGroupIDAction func(ctx context.Context, id int) (*misoca.ContactGroup, error)

func (a *App) newContactGroupCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "contact-group", Short: "取引先の操作"}

	var trashed bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "取引先一覧を取得します",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := misoca.ListContactGroupsParams{}
			if cmd.Flags().Changed("trashed") {
				params.Trashed = &trashed
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
			groups, err := client.ListContactGroups(ctx, params)
			if err != nil {
				return err
			}
			return output.Write(a.Stdout, format, groups, contactGroupTable(groups))
		},
	}
	listCmd.Flags().BoolVar(&trashed, "trashed", false, "非表示にした取引先のみ取得します")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <id>",
		Short: "取引先を取得します",
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
			group, err := client.GetContactGroup(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, group)
		},
	})

	var filePath, recipientName string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "取引先を作成します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			var req misoca.CreateContactGroupRequest
			if filePath != "" {
				if err := readJSONBody(a.Stdin, filePath, &req); err != nil {
					return err
				}
			}
			if recipientName != "" {
				req.RecipientName = recipientName
			}
			client, err := a.NewClient(flags.profile)
			if err != nil {
				return err
			}
			ctx, cancel := timeoutContext(cmd.Context(), flags)
			defer cancel()
			created, err := client.CreateContactGroup(ctx, req)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, created)
		},
	}
	createCmd.Flags().StringVar(&filePath, "file", "", "リクエストボディのJSONファイル（\"-\"で標準入力）")
	createCmd.Flags().StringVar(&recipientName, "recipient-name", "", "取引先名（--fileの値を上書きします）")
	cmd.AddCommand(createCmd)

	cmd.AddCommand(a.newContactGroupActionCmd(flags, "trash", "取引先を非表示にします", func(c *misoca.Client) contactGroupIDAction { return c.TrashContactGroup }))
	cmd.AddCommand(a.newContactGroupActionCmd(flags, "untrash", "非表示にした取引先を表示に戻します", func(c *misoca.Client) contactGroupIDAction { return c.UntrashContactGroup }))

	return cmd
}

func (a *App) newContactGroupActionCmd(flags *globalFlags, use, short string, actionFor func(*misoca.Client) contactGroupIDAction) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id>",
		Short: short,
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
			group, err := actionFor(client)(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, group)
		},
	}
}
