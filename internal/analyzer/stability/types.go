package stability

import "github.com/harakeishi/depsee/internal/types"

// NodeStability represents stability metrics for a node
type NodeStability struct {
	NodeID      types.NodeID
	OutDegree   int     // Ce: 出次数（このノードが依存している数）
	InDegree    int     // Ca: 入次数（このノードに依存している数）
	Instability float64 // I = Ce / (Ca + Ce): 不安定度（0=安定、1=不安定）
}

// PackageStability represents stability metrics for a package
type PackageStability struct {
	PackageName string
	OutDegree   int     // Ce: パッケージが依存している他パッケージの数
	InDegree    int     // Ca: パッケージに依存している他パッケージの数
	Instability float64 // I = Ce / (Ca + Ce): 不安定度（0=安定、1=不安定）
}

// SDPViolation represents a Stable Dependencies Principle violation
type SDPViolation struct {
	From              types.NodeID // 依存元ノード
	To                types.NodeID // 依存先ノード
	FromInstability   float64      // 依存元の不安定度
	ToInstability     float64      // 依存先の不安定度
	ViolationSeverity float64      // 違反の深刻度（不安定度の差）
}

// Result contains the complete stability analysis results
type Result struct {
	NodeStabilities    map[types.NodeID]*NodeStability
	PackageStabilities map[string]*PackageStability
	SDPViolations      []SDPViolation // SDP違反のリスト
}

// NewResult creates a new stability result
func NewResult() *Result {
	return &Result{
		NodeStabilities:    make(map[types.NodeID]*NodeStability),
		PackageStabilities: make(map[string]*PackageStability),
		SDPViolations:      make([]SDPViolation, 0),
	}
}