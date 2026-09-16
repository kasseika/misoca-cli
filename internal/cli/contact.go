// contact.go は `misoca contact`（送り先）コマンド群を実装します。
package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kasseika/misoca-cli/internal/misoca"
	"github.com/kasseika/misoca-cli/internal/output"
)

func contactTable(contacts []misoca.Contact) *output.Table {
	table := &output.Table{Header: []string{"ID", "名前", "郵便番号", "電話番号"}}
	for _, c := range contacts {
		id := ""
		if c.ID != nil {
			id = strconv.Itoa(*c.ID)
		}
		table.Rows = append(table.Rows, []string{id, derefString(c.RecipientName), derefString(c.RecipientZipCode), derefString(c.RecipientTelNo)})
	}
	return table
}

// contactIDAction はIDのみを引数に取る送り先操作（trash/untrash）の型です。
type contactIDAction func(ctx context.Context, id int) (*misoca.Contact, error)

func (a *App) newContactCmd(flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "contact", Short: "送り先の操作"}

	var trashed bool
	var contactGroupID int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "送り先一覧を取得します",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := misoca.ListContactsParams{}
			if cmd.Flags().Changed("trashed") {
				params.Trashed = &trashed
			}
			if cmd.Flags().Changed("contact-group-id") {
				params.ContactGroupID = &contactGroupID
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
			contacts, err := client.ListContacts(ctx, params)
			if err != nil {
				return err
			}
			return output.Write(a.Stdout, format, contacts, contactTable(contacts))
		},
	}
	listCmd.Flags().BoolVar(&trashed, "trashed", false, "非表示にした送り先のみ取得します")
	listCmd.Flags().IntVar(&contactGroupID, "contact-group-id", 0, "特定の取引先の送り先のみ取得します")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <id>",
		Short: "送り先を取得します",
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
			contact, err := client.GetContact(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, contact)
		},
	})

	var filePath, contactGroupIDStr, recipientName string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "送り先を作成します",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := parseFormat(flags)
			if err != nil {
				return err
			}
			var req misoca.CreateContactRequest
			if filePath != "" {
				if err := readJSONBody(a.Stdin, filePath, &req); err != nil {
					return err
				}
			}
			if contactGroupIDStr != "" {
				id, err := parseID(contactGroupIDStr)
				if err != nil {
					return err
				}
				req.ContactGroupID = id
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
			created, err := client.CreateContact(ctx, req)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, created)
		},
	}
	createCmd.Flags().StringVar(&filePath, "file", "", "リクエストボディのJSONファイル（\"-\"で標準入力）")
	createCmd.Flags().StringVar(&contactGroupIDStr, "contact-group-id", "", "取引先ID（--fileの値を上書きします）")
	createCmd.Flags().StringVar(&recipientName, "recipient-name", "", "取引先名（--fileの値を上書きします）")
	cmd.AddCommand(createCmd)

	cmd.AddCommand(a.newContactActionCmd(flags, "trash", "送り先を非表示にします", func(c *misoca.Client) contactIDAction { return c.TrashContact }))
	cmd.AddCommand(a.newContactActionCmd(flags, "untrash", "非表示にした送り先を表示に戻します", func(c *misoca.Client) contactIDAction { return c.UntrashContact }))

	return cmd
}

func (a *App) newContactActionCmd(flags *globalFlags, use, short string, actionFor func(*misoca.Client) contactIDAction) *cobra.Command {
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
			contact, err := actionFor(client)(ctx, id)
			if err != nil {
				return err
			}
			return writeSingle(a.Stdout, format, contact)
		},
	}
}
