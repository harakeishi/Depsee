// Package types はDepseeアプリケーション全体で使用される共通型定義を提供します。
// ドメインモデルの基本型やValue Objectsを集約し、パッケージ間の型の重複を防ぎます。
package types

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