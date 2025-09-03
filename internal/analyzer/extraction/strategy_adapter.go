package extraction

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/types"
)

// StrategyBasedExtractor adapts the new strategy-based extraction to the old interface
type StrategyBasedExtractor struct {
	strategies []ExtractionStrategy
	targetDir  string
}

// NewStrategyBasedExtractor creates a new strategy-based extractor
func NewStrategyBasedExtractor(targetDir string) *StrategyBasedExtractor {
	return &StrategyBasedExtractor{
		strategies: make([]ExtractionStrategy, 0),
		targetDir:  targetDir,
	}
}

// AddStrategy adds an extraction strategy
func (e *StrategyBasedExtractor) AddStrategy(strategy ExtractionStrategy) {
	e.strategies = append(e.strategies, strategy)
}

// ExtractFromFiles extracts dependencies from a list of Go files
func (e *StrategyBasedExtractor) ExtractFromFiles(files []string) ([]DependencyInfo, error) {
	var allDependencies []DependencyInfo
	
	logger.Debug("新しいstrategyベース依存関係抽出開始", "files", len(files), "strategies", len(e.strategies))
	
	for _, filePath := range files {
		deps, err := e.extractFromFile(filePath)
		if err != nil {
			logger.Error("ファイル依存関係抽出エラー", "file", filePath, "error", err)
			continue // エラーがあっても他のファイルは処理を続行
		}
		allDependencies = append(allDependencies, deps...)
	}
	
	logger.Debug("新しいstrategyベース依存関係抽出完了", "total_dependencies", len(allDependencies))
	return allDependencies, nil
}

// extractFromFile extracts dependencies from a single file
func (e *StrategyBasedExtractor) extractFromFile(filePath string) ([]DependencyInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	
	packageName := file.Name.Name
	ctx := NewContext(fset, packageName)
	
	// Extract imports for the context
	ctx.ImportMap = e.extractImportMap(file)
	
	var allDependencies []DependencyInfo
	
	for _, strategy := range e.strategies {
		deps, err := strategy.ExtractDependencies(file, fset, packageName)
		if err != nil {
			logger.Error("戦略依存関係抽出エラー", "strategy", strategy.Name(), "file", filePath, "error", err)
			continue
		}
		allDependencies = append(allDependencies, deps...)
		
		logger.Debug("戦略依存関係抽出", "strategy", strategy.Name(), "file", filePath, "dependencies", len(deps))
	}
	
	return allDependencies, nil
}

// extractImportMap extracts import mappings from the AST file
func (e *StrategyBasedExtractor) extractImportMap(file *ast.File) map[string]string {
	importMap := make(map[string]string)
	
	for _, imp := range file.Imports {
		importPath := imp.Path.Value
		// Remove quotes
		importPath = importPath[1 : len(importPath)-1]
		
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			// Extract package name from path
			parts := importPath
			if lastSlash := len(importPath) - 1; lastSlash >= 0 {
				for i := lastSlash; i >= 0; i-- {
					if importPath[i] == '/' {
						parts = importPath[i+1:]
						break
					}
				}
			}
			alias = parts
		}
		
		importMap[alias] = importPath
	}
	
	return importMap
}

// createStrategyContext creates a context for a specific strategy
func (e *StrategyBasedExtractor) createStrategyContext(strategy ExtractionStrategy, baseCtx *Context) *Context {
	// For strategies that need targetDir, we need to pass it somehow
	// This is a simple approach - in a more complex system you might use dependency injection
	switch strategy.(type) {
	case *PackageDependencyExtractor:
		// PackageDependencyExtractor already has targetDir set during creation
		return baseCtx
	default:
		return baseCtx
	}
}

// DefaultStrategyBasedExtractor creates a default strategy-based extractor with all strategies
func DefaultStrategyBasedExtractor(targetDir string) *StrategyBasedExtractor {
	extractor := NewStrategyBasedExtractor(targetDir)
	
	// Create a shared context (this will be overridden per file)
	ctx := NewContext(token.NewFileSet(), "")
	
	// Add all strategies
	extractor.AddStrategy(NewFieldDependencyExtractor(ctx))
	extractor.AddStrategy(NewSignatureDependencyExtractor(ctx))
	extractor.AddStrategy(NewBodyCallDependencyExtractor(ctx))
	extractor.AddStrategy(NewPackageDependencyExtractor(ctx, targetDir))
	extractor.AddStrategy(NewCrossPackageDependencyExtractor(ctx))
	
	return extractor
}

// ConvertToDependencyInfo converts extraction.DependencyInfo to analyzer.DependencyInfo
func ConvertToDependencyInfo(deps []DependencyInfo) []interface{} {
	result := make([]interface{}, len(deps))
	for i, dep := range deps {
		result[i] = struct {
			From types.NodeID
			To   types.NodeID
			Type types.DependencyType
		}{
			From: dep.From,
			To:   dep.To,
			Type: dep.Type,
		}
	}
	return result
}
