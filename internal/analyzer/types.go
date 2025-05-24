package analyzer

import "go/token"

// StructInfo は構造体の情報を表します
type StructInfo struct {
	Name     string
	Package  string
	File     string
	Position token.Position
	Fields   []FieldInfo
	Methods  []FuncInfo
}

// InterfaceInfo はインターフェースの情報を表します
type InterfaceInfo struct {
	Name     string
	Package  string
	File     string
	Position token.Position
	Methods  []FuncInfo
}

// FuncInfo は関数・メソッドの情報を表します
type FuncInfo struct {
	Name      string
	Package   string
	File      string
	Position  token.Position
	Receiver  string
	Params    []FieldInfo
	Results   []FieldInfo
	BodyCalls []string
}

// FieldInfo はフィールドの情報を表します
type FieldInfo struct {
	Name string
	Type string
}

// PackageInfo はパッケージの情報を表します
type PackageInfo struct {
	Name     string
	Path     string
	File     string
	Position token.Position
	Imports  []ImportInfo
}

// ImportInfo はimport文の情報を表します
type ImportInfo struct {
	Path  string
	Alias string
}
