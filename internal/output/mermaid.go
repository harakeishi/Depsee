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
		Instability float64
	}
	var nodes []nodeWithStability
	for id, n := range g.Nodes {
		inst := 0.0
		if s, ok := stability.NodeStabilities[id]; ok {
			inst = s.Instability
		}
		nodes = append(nodes, nodeWithStability{
			ID: id, Name: n.Name, Instability: inst,
		})
	}
	// 安定度降順でソート
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Instability > nodes[j].Instability
	})

	out := "graph TD\n"
	for _, n := range nodes {
		safeID := mermaidSafeID(n.ID)
		out += fmt.Sprintf("    %s[\"%s<br>安定度:%.2f\"]\n", safeID, n.Name, n.Instability)
	}
	for from, tos := range g.Edges {
		safeFrom := mermaidSafeID(from)
		for to := range tos {
			safeTo := mermaidSafeID(to)
			out += fmt.Sprintf("    %s --> %s\n", safeFrom, safeTo)
		}
	}
	return out
}
