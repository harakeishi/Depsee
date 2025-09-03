package output

import (
	"github.com/harakeishi/depsee/internal/analyzer/stability"
	"github.com/harakeishi/depsee/internal/graph"
)

// OutputGenerator はMermaid記法の出力を生成するインターフェース
type OutputGenerator interface {
	GenerateMermaid(dependencyGraph *graph.DependencyGraph, stabilityResult *stability.Result) string
	GenerateMermaidWithOptions(dependencyGraph *graph.DependencyGraph, stabilityResult *stability.Result, highlightSDPViolations bool) string
}
