// generate.go は Misoca API v3 の Swagger 2.0 仕様書（api/swagger.json）から
// internal/misoca 用の Go 構造体定義を生成するロジックです。
//
// Swagger の JSON オブジェクトはキー順序に意味を持たせているため（定義順・プロパティ順が
// ドキュメントの読みやすさに影響する）、標準の map デコードでは失われる順序情報を
// json.Decoder のトークンストリームを使って独自に保持しています。
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"strings"
)

// property は Swagger の Schema Object のうち、本ツールが扱うフィールドのみを表します。
//
// allOf は Misoca の Swagger 仕様書で「単一の $ref をラップするだけ」の用途にのみ
// 使われているため、要素数 1 の allOf はその $ref と同一視して扱います。
type property struct {
	Type        string     `json:"type"`
	Format      string     `json:"format"`
	Ref         string     `json:"$ref"`
	AllOf       []property `json:"allOf"`
	Items       *property  `json:"items"`
	Description string     `json:"description"`
}

// effectiveRef は $ref を直接持つ場合はそれを、allOf 経由で単一の $ref を
// 参照している場合はその $ref を返します。どちらでもない場合は空文字を返します。
func (p property) effectiveRef() string {
	if p.Ref != "" {
		return p.Ref
	}
	if len(p.AllOf) == 1 && p.AllOf[0].Ref != "" {
		return p.AllOf[0].Ref
	}
	return ""
}

// oneLineDescription は Swagger の description に含まれる改行・連続空白を
// 単一の半角スペースに畳み込みます。Go の行コメントは改行を含められないため、
// 生成コードの構文エラーを防ぐのに必要です。
func oneLineDescription(desc string) string {
	return strings.Join(strings.Fields(desc), " ")
}

// rawDefinition は Swagger definitions 配下の 1 モデル定義です。
// properties はキー順序を保持するため json.RawMessage のまま保持します。
type rawDefinition struct {
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Properties  json.RawMessage `json:"properties"`
}

// rawSpec は本ツールが読み取る Swagger 仕様書の最小サブセットです。
type rawSpec struct {
	Definitions json.RawMessage `json:"definitions"`
}

// requestRenames は post*/put* のようなオペレーション由来の定義名を、
// リクエストボディとして分かりやすい構造体名に変換するための明示的な対応表です。
// これらは機械的な命名規則（ApiEntity_ プレフィックスの除去や重複語の畳み込み）が
// 通用しないため、個別に列挙します。
var requestRenames = map[string]string{
	"postContact":              "CreateContactRequest",
	"postContactGroup":         "CreateContactGroupRequest",
	"postInvoice":              "CreateInvoiceRequest",
	"putInvoiceIdPaid":         "MarkInvoicePaidRequest",
	"postEstimate":             "CreateEstimateRequest",
	"postEstimateIdDistribute": "DistributeEstimateRequest",
	"postDeliverySlip":         "CreateDeliverySlipRequest",
	"postDealingItem":          "CreateDealingItemRequest",
}

