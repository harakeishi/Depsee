package graph

type NodeStability struct {
	NodeID      NodeID
	OutDegree   int     // Ce: 出次数（このノードが依存している数）
	InDegree    int     // Ca: 入次数（このノードに依存している数）
	Instability float64 // I = Ce / (Ca + Ce): 不安定度（0=安定、1=不安定）
}

type StabilityResult struct {
	NodeStabilities map[NodeID]*NodeStability
}

// CalculateStability: 依存グラフから各ノードの安定度を算出
func CalculateStability(g *DependencyGraph) *StabilityResult {
	inDegree := make(map[NodeID]int)
	outDegree := make(map[NodeID]int)

	// 出次数（依存数）
	for from, tos := range g.Edges {
		outDegree[from] = len(tos)
		for to := range tos {
			inDegree[to]++
		}
	}

	result := &StabilityResult{NodeStabilities: make(map[NodeID]*NodeStability)}
	for id := range g.Nodes {
		ce := outDegree[id] // Ce: 出次数（このノードが依存している数）
		ca := inDegree[id]  // Ca: 入次数（このノードに依存している数）
		var instability float64
		if ce+ca == 0 {
			instability = 1.0 // 孤立ノードは不安定とする
		} else {
			instability = float64(ce) / float64(ce+ca) // I = Ce / (Ca + Ce)
		}
		result.NodeStabilities[id] = &NodeStability{
			NodeID:      id,
			OutDegree:   ce,
			InDegree:    ca,
			Instability: instability,
		}
	}
	return result
}
