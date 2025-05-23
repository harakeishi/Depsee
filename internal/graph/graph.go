package graph

import (
	"strings"

	"github.com/harakeishi/depsee/internal/analyzer"
)

type NodeKind int

const (
	NodeStruct NodeKind = iota
	NodeInterface
	NodeFunc
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
// （まずはノード登録のみ。今後エッジ抽出を拡張）
func BuildDependencyGraph(result *analyzer.AnalysisResult) *DependencyGraph {
	g := NewDependencyGraph()

	// ノード登録
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

	// --- 依存エッジ抽出 ---
	// 1. 構造体のフィールド型依存
	for _, s := range result.Structs {
		fromID := NodeID(s.Package + "." + s.Name)
		for _, field := range s.Fields {
			// フィールド型名から依存先ノードIDを推測（同一パッケージ前提、今後拡張可）
			typeName := field.Type
			// ポインタやスライス、map等のプリフィックスを除去
			typeName = strings.TrimLeft(typeName, "*[]")
			if typeName == "" || typeName == s.Name {
				continue // 自己参照や型名不明はスキップ
			}
			// 構造体・インターフェース・外部型の区別は今後拡張
			toID := NodeID(s.Package + "." + typeName)
			if _, ok := g.Nodes[toID]; ok {
				g.AddEdge(fromID, toID)
			}
		}
	}

	// 2. 関数・メソッドのシグネチャ依存
	for _, f := range result.Functions {
		fromID := NodeID(f.Package + "." + f.Name)
		for _, param := range f.Params {
			typeName := param.Type
			typeName = strings.TrimLeft(typeName, "*[]")
			if typeName == "" {
				continue
			}
			toID := NodeID(f.Package + "." + typeName)
			if _, ok := g.Nodes[toID]; ok {
				g.AddEdge(fromID, toID)
			}
		}
		for _, resultField := range f.Results {
			typeName := resultField.Type
			typeName = strings.TrimLeft(typeName, "*[]")
			if typeName == "" {
				continue
			}
			toID := NodeID(f.Package + "." + typeName)
			if _, ok := g.Nodes[toID]; ok {
				g.AddEdge(fromID, toID)
			}
		}
	}
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := NodeID(s.Package + "." + m.Name)
			for _, param := range m.Params {
				typeName := param.Type
				typeName = strings.TrimLeft(typeName, "*[]")
				if typeName == "" {
					continue
				}
				toID := NodeID(s.Package + "." + typeName)
				if _, ok := g.Nodes[toID]; ok {
					g.AddEdge(fromID, toID)
				}
			}
			for _, resultField := range m.Results {
				typeName := resultField.Type
				typeName = strings.TrimLeft(typeName, "*[]")
				if typeName == "" {
					continue
				}
				toID := NodeID(s.Package + "." + typeName)
				if _, ok := g.Nodes[toID]; ok {
					g.AddEdge(fromID, toID)
				}
			}
		}
	}

	// 3. インターフェース実装依存なども今後拡張

	return g
}
