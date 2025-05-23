package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/harakeishi/depsee/internal/graph"
)

// mermaidSafeID はMermaidの予約語を含むIDを安全な形式に変換します
func mermaidSafeID(id graph.NodeID) string {
	// 予約語のリスト
	reservedWords := []string{"graph", "subgraph", "end", "flowchart", "TD", "BT", "RL", "LR"}

	// IDをパッケージ名と識別子に分割
	parts := strings.Split(string(id), ".")
	if len(parts) != 2 {
		return string(id)
	}

	pkgName, ident := parts[0], parts[1]

	// 予約語を含む場合は接頭辞を追加
	for _, word := range reservedWords {
		if strings.EqualFold(pkgName, word) {
			pkgName = "pkg_" + pkgName
		}
		if strings.EqualFold(ident, word) {
			ident = "id_" + ident
		}
	}

	return pkgName + "." + ident
}

func GenerateMermaid(g *graph.DependencyGraph, stability *graph.StabilityResult) string {
	type nodeWithStability struct {
		ID          graph.NodeID
		Name        string
		Kind        graph.NodeKind
		Instability float64
	}
	var nodes []nodeWithStability
	for id, n := range g.Nodes {
		inst := 0.0
		if s, ok := stability.NodeStabilities[id]; ok {
			inst = s.Instability
		}
		nodes = append(nodes, nodeWithStability{
			ID: id, Name: n.Name, Kind: n.Kind, Instability: inst,
		})
	}
	// 安定度降順でソート
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Instability > nodes[j].Instability
	})

	out := "graph TD\n"
	// スタイル定義
	out += "    %% スタイル定義\n"
	out += "    classDef default fill:#fff,stroke:#333,stroke-width:2px\n"

	// パッケージごとにノードをグループ化
	packageNodes := make(map[string][]nodeWithStability)
	for _, n := range nodes {
		pkg := strings.Split(string(n.ID), ".")[0]
		packageNodes[pkg] = append(packageNodes[pkg], n)
	}

	// パッケージごとにsubgraphを作成
	for pkg, pkgNodes := range packageNodes {
		out += fmt.Sprintf("    subgraph %s\n", pkg)
		for _, n := range pkgNodes {
			safeID := mermaidSafeID(n.ID)
			var shape string
			switch n.Kind {
			case graph.NodeStruct:
				shape = "[\"%s<br>構造体<br>安定度:%.2f\"]"
			case graph.NodeInterface:
				shape = "{{\"%s<br>インターフェース<br>安定度:%.2f\"}}"
			case graph.NodeFunc:
				shape = "(\"%s<br>関数<br>安定度:%.2f\")"
			}
			out += fmt.Sprintf("        %s%s\n", safeID, fmt.Sprintf(shape, n.Name, n.Instability))
			out += fmt.Sprintf("        class %s default\n", safeID)
		}
		out += "    end\n"
	}

	// エッジの定義（subgraphの外に配置）
	for from, tos := range g.Edges {
		safeFrom := mermaidSafeID(from)
		for to := range tos {
			safeTo := mermaidSafeID(to)
			out += fmt.Sprintf("    %s --> %s\n", safeFrom, safeTo)
		}
	}
	return out
}
