package analyzer

import "go/token"

// TypeInfo は基本的な型情報のインターフェース
type TypeInfo interface {
	GetName() string
	GetPackage() string
	GetFile() string
	GetPosition() token.Position
}

// StructType は構造体専用のインターフェース
type StructType interface {
	TypeInfo
	GetFields() []FieldInfo
	GetMethods() []FuncInfo
}

// InterfaceType はインターフェース専用のインターフェース
type InterfaceType interface {
	TypeInfo
	GetMethods() []FuncInfo
}

// FuncType は関数専用のインターフェース
type FuncType interface {
	TypeInfo
	GetReceiver() string
	GetParams() []FieldInfo
	GetResults() []FieldInfo
	GetBodyCalls() []string
}
