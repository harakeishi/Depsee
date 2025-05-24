package output

import "github.com/harakeishi/depsee/internal/graph"

// Generator は出力を生成するサービス
type Generator struct{}

// NewGenerator は新しいGeneratorインスタンスを作成
func NewGenerator() OutputGenerator {
	return &Generator{}
}

// NewOutputGenerator は新しいOutputGeneratorインスタンスを作成
func NewOutputGenerator() OutputGenerator {
	return &Generator{}
}

// GenerateMermaid はMermaid記法の相関図を生成
func (g *Generator) GenerateMermaid(dependencyGraph *graph.DependencyGraph, stability *graph.StabilityResult) string {
	return GenerateMermaid(dependencyGraph, stability)
}

// GenerateMermaidWithOptions はオプション付きでMermaid記法の相関図を生成
func (g *Generator) GenerateMermaidWithOptions(dependencyGraph *graph.DependencyGraph, stability *graph.StabilityResult, highlightSDPViolations bool) string {
	return GenerateMermaidWithOptions(dependencyGraph, stability, highlightSDPViolations)
}
