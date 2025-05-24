package graph

import "github.com/harakeishi/depsee/internal/analyzer"

// Builder は依存グラフを構築するサービス
type Builder struct{}

// NewBuilder は新しいBuilderインスタンスを作成
func NewBuilder() GraphBuilder {
	return &Builder{}
}

// NewGraphBuilder は新しいGraphBuilderインスタンスを作成
func NewGraphBuilder() GraphBuilder {
	return &Builder{}
}

// BuildDependencyGraph は静的解析結果から依存グラフを構築
func (b *Builder) BuildDependencyGraph(result *analyzer.AnalysisResult) *DependencyGraph {
	return BuildDependencyGraph(result)
}

// BuildDependencyGraphWithPackages はパッケージ間依存関係を含む依存グラフを構築
func (b *Builder) BuildDependencyGraphWithPackages(result *analyzer.AnalysisResult, targetDir string) *DependencyGraph {
	return BuildDependencyGraphWithPackages(result, targetDir)
}
