// Package types はDepseeアプリケーション全体で使用される共通型定義を提供します。
// ドメインモデルの基本型やValue Objectsを集約し、パッケージ間の型の重複を防ぎます。
package types

import (
	"go/token"
)

// NodeID はグラフのノードを一意に識別するIDです。
// 通常は "package.Name" 形式の文字列で構成されます。
// パッケージノードの場合は "package:name" 形式を使用します。
type NodeID string

// NewNodeID は通常のノード（構造体、インターフェース、関数）のIDを生成します。
func NewNodeID(packageName, name string) NodeID {
	return NodeID(packageName + "." + name)
}

// NewPackageNodeID はパッケージノードのIDを生成します。
func NewPackageNodeID(packageName string) NodeID {
	return NodeID("package:" + packageName)
}

// IsPackageNode はNodeIDがパッケージノードのIDかどうかを判定します。
func (id NodeID) IsPackageNode() bool {
	return len(id) > 8 && id[:8] == "package:"
}

// String はNodeIDを文字列として返します。
func (id NodeID) String() string {
	return string(id)
}

// DependencyType は依存関係の種類を表す列挙型です。
type DependencyType int

const (
	// FieldDependency は構造体のフィールドによる依存関係です
	FieldDependency DependencyType = iota
	// SignatureDependency は関数・メソッドのシグネチャ（引数・戻り値）による依存関係です
	SignatureDependency
	// BodyCallDependency は関数本体内の呼び出しによる依存関係です
	BodyCallDependency
	// CrossPackageDependency はパッケージ間の関数呼び出しや型使用による依存関係です
	CrossPackageDependency
	// PackageDependency はimport文によるパッケージ間依存関係です
	PackageDependency
)

// String はDependencyTypeを文字列として返します。
func (t DependencyType) String() string {
	switch t {
	case FieldDependency:
		return "field"
	case SignatureDependency:
		return "signature"
	case BodyCallDependency:
		return "body_call"
	case CrossPackageDependency:
		return "cross_package"
	case PackageDependency:
		return "package"
	default:
		return "unknown"
	}
}

// DependencyInfo は依存関係情報を表す構造体です。
// 依存元ノード、依存先ノード、依存関係の種類を定義します。
type DependencyInfo struct {
	From NodeID         // 依存元のノードID
	To   NodeID         // 依存先のノードID
	Type DependencyType // 依存関係の種類
}

// StructInfo は構造体の情報を表します。
// 構造体の基本情報（名前、パッケージ、ファイル位置等）とフィールド、
// およびメソッドの情報を含みます。
type StructInfo struct {
	Name     string         // 構造体名
	Package  string         // 所属パッケージ名
	File     string         // 定義されているファイルパス
	Position token.Position // ファイル内での位置情報
	Fields   []FieldInfo    // 構造体が持つフィールドの一覧
	Methods  []FuncInfo     // 構造体に関連付けられたメソッドの一覧
}

// InterfaceInfo はインターフェースの情報を表します。
// インターフェースの基本情報（名前、パッケージ、ファイル位置等）と
// 定義されているメソッドの情報を含みます。
type InterfaceInfo struct {
	Name     string         // インターフェース名
	Package  string         // 所属パッケージ名
	File     string         // 定義されているファイルパス
	Position token.Position // ファイル内での位置情報
	Methods  []FuncInfo     // インターフェースで定義されているメソッドの一覧
}

// FuncInfo は関数・メソッドの情報を表します。
// 関数の基本情報（名前、パッケージ、ファイル位置等）とシグネチャ、
// および関数本体での呼び出し情報を含みます。
type FuncInfo struct {
	Name      string         // 関数・メソッド名
	Package   string         // 所属パッケージ名
	File      string         // 定義されているファイルパス
	Position  token.Position // ファイル内での位置情報
	Receiver  string         // レシーバ型名（メソッドの場合のみ設定される）
	Params    []FieldInfo    // 引数の一覧
	Results   []FieldInfo    // 戻り値の一覧
	BodyCalls []string       // 関数本体内で呼び出している関数名の一覧
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
	Name     string         // パッケージ名
	Path     string         // パッケージパス（現在未使用）
	File     string         // パッケージ宣言があるファイルパス
	Position token.Position // ファイル内での位置情報
	Imports  []ImportInfo   // パッケージがimportしているパッケージの一覧
}

// ImportInfo はimport文の情報を表します。
// importされているパッケージのパスとエイリアス名を保持します。
type ImportInfo struct {
	Path  string // importパス（例: "github.com/example/pkg"）
	Alias string // エイリアス名（エイリアスがない場合はパッケージ名）
}

// Result は解析結果を格納する構造体です。
// 解析で抽出された構造体、インターフェース、関数、パッケージの情報と
// それらの間の依存関係情報を含みます。
type Result struct {
	Structs      []StructInfo     // 抽出された構造体の一覧
	Interfaces   []InterfaceInfo  // 抽出されたインターフェースの一覧
	Functions    []FuncInfo       // 抽出された関数の一覧
	Packages     []PackageInfo    // 解析対象パッケージの一覧
	Dependencies []DependencyInfo // 抽出された依存関係の一覧
}

// CreateNodeMap は解析結果から全ノードの存在チェック用マップを作成します。
// 構造体、インターフェース、関数の全てのノードIDを登録し、
// 依存先ノードの存在確認に使用されます。
// このメソッドは主に依存関係抽出で使用されます。
func (r *Result) CreateNodeMap() map[NodeID]struct{} {
	nodeMap := make(map[NodeID]struct{})

	// 構造体ノード登録
	for _, s := range r.Structs {
		id := NewNodeID(s.Package, s.Name)
		nodeMap[id] = struct{}{}
	}

	// インターフェースノード登録
	for _, i := range r.Interfaces {
		id := NewNodeID(i.Package, i.Name)
		nodeMap[id] = struct{}{}
	}

	// 関数ノード登録
	for _, f := range r.Functions {
		id := NewNodeID(f.Package, f.Name)
		nodeMap[id] = struct{}{}
	}

	return nodeMap
}