// GenerateModels は Swagger 仕様書（JSON バイト列）から gofmt 済みの Go ソースを生成します。
func GenerateModels(specJSON []byte) ([]byte, error) {
	var spec rawSpec
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, fmt.Errorf("swagger仕様のパースに失敗しました: %w", err)
	}

	defNames, err := orderedObjectKeys(spec.Definitions)
	if err != nil {
		return nil, fmt.Errorf("definitionsのキー順序取得に失敗しました: %w", err)
	}

	var defs map[string]rawDefinition
	if err := json.Unmarshal(spec.Definitions, &defs); err != nil {
		return nil, fmt.Errorf("definitionsのパースに失敗しました: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("package misoca\n\n")

	for _, defName := range defNames {
		def := defs[defName]
		structName := goStructName(defName)
		isResponse := strings.HasPrefix(defName, "ApiEntity_")

		fieldNames, err := orderedObjectKeys(def.Properties)
		if err != nil {
			return nil, fmt.Errorf("%s のプロパティ順序取得に失敗しました: %w", defName, err)
		}
		var props map[string]property
		if err := json.Unmarshal(def.Properties, &props); err != nil {
			return nil, fmt.Errorf("%s のプロパティのパースに失敗しました: %w", defName, err)
		}

		fmt.Fprintf(&buf, "// %s は %s を表す構造体です。\n", structName, oneLineDescription(def.Description))
		fmt.Fprintf(&buf, "type %s struct {\n", structName)
		for _, fieldName := range fieldNames {
			p := props[fieldName]
			if p.Description != "" {
				fmt.Fprintf(&buf, "\t// %s\n", oneLineDescription(p.Description))
			}
			goName := goFieldName(fieldName)
			goType := fieldGoType(p, isResponse)
			fmt.Fprintf(&buf, "\t%s %s `json:\"%s,omitempty\"`\n", goName, goType, fieldName)
		}
		buf.WriteString("}\n\n")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("生成コードのフォーマットに失敗しました: %w\n---\n%s", err, buf.String())
	}
	return formatted, nil
}

// orderedObjectKeys は JSON オブジェクトのトップレベルキーを、出現順を保った
// まま返します。raw が空（未設定）の場合は空スライスを返します。
func orderedObjectKeys(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("オブジェクトの開始が期待されましたが %v でした", tok)
	}

	var keys []string
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("オブジェクトキーが文字列ではありません: %v", keyTok)
		}
		keys = append(keys, key)

		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

// goStructName は Swagger の definition 名を Go の構造体名に変換します。
//
// ルール:
//   - requestRenames に登録されている名前はそのまま採用する
//   - "ApiEntity_" プレフィックスは除去する
//   - アンダースコア区切りの隣接する語について、後の語が前の語で始まる場合は
//     前の語を畳み込む（例: Invoice_InvoiceItem -> InvoiceItem）
func goStructName(defName string) string {
	if renamed, ok := requestRenames[defName]; ok {
		return renamed
	}

	parts := strings.Split(defName, "_")
	if len(parts) > 0 && parts[0] == "ApiEntity" {
		parts = parts[1:]
	}
	for len(parts) > 1 && strings.HasPrefix(parts[1], parts[0]) {
		parts = parts[1:]
	}
	return strings.Join(parts, "")
}

// goFieldName は Swagger の snake_case なフィールド名を Go の PascalCase な
// フィールド名に変換します。"id" またはそれで終わる語（例: contact_id）は
// Go の慣習に合わせて "ID" と表記します。
func goFieldName(jsonName string) string {
	parts := strings.Split(jsonName, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if p == "id" {
			parts[i] = "ID"
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

// resolveRefName は "#/definitions/Xxx" 形式の $ref から Go の構造体名を解決します。
func resolveRefName(ref string) string {
	segs := strings.Split(ref, "/")
	return goStructName(segs[len(segs)-1])
}

// itemGoType は配列要素の Go 型を返します。配列要素は常に非ポインタ型として
// 扱います（配列自体が nil によって「値なし」を表現できるため）。
func itemGoType(p property) string {
	if ref := p.effectiveRef(); ref != "" {
		return resolveRefName(ref)
	}
	switch p.Type {
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "string":
		return "string"
	default:
		return "interface{}"
	}
}

// fieldGoType はフィールドの Go 型を返します。
//
// pointerForScalars が true の場合（レスポンスモデル）、スカラー型（文字列・
// 数値・真偽値）はゼロ値と未設定を区別するためポインタ型にします。オブジェクト
// 参照は常にポインタ、配列は常にスライス（要素は非ポインタ）です。
func fieldGoType(p property, pointerForScalars bool) string {
	if ref := p.effectiveRef(); ref != "" {
		return "*" + resolveRefName(ref)
	}
	if p.Type == "array" {
		if p.Items == nil {
			return "[]interface{}"
		}
		return "[]" + itemGoType(*p.Items)
	}
	switch p.Type {
	case "integer":
		if pointerForScalars {
			return "*int"
		}
		return "int"
	case "number":
		if pointerForScalars {
			return "*float64"
		}
		return "float64"
	case "boolean":
		if pointerForScalars {
			return "*bool"
		}
		return "bool"
	case "string":
		if pointerForScalars {
			return "*string"
		}
		return "string"
	default:
		return "interface{}"
	}
}
