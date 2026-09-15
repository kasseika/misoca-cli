// generate_test.go は internal/codegen の Swagger 定義からの Go 構造体生成ロジックを検証するテストです。
package main

import (
	"go/format"
	"os"
	"strings"
	"testing"
)

// TestGenerate_Fixture は、固定の Swagger フィクスチャから期待通りの Go ソースが
// 生成されることを検証します（RED: 実装前に失敗することを確認するテスト）。
func TestGenerate_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/fixture.json")
	if err != nil {
		t.Fatalf("フィクスチャの読み込みに失敗しました: %v", err)
	}

	got, err := GenerateModels(raw)
	if err != nil {
		t.Fatalf("GenerateModels が失敗しました: %v", err)
	}

	wantSrc := `package misoca

// ContactGroup は ApiEntity_ContactGroup model を表す構造体です。
type ContactGroup struct {
	// 取引先を識別する一意なID
	ID *int ` + "`json:\"id,omitempty\"`" + `
	// 取引先名
	RecipientName *string ` + "`json:\"recipient_name,omitempty\"`" + `
	// 銀行口座
	BankAccounts []InvoiceBankAccount ` + "`json:\"bank_accounts,omitempty\"`" + `
	// メールの件名
	MailSettings *ContactGroupMailSettings ` + "`json:\"mail_settings,omitempty\"`" + `
	// 消費税設定('USE_SENDER': 自社情報の設定, 'INCLUDE': 税込表示)
	TaxOption *string ` + "`json:\"tax_option,omitempty\"`" + `
}

// ContactGroupMailSettings は ApiEntity_ContactGroup_MailSettings model を表す構造体です。
type ContactGroupMailSettings struct {
	// メールの件名:請求書
	Invoice *string ` + "`json:\"invoice,omitempty\"`" + `
}

// InvoiceBankAccount は ApiEntity_Invoice_BankAccount model を表す構造体です。
type InvoiceBankAccount struct {
	// 口座詳細
	Detail *string ` + "`json:\"detail,omitempty\"`" + `
}

// InvoiceItem は ApiEntity_Invoice_InvoiceItem model を表す構造体です。
type InvoiceItem struct {
	// 品目名
	Name *string ` + "`json:\"name,omitempty\"`" + `
	// 単価
	UnitPrice *int ` + "`json:\"unit_price,omitempty\"`" + `
}

// CreateInvoiceRequest は postInvoice model を表す構造体です。
type CreateInvoiceRequest struct {
	// 取引先ID
	ContactID int ` + "`json:\"contact_id,omitempty\"`" + `
	// 明細
	Items []InvoiceItem ` + "`json:\"items,omitempty\"`" + `
	// タグ
	Tags []string ` + "`json:\"tags,omitempty\"`" + `
}
`
	want, err := format.Source([]byte(wantSrc))
	if err != nil {
		t.Fatalf("期待値のフォーマットに失敗しました: %v", err)
	}

	if strings.TrimSpace(string(got)) != strings.TrimSpace(string(want)) {
		t.Errorf("生成結果が期待値と一致しません\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
