package analyzer

import "go/token"

type FieldInfo struct {
	Name string
	Type string // 型名（例: OtherStruct, *OtherStruct, []OtherStruct など）
}

type StructInfo struct {
	Name     string
	Package  string
	File     string
	Position token.Position
	Methods  []FuncInfo  // この構造体に属するメソッド
	Fields   []FieldInfo // フィールド情報
	// フィールド情報なども拡張可能
}

type InterfaceInfo struct {
	Name     string
	Package  string
	File     string
	Position token.Position
	Methods  []FuncInfo // インターフェースのメソッド
}

type FuncInfo struct {
	Name      string
	Receiver  string // メソッドの場合、レシーバ型名
	Package   string
	File      string
	Position  token.Position
	Params    []FieldInfo // 引数情報
	Results   []FieldInfo // 戻り値情報
	BodyCalls []string    // 本体で呼び出している関数名リスト
}
