package analyzer

import "go/token"

// Analyzer はコード解析を行うインターフェース
type Analyzer interface {
	AnalyzeDir(dir string) (*AnalysisResult, error)
}

// Named は名前を持つ要素のインターフェース
type Named interface {
	GetName() string
}

// Positioned は位置情報を持つ要素のインターフェース
type Positioned interface {
	GetPosition() token.Position
}

// Packaged はパッケージ情報を持つ要素のインターフェース
type Packaged interface {
	GetPackage() string
	GetFile() string
}

// CodeElement は基本的なコード要素のインターフェース
type CodeElement interface {
	Named
	Positioned
	Packaged
}

// StructType は構造体型のインターフェース
type StructType interface {
	CodeElement
	GetFields() []FieldInfo
	GetMethods() []FuncInfo
}

// InterfaceType はインターフェース型のインターフェース
type InterfaceType interface {
	CodeElement
	GetMethods() []FuncInfo
}

// FuncType は関数型のインターフェース
type FuncType interface {
	CodeElement
	GetReceiver() string
	GetParams() []FieldInfo
	GetResults() []FieldInfo
	GetBodyCalls() []string
}
