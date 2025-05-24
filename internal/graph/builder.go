package graph

import "github.com/harakeishi/depsee/internal/analyzer"

// Builder は依存グラフを構築するサービス
type Builder struct{}

// NewBuilder は新しいBuilderインスタンスを作成
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildDependencyGraph は静的解析結果から依存グラフを構築
func (b *Builder) BuildDependencyGraph(result *analyzer.AnalysisResult) *DependencyGraph {
	return BuildDependencyGraph(result)
}
