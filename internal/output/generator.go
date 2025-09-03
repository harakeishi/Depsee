package output

import (
	"github.com/harakeishi/depsee/internal/analyzer/stability"
	"github.com/harakeishi/depsee/internal/graph"
)

// Generator は出力を生成するサービス
type Generator struct{}

// NewGenerator は新しいGeneratorインスタンスを作成
func NewGenerator() OutputGenerator {
	return &Generator{}
}



// GenerateMermaid はMermaid記法の相関図を生成
func (g *Generator) GenerateMermaid(dependencyGraph *graph.DependencyGraph, stabilityResult *stability.Result) string {
	return GenerateMermaid(dependencyGraph, stabilityResult)
}

// GenerateMermaidWithOptions はオプション付きでMermaid記法の相関図を生成
func (g *Generator) GenerateMermaidWithOptions(dependencyGraph *graph.DependencyGraph, stabilityResult *stability.Result, highlightSDPViolations bool) string {
	return GenerateMermaidWithOptions(dependencyGraph, stabilityResult, highlightSDPViolations)
}
