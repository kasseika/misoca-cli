// invoice_test.go は `misoca invoice` コマンド群の振る舞いを検証します。
package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestInvoiceList_TableFormat は、一覧取得結果が表形式で
// 請求ステータス・入金ステータスを日本語ラベルに変換して出力されることを検証します。
func TestInvoiceList_TableFormat(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoices" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"subject":"9月分システム開発費","invoice_status":1,"payment_status":0}]`))
	})

	code := app.Execute([]string{"invoice", "list", "--format", "table"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "9月分システム開発費") {
		t.Errorf("件名が出力に含まれていません: %q", out)
	}
	if !strings.Contains(out, "請求済") {
		t.Errorf("請求ステータスのラベルが出力に含まれていません: %q", out)
	}
	if !strings.Contains(out, "未入金") {
		t.Errorf("入金ステータスのラベルが出力に含まれていません: %q", out)
	}
}

// TestInvoiceList_QueryFlags は、検索用フラグがクエリパラメータへ
// 正しく反映されることを検証します。
func TestInvoiceList_QueryFlags(t *testing.T) {
	var gotQuery string
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})

	code := app.Execute([]string{
		"invoice", "list",
		"--payment-status", "unpaid",
		"--invoice-status", "submitted",
	})
	if code != 0 {
		t.Fatalf("終了コード = %d", code)
	}
	if !strings.Contains(gotQuery, "payment_status=unpaid") {
		t.Errorf("クエリにpayment_statusが含まれていません: %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "invoice_status=submitted") {
		t.Errorf("クエリにinvoice_statusが含まれていません: %q", gotQuery)
	}
}

// TestInvoiceGet_JSON は `misoca invoice get <id>` がJSONを出力することを検証します。
func TestInvoiceGet_JSON(t *testing.T) {
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoice/42" {
			t.Fatalf("パス = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"subject":"9月分システム開発費"}`))
	})

	code := app.Execute([]string{"invoice", "get", "42"})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("出力のJSON解析に失敗しました: %v (%s)", err, stdout.String())
	}
	if got["subject"] != "9月分システム開発費" {
		t.Errorf("subject = %v", got["subject"])
	}
}

// TestInvoiceGet_InvalidID は数値でないIDが使い方の誤り（終了コード2）に
// なることを検証します。
func TestInvoiceGet_InvalidID(t *testing.T) {
	app, _, stderr := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("不正なIDの場合、APIは呼ばれるべきではありません")
	})

	code := app.Execute([]string{"invoice", "get", "abc"})
	if code != 2 {
		t.Fatalf("終了コード = %d, 期待値 = 2, stderr=%s", code, stderr.String())
	}
}

// TestInvoiceCreate_WithFile は --file 経由でリクエストボディを
// JSONファイルから読み込めることを検証します。
func TestInvoiceCreate_WithFile(t *testing.T) {
	var gotBody map[string]any
	app, stdout, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"subject":"9月分システム開発費"}`))
	})

	dir := t.TempDir()
	bodyPath := dir + "/invoice.json"
	if err := writeTestFile(bodyPath, `{"contact_id":5,"subject":"9月分システム開発費"}`); err != nil {
		t.Fatalf("テスト用ファイルの作成に失敗しました: %v", err)
	}

	code := app.Execute([]string{"invoice", "create", "--file", bodyPath})
	if code != 0 {
		t.Fatalf("終了コード = %d, stdout=%s", code, stdout.String())
	}
	if gotBody["subject"] != "9月分システム開発費" {
		t.Errorf("リクエストボディ = %+v", gotBody)
	}
}

// TestInvoiceCreate_FlagsOverrideFile は、--file と併用した頻出フラグが
// ファイルの値を上書きすることを検証します。
func TestInvoiceCreate_FlagsOverrideFile(t *testing.T) {
	var gotBody map[string]any
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	})

	dir := t.TempDir()
	bodyPath := dir + "/invoice.json"
	if err := writeTestFile(bodyPath, `{"contact_id":5,"subject":"元の件名"}`); err != nil {
		t.Fatalf("テスト用ファイルの作成に失敗しました: %v", err)
	}

	code := app.Execute([]string{"invoice", "create", "--file", bodyPath, "--subject", "上書き後の件名"})
	if code != 0 {
		t.Fatalf("終了コード = %d", code)
	}
	if gotBody["subject"] != "上書き後の件名" {
		t.Errorf("subject = %v, 期待値 = 上書き後の件名", gotBody["subject"])
	}
}

// TestInvoicePdf_SavesToFile は取得したPDFバイナリが指定パスに保存されることを検証します。
func TestInvoicePdf_SavesToFile(t *testing.T) {
	content := []byte("%PDF-1.4 fake")
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	})

	dir := t.TempDir()
	outPath := dir + "/out.pdf"
	code := app.Execute([]string{"invoice", "pdf", "1", "-o", outPath})
	if code != 0 {
		t.Fatalf("終了コード = %d", code)
	}
	got, err := readTestFile(outPath)
	if err != nil {
		t.Fatalf("保存されたファイルの読み込みに失敗しました: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("保存内容 = %q, 期待値 = %q", got, content)
	}
}

// TestInvoicePostalMail_RequiresConfirmation は、課金の可能性がある郵送指示が
// 確認プロンプトなしには実行されず、--yes を付けた場合のみ実行されることを検証します。
func TestInvoicePostalMail_RequiresConfirmation(t *testing.T) {
	called := false
	app, _, _ := newTestApp(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoice":{"id":1}}`))
	})
	app.Stdin = strings.NewReader("n\n")

	code := app.Execute([]string{"invoice", "postal-mail", "1"})
	if code == 0 {
		t.Error("確認をしなかった場合は失敗すべきです")
	}
	if called {
		t.Error("確認なしにAPIが呼ばれてはいけません")
	}

	code = app.Execute([]string{"invoice", "postal-mail", "1", "--yes"})
	if code != 0 {
		t.Fatalf("--yes 指定時の終了コード = %d", code)
	}
	if !called {
		t.Error("--yes 指定時はAPIが呼ばれるべきです")
	}
}
