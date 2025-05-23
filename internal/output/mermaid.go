package output

import (
	"fmt"
	"sort"

	"github.com/harakeishi/depsee/internal/graph"
)

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
		out += fmt.Sprintf("    %s[\"%s<br>安定度:%.2f\"]\n", n.ID, n.Name, n.Instability)
	}
	for from, tos := range g.Edges {
		for to := range tos {
			out += fmt.Sprintf("    %s --> %s\n", from, to)
		}
	}
	return out
}
