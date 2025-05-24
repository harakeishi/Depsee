package graph

import (
	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/logger"
)

type NodeKind int

const (
	NodeStruct NodeKind = iota
	NodeInterface
	NodeFunc
	NodePackage
)

type NodeID string // 例: "package.StructName"

type Node struct {
	ID      NodeID
	Kind    NodeKind
	Name    string
	Package string
}

type Edge struct {
	From NodeID
	To   NodeID
	// 依存の種類（フィールド/シグネチャ/本体/実装）も必要に応じて
}

type DependencyGraph struct {
	Nodes map[NodeID]*Node
	Edges map[NodeID]map[NodeID]struct{} // From→Toの多重辺排除
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[NodeID]*Node),
		Edges: make(map[NodeID]map[NodeID]struct{}),
	}
}

func (g *DependencyGraph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
}

func (g *DependencyGraph) AddEdge(from, to NodeID) {
	if g.Edges[from] == nil {
		g.Edges[from] = make(map[NodeID]struct{})
	}
	g.Edges[from][to] = struct{}{}
}

// BuildDependencyGraph: 静的解析結果から依存グラフを構築
func BuildDependencyGraph(result *analyzer.AnalysisResult) *DependencyGraph {
	logger.Info("依存グラフ構築開始")

	g := NewDependencyGraph()

	// ノード登録
	registerNodes(result, g)

	// 型解析器の初期化
	typeResolver := analyzer.NewTypeResolver()

	// 依存関係抽出（戦略パターンを使用）
	extractors := []DependencyExtractor{
		NewFieldDependencyExtractor(typeResolver),
		&SignatureDependencyExtractor{},
		&BodyCallDependencyExtractor{},
	}

	for _, extractor := range extractors {
		extractor.Extract(result, g)
	}

	logger.Info("依存グラフ構築完了", "nodes", len(g.Nodes), "edges", countEdges(g))
	return g
}

// BuildDependencyGraphWithPackages: パッケージ間依存関係を含む依存グラフを構築
func BuildDependencyGraphWithPackages(result *analyzer.AnalysisResult, targetDir string) *DependencyGraph {
	logger.Info("パッケージ間依存関係を含む依存グラフ構築開始")

	g := NewDependencyGraph()

	// ノード登録
	registerNodes(result, g)

	// 型解析器の初期化
	typeResolver := analyzer.NewTypeResolver()

	// 依存関係抽出（戦略パターンを使用）
	extractors := []DependencyExtractor{
		NewFieldDependencyExtractor(typeResolver),
		&SignatureDependencyExtractor{},
		&BodyCallDependencyExtractor{},
		NewPackageDependencyExtractor(targetDir), // パッケージ間依存関係抽出器を追加
	}

	for _, extractor := range extractors {
		extractor.Extract(result, g)
	}

	logger.Info("パッケージ間依存関係を含む依存グラフ構築完了", "nodes", len(g.Nodes), "edges", countEdges(g))
	return g
}

// countEdges はエッジ数をカウント
func countEdges(g *DependencyGraph) int {
	count := 0
	for _, tos := range g.Edges {
		count += len(tos)
	}
	return count
}

// registerNodes はノードを登録する
func registerNodes(result *analyzer.AnalysisResult, g *DependencyGraph) {
	// 構造体ノード登録
	for _, s := range result.Structs {
		id := NodeID(s.Package + "." + s.Name)
		node := &Node{
			ID:      id,
			Kind:    NodeStruct,
			Name:    s.Name,
			Package: s.Package,
		}
		g.AddNode(node)
	}

	// インターフェースノード登録
	for _, i := range result.Interfaces {
		id := NodeID(i.Package + "." + i.Name)
		node := &Node{
			ID:      id,
			Kind:    NodeInterface,
			Name:    i.Name,
			Package: i.Package,
		}
		g.AddNode(node)
	}

	// 関数ノード登録
	for _, f := range result.Functions {
		id := NodeID(f.Package + "." + f.Name)
		node := &Node{
			ID:      id,
			Kind:    NodeFunc,
			Name:    f.Name,
			Package: f.Package,
		}
		g.AddNode(node)
	}
}
