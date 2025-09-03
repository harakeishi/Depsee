package stability

import (
	"github.com/harakeishi/depsee/internal/graph"
	"github.com/harakeishi/depsee/internal/types"
)

// Analyzer defines the interface for stability analysis
type Analyzer interface {
	// Analyze calculates stability metrics for the dependency graph
	Analyze(g *graph.DependencyGraph) *Result
	
	// AnalyzeNode calculates stability for a specific node
	AnalyzeNode(nodeID types.NodeID, g *graph.DependencyGraph) *NodeStability
	
	// AnalyzePackage calculates stability for a specific package
	AnalyzePackage(packageName string, g *graph.DependencyGraph) *PackageStability
	
	// DetectSDPViolations finds violations of the Stable Dependencies Principle
	DetectSDPViolations(g *graph.DependencyGraph) []SDPViolation
}

// analyzer is the default implementation of Analyzer
type analyzer struct{}

// NewAnalyzer creates a new stability analyzer
func NewAnalyzer() Analyzer {
	return &analyzer{}
}

// Analyze performs complete stability analysis
func (a *analyzer) Analyze(g *graph.DependencyGraph) *Result {
	result := NewResult()
	
	// Calculate in/out degrees
	inDegree := make(map[types.NodeID]int)
	outDegree := make(map[types.NodeID]int)
	
	for from, tos := range g.Edges {
		outDegree[from] = len(tos)
		for to := range tos {
			inDegree[to]++
		}
	}
	
	// Calculate node stability
	for id := range g.Nodes {
		ce := outDegree[id] // Ce: 出次数
		ca := inDegree[id]  // Ca: 入次数
		
		var instability float64
		if ce+ca == 0 {
			instability = 1.0 // 孤立ノードは不安定とする
		} else {
			instability = float64(ce) / float64(ce+ca)
		}
		
		result.NodeStabilities[id] = &NodeStability{
			NodeID:      id,
			OutDegree:   ce,
			InDegree:    ca,
			Instability: instability,
		}
	}
	
	// Calculate package stability
	result.PackageStabilities = a.calculatePackageStability(g)
	
	// Detect SDP violations
	result.SDPViolations = a.detectSDPViolations(g, result.NodeStabilities)
	
	return result
}

// AnalyzeNode calculates stability for a specific node
func (a *analyzer) AnalyzeNode(nodeID types.NodeID, g *graph.DependencyGraph) *NodeStability {
	inDegree := 0
	outDegree := 0
	
	// Calculate out-degree
	if edges, exists := g.Edges[nodeID]; exists {
		outDegree = len(edges)
	}
	
	// Calculate in-degree
	for _, tos := range g.Edges {
		if _, exists := tos[nodeID]; exists {
			inDegree++
		}
	}
	
	var instability float64
	if inDegree+outDegree == 0 {
		instability = 1.0
	} else {
		instability = float64(outDegree) / float64(inDegree+outDegree)
	}
	
	return &NodeStability{
		NodeID:      nodeID,
		OutDegree:   outDegree,
		InDegree:    inDegree,
		Instability: instability,
	}
}

// AnalyzePackage calculates stability for a specific package
func (a *analyzer) AnalyzePackage(packageName string, g *graph.DependencyGraph) *PackageStability {
	packageDeps := make(map[string]struct{})
	packageRevDeps := make(map[string]struct{})
	
	for from, tos := range g.Edges {
		fromNode := g.Nodes[from]
		if fromNode == nil {
			continue
		}
		
		for to := range tos {
			toNode := g.Nodes[to]
			if toNode == nil {
				continue
			}
			
			// Track package dependencies
			if fromNode.Package == packageName && toNode.Package != packageName {
				packageDeps[toNode.Package] = struct{}{}
			}
			
			// Track reverse dependencies
			if toNode.Package == packageName && fromNode.Package != packageName {
				packageRevDeps[fromNode.Package] = struct{}{}
			}
		}
	}
	
	outDegree := len(packageDeps)
	inDegree := len(packageRevDeps)
	
	var instability float64
	if inDegree+outDegree == 0 {
		instability = 1.0
	} else {
		instability = float64(outDegree) / float64(inDegree+outDegree)
	}
	
	return &PackageStability{
		PackageName: packageName,
		OutDegree:   outDegree,
		InDegree:    inDegree,
		Instability: instability,
	}
}

