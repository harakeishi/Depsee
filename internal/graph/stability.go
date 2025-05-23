package graph

type NodeStability struct {
	NodeID      NodeID
	OutDegree   int     // 依存数
	InDegree    int     // 非依存数
	Instability float64 // 安定度（変更容易度）
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
		out := outDegree[id]
		in_ := inDegree[id]
		var instability float64
		if out+in_ == 0 {
			instability = 1.0
		} else {
			instability = float64(in_) / float64(out+in_)
		}
		result.NodeStabilities[id] = &NodeStability{
			NodeID:      id,
			OutDegree:   out,
			InDegree:    in_,
			Instability: instability,
		}
	}
	return result
}
