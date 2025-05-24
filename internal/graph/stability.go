package graph

import (
	"strings"
)

type NodeStability struct {
	NodeID      NodeID
	OutDegree   int     // Ce: 出次数（このノードが依存している数）
	InDegree    int     // Ca: 入次数（このノードに依存している数）
	Instability float64 // I = Ce / (Ca + Ce): 不安定度（0=安定、1=不安定）
}

// PackageStability はパッケージレベルの不安定度情報
type PackageStability struct {
	PackageName string
	OutDegree   int     // Ce: パッケージが依存している他パッケージの数
	InDegree    int     // Ca: パッケージに依存している他パッケージの数
	Instability float64 // I = Ce / (Ca + Ce): 不安定度（0=安定、1=不安定）
}

type StabilityResult struct {
	NodeStabilities    map[NodeID]*NodeStability
	PackageStabilities map[string]*PackageStability
}

// CalculateStability: 依存グラフから各ノードの不安定度を算出
func CalculateStability(g *DependencyGraph) *StabilityResult {
	inDegree := make(map[NodeID]int)
	outDegree := make(map[NodeID]int)

	// 出次数（依存数）
	for from, tos := range g.Edges {
		outDegree[from] = len(tos)
		for to := range tos {
			inDegree[to]++
		}
	}

	result := &StabilityResult{
		NodeStabilities:    make(map[NodeID]*NodeStability),
		PackageStabilities: make(map[string]*PackageStability),
	}

	// ノードレベルの不安定度計算
	for id := range g.Nodes {
		ce := outDegree[id] // Ce: 出次数（このノードが依存している数）
		ca := inDegree[id]  // Ca: 入次数（このノードに依存している数）
		var instability float64
		if ce+ca == 0 {
			instability = 1.0 // 孤立ノードは不安定とする
		} else {
			instability = float64(ce) / float64(ce+ca) // I = Ce / (Ca + Ce)
		}
		result.NodeStabilities[id] = &NodeStability{
			NodeID:      id,
			OutDegree:   ce,
			InDegree:    ca,
			Instability: instability,
		}
	}

	// パッケージレベルの不安定度計算
	result.PackageStabilities = calculatePackageStability(g)

	return result
}

// calculatePackageStability はパッケージレベルの不安定度を計算
func calculatePackageStability(g *DependencyGraph) map[string]*PackageStability {
	packageDeps := make(map[string]map[string]struct{}) // パッケージ間の依存関係
	packages := make(map[string]struct{})               // 全パッケージのセット

	// 1. 構造体・インターフェース・関数間の依存関係からパッケージ間依存を抽出
	for from, tos := range g.Edges {
		fromNode := g.Nodes[from]
		if fromNode == nil || fromNode.Kind == NodePackage {
			continue
		}

		fromPkg := fromNode.Package
		packages[fromPkg] = struct{}{}

		if packageDeps[fromPkg] == nil {
			packageDeps[fromPkg] = make(map[string]struct{})
		}

		for to := range tos {
			toNode := g.Nodes[to]
			if toNode == nil || toNode.Kind == NodePackage {
				continue
			}

			toPkg := toNode.Package
			packages[toPkg] = struct{}{}

			// 同じパッケージ内の依存関係は除外
			if fromPkg != toPkg {
				packageDeps[fromPkg][toPkg] = struct{}{}
			}
		}
	}

	// 2. パッケージノード間の直接的な依存関係も含める
	for from, tos := range g.Edges {
		fromNode := g.Nodes[from]
		if fromNode == nil || fromNode.Kind != NodePackage {
			continue
		}

		// パッケージノードのIDから実際のパッケージ名を抽出
		fromPkg := extractPackageNameFromNodeID(string(fromNode.ID))
		if fromPkg == "" {
			continue
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
			if toNode.Kind == NodePackage {
				// パッケージノードの場合
				toPkg = extractPackageNameFromNodeID(string(toNode.ID))
			} else {
				// 通常のノードの場合
				toPkg = toNode.Package
			}

			if toPkg == "" {
				continue
			}
			packages[toPkg] = struct{}{}

			// 同じパッケージ内の依存関係は除外
			if fromPkg != toPkg {
				packageDeps[fromPkg][toPkg] = struct{}{}
			}
		}
	}

	// パッケージレベルの入次数・出次数を計算
	packageInDegree := make(map[string]int)
	packageOutDegree := make(map[string]int)

	for fromPkg, toPkgs := range packageDeps {
		packageOutDegree[fromPkg] = len(toPkgs)
		for toPkg := range toPkgs {
			packageInDegree[toPkg]++
		}
	}

	// パッケージレベルの不安定度を計算
	result := make(map[string]*PackageStability)
	for pkg := range packages {
		ce := packageOutDegree[pkg] // Ce: 出次数（このパッケージが依存している数）
		ca := packageInDegree[pkg]  // Ca: 入次数（このパッケージに依存している数）
		var instability float64
		if ce+ca == 0 {
			instability = 1.0 // 孤立パッケージは不安定とする
		} else {
			instability = float64(ce) / float64(ce+ca) // I = Ce / (Ca + Ce)
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

// extractPackageNameFromNodeID はパッケージノードのIDからパッケージ名を抽出
func extractPackageNameFromNodeID(nodeID string) string {
	// "package:パッケージ名" の形式からパッケージ名を抽出
	if strings.HasPrefix(nodeID, "package:") {
		return strings.TrimPrefix(nodeID, "package:")
	}
	return ""
}