// DetectSDPViolations finds SDP violations in the graph
func (a *analyzer) DetectSDPViolations(g *graph.DependencyGraph) []SDPViolation {
	stabilities := make(map[types.NodeID]*NodeStability)
	
	// First calculate all node stabilities
	for id := range g.Nodes {
		stabilities[id] = a.AnalyzeNode(id, g)
	}
	
	return a.detectSDPViolations(g, stabilities)
}

// calculatePackageStability calculates stability for all packages
func (a *analyzer) calculatePackageStability(g *graph.DependencyGraph) map[string]*PackageStability {
	packages := make(map[string]struct{})
	packageDeps := make(map[string]map[string]struct{})
	
	// Collect all packages and their dependencies
	for from, tos := range g.Edges {
		fromNode := g.Nodes[from]
		if fromNode == nil {
			continue
		}
		
		var fromPkg string
		if fromNode.Kind == graph.NodePackage {
			// パッケージノードの場合、パッケージ名を直接使用
			fromPkg = fromNode.Package
		} else {
			// 通常のノードの場合、そのノードのパッケージを使用
			fromPkg = fromNode.Package
		}
		
		packages[fromPkg] = struct{}{}
		
		if packageDeps[fromPkg] == nil {
			packageDeps[fromPkg] = make(map[string]struct{})
		}
		
		for to := range tos {
			toNode := g.Nodes[to]
			if toNode == nil {
				continue
			}
			
			var toPkg string
			if toNode.Kind == graph.NodePackage {
				// パッケージノードの場合、パッケージ名を直接使用
				toPkg = toNode.Package
			} else {
				// 通常のノードの場合、そのノードのパッケージを使用
				toPkg = toNode.Package
			}
			
			packages[toPkg] = struct{}{}
			
			if fromPkg != toPkg {
				packageDeps[fromPkg][toPkg] = struct{}{}
			}
		}
	}
	
	// Calculate in/out degrees for packages
	packageInDegree := make(map[string]int)
	packageOutDegree := make(map[string]int)
	
	for fromPkg, toPkgs := range packageDeps {
		packageOutDegree[fromPkg] = len(toPkgs)
		for toPkg := range toPkgs {
			packageInDegree[toPkg]++
		}
	}
	
	// Create package stability results
	result := make(map[string]*PackageStability)
	for pkg := range packages {
		ce := packageOutDegree[pkg]
		ca := packageInDegree[pkg]
		
		var instability float64
		if ce+ca == 0 {
			instability = 1.0
		} else {
			instability = float64(ce) / float64(ce+ca)
		}
		
		result[pkg] = &PackageStability{
			PackageName: pkg,
			OutDegree:   ce,
			InDegree:    ca,
			Instability: instability,
		}
	}
	
	return result
}

// detectSDPViolations detects violations of the Stable Dependencies Principle
func (a *analyzer) detectSDPViolations(g *graph.DependencyGraph, nodeStabilities map[types.NodeID]*NodeStability) []SDPViolation {
	var violations []SDPViolation
	
	for from, tos := range g.Edges {
		fromStability, fromExists := nodeStabilities[from]
		if !fromExists {
			continue
		}
		
		for to := range tos {
			toStability, toExists := nodeStabilities[to]
			if !toExists {
				continue
			}
			
			// SDP違反の条件: 依存元の不安定度 < 依存先の不安定度
			if fromStability.Instability < toStability.Instability {
				violation := SDPViolation{
					From:              from,
					To:                to,
					FromInstability:   fromStability.Instability,
					ToInstability:     toStability.Instability,
					ViolationSeverity: toStability.Instability - fromStability.Instability,
				}
				violations = append(violations, violation)
			}
		}
	}
	
	return violations
}