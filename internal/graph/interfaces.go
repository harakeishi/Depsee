package graph

import "github.com/harakeishi/depsee/internal/analyzer"

// GraphBuilder defines the interface for building dependency graphs
type GraphBuilder interface {
	// BuildDependencyGraph builds a dependency graph from analysis results
	BuildDependencyGraph(result *analyzer.Result) *DependencyGraph
	
	// BuildDependencyGraphWithPackages builds a dependency graph including package dependencies
	BuildDependencyGraphWithPackages(result *analyzer.Result, targetDir string) *DependencyGraph
}
