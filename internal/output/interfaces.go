package output

import "github.com/harakeishi/depsee/internal/graph"

// OutputGenerator はMermaid記法の出力を生成するインターフェース
type OutputGenerator interface {
	GenerateMermaid(dependencyGraph *graph.DependencyGraph, stability *graph.StabilityResult) string
	GenerateMermaidWithOptions(dependencyGraph *graph.DependencyGraph, stability *graph.StabilityResult, highlightSDPViolations bool) string
}
