package graph

import "github.com/harakeishi/depsee/internal/analyzer"

// GraphBuilder は依存グラフを構築するインターフェース
type GraphBuilder interface {
	BuildDependencyGraph(result *analyzer.Result) *DependencyGraph
	BuildDependencyGraphWithPackages(result *analyzer.Result, targetDir string) *DependencyGraph
}

// StabilityCalculator は安定度を計算するインターフェース
type StabilityCalculator interface {
	CalculateStability(graph *DependencyGraph) *StabilityResult
}
