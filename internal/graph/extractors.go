package graph

import (
	"strings"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/logger"
)

// DependencyExtractor は依存関係抽出の戦略インターフェース
type DependencyExtractor interface {
	Extract(result *analyzer.AnalysisResult, g *DependencyGraph)
}

// FieldDependencyExtractor は構造体フィールドの依存関係を抽出
type FieldDependencyExtractor struct {
	typeResolver *analyzer.TypeResolver
}

// NewFieldDependencyExtractor は新しいFieldDependencyExtractorを作成
func NewFieldDependencyExtractor(typeResolver *analyzer.TypeResolver) *FieldDependencyExtractor {
	return &FieldDependencyExtractor{
		typeResolver: typeResolver,
	}
}

func (e *FieldDependencyExtractor) Extract(result *analyzer.AnalysisResult, g *DependencyGraph) {
	logger.Debug("フィールド依存関係抽出開始")

	for _, s := range result.Structs {
		fromID := NodeID(s.Package + "." + s.Name)
		logger.Debug("構造体フィールド解析", "struct", s.Name, "package", s.Package)

		for _, field := range s.Fields {
			if toID := e.parseTypeToNodeID(field.Type, s.Package); toID != "" {
				if _, ok := g.Nodes[toID]; ok {
					g.AddEdge(fromID, toID)
					logger.Debug("フィールド依存関係追加", "from", fromID, "to", toID, "field", field.Name)
				} else {
					logger.Debug("依存先ノード未発見", "from", fromID, "to", toID, "field", field.Name)
				}
			}
		}
	}
}

func (e *FieldDependencyExtractor) parseTypeToNodeID(typeName, pkg string) NodeID {
	// より安全な型解析ロジック
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return NodeID(pkg + "." + cleaned)
}

// SignatureDependencyExtractor は関数シグネチャの依存関係を抽出
type SignatureDependencyExtractor struct{}

func (e *SignatureDependencyExtractor) Extract(result *analyzer.AnalysisResult, g *DependencyGraph) {
	// 関数の引数・戻り値の依存関係抽出
	for _, f := range result.Functions {
		fromID := NodeID(f.Package + "." + f.Name)
		e.extractFromParams(f.Params, fromID, f.Package, g)
		e.extractFromParams(f.Results, fromID, f.Package, g)
	}

	// メソッドの引数・戻り値の依存関係抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := NodeID(s.Package + "." + m.Name)
			e.extractFromParams(m.Params, fromID, s.Package, g)
			e.extractFromParams(m.Results, fromID, s.Package, g)
		}
	}
}

func (e *SignatureDependencyExtractor) extractFromParams(params []analyzer.FieldInfo, fromID NodeID, pkg string, g *DependencyGraph) {
	for _, param := range params {
		if toID := e.parseTypeToNodeID(param.Type, pkg); toID != "" {
			if _, ok := g.Nodes[toID]; ok {
				g.AddEdge(fromID, toID)
			}
		}
	}
}

func (e *SignatureDependencyExtractor) parseTypeToNodeID(typeName, pkg string) NodeID {
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return NodeID(pkg + "." + cleaned)
}

// BodyCallDependencyExtractor は関数本体の呼び出し依存関係を抽出
type BodyCallDependencyExtractor struct{}

func (e *BodyCallDependencyExtractor) Extract(result *analyzer.AnalysisResult, g *DependencyGraph) {
	// 関数本体の依存関係抽出
	for _, f := range result.Functions {
		fromID := NodeID(f.Package + "." + f.Name)
		for _, called := range f.BodyCalls {
			toID := NodeID(f.Package + "." + called)
			if _, ok := g.Nodes[toID]; ok {
				g.AddEdge(fromID, toID)
			}
		}
	}

	// メソッド本体の依存関係抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := NodeID(s.Package + "." + m.Name)
			for _, called := range m.BodyCalls {
				toID := NodeID(s.Package + "." + called)
				if _, ok := g.Nodes[toID]; ok {
					g.AddEdge(fromID, toID)
				}
			}
		}
	}
}
