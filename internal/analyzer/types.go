// Package analyzer は Go言語のコード解析機能を提供します。
// 構造体、インターフェース、関数の情報抽出や依存関係の分析を行います。
package analyzer

import "go/token"

// StructInfo は構造体の情報を表します。
// 構造体の基本情報（名前、パッケージ、ファイル位置等）とフィールド、
// およびメソッドの情報を含みます。
type StructInfo struct {
	Name     string          // 構造体名
	Package  string          // 所属パッケージ名
	File     string          // 定義されているファイルパス
	Position token.Position  // ファイル内での位置情報
	Fields   []FieldInfo     // 構造体が持つフィールドの一覧
	Methods  []FuncInfo      // 構造体に関連付けられたメソッドの一覧
}

// InterfaceInfo はインターフェースの情報を表します。
// インターフェースの基本情報（名前、パッケージ、ファイル位置等）と
// 定義されているメソッドの情報を含みます。
type InterfaceInfo struct {
	Name     string          // インターフェース名
	Package  string          // 所属パッケージ名
	File     string          // 定義されているファイルパス
	Position token.Position  // ファイル内での位置情報
	Methods  []FuncInfo      // インターフェースで定義されているメソッドの一覧
}

// FuncInfo は関数・メソッドの情報を表します。
// 関数の基本情報（名前、パッケージ、ファイル位置等）とシグネチャ、
// および関数本体での呼び出し情報を含みます。
type FuncInfo struct {
	Name      string          // 関数・メソッド名
	Package   string          // 所属パッケージ名
	File      string          // 定義されているファイルパス
	Position  token.Position  // ファイル内での位置情報
	Receiver  string          // レシーバ型名（メソッドの場合のみ設定される）
	Params    []FieldInfo     // 引数の一覧
	Results   []FieldInfo     // 戻り値の一覧
	BodyCalls []string        // 関数本体内で呼び出している関数名の一覧
}

// FieldInfo はフィールドの情報を表します。
// 構造体のフィールドや関数の引数・戻り値の型情報を保持します。
type FieldInfo struct {
	Name string // フィールド名（無名フィールドの場合は空文字）
	Type string // フィールドの型名
}

// PackageInfo はパッケージの情報を表します。
// パッケージの基本情報（名前、パス、ファイル位置等）と
// import文の情報を含みます。
type PackageInfo struct {
	Name     string          // パッケージ名
	Path     string          // パッケージパス（現在未使用）
	File     string          // パッケージ宣言があるファイルパス
	Position token.Position  // ファイル内での位置情報
	Imports  []ImportInfo    // パッケージがimportしているパッケージの一覧
}

// ImportInfo はimport文の情報を表します。
// importされているパッケージのパスとエイリアス名を保持します。
type ImportInfo struct {
	Path  string // importパス（例: "github.com/example/pkg"）
	Alias string // エイリアス名（エイリアスがない場合はパッケージ名）
}
