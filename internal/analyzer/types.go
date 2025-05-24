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

// StructType インターフェースの実装
func (s StructInfo) GetName() string             { return s.Name }
func (s StructInfo) GetPackage() string          { return s.Package }
func (s StructInfo) GetFile() string             { return s.File }
func (s StructInfo) GetPosition() token.Position { return s.Position }
func (s StructInfo) GetFields() []FieldInfo      { return s.Fields }
func (s StructInfo) GetMethods() []FuncInfo      { return s.Methods }

type InterfaceInfo struct {
	Name     string
	Package  string
	File     string
	Position token.Position
	Methods  []FuncInfo // インターフェースのメソッド
}

// InterfaceType インターフェースの実装
func (i InterfaceInfo) GetName() string             { return i.Name }
func (i InterfaceInfo) GetPackage() string          { return i.Package }
func (i InterfaceInfo) GetFile() string             { return i.File }
func (i InterfaceInfo) GetPosition() token.Position { return i.Position }
func (i InterfaceInfo) GetMethods() []FuncInfo      { return i.Methods }

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

// FuncType インターフェースの実装
func (f FuncInfo) GetName() string             { return f.Name }
func (f FuncInfo) GetPackage() string          { return f.Package }
func (f FuncInfo) GetFile() string             { return f.File }
func (f FuncInfo) GetPosition() token.Position { return f.Position }
func (f FuncInfo) GetReceiver() string         { return f.Receiver }
func (f FuncInfo) GetParams() []FieldInfo      { return f.Params }
func (f FuncInfo) GetResults() []FieldInfo     { return f.Results }
func (f FuncInfo) GetBodyCalls() []string      { return f.BodyCalls }
