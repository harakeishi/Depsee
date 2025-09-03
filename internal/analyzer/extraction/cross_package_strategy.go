package extraction

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/types"
	"github.com/harakeishi/depsee/internal/utils"
)

// CrossPackageDependencyExtractor extracts cross-package dependencies
type CrossPackageDependencyExtractor struct {
	ctx *Context
}

// NewCrossPackageDependencyExtractor creates a new cross-package dependency extractor
func NewCrossPackageDependencyExtractor(ctx *Context) *CrossPackageDependencyExtractor {
	return &CrossPackageDependencyExtractor{ctx: ctx}
}

// ExtractDependencies extracts cross-package dependencies
func (e *CrossPackageDependencyExtractor) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo
	
	// First, extract import mappings
	imports := e.extractImports(file)
	
	// Then extract cross-package calls and type usage
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				funcName := node.Name.Name
				fromID := types.NewNodeID(packageName, funcName)
				
				// Extract cross-package calls from function body
				crossCalls := e.extractCrossPackageCalls(node.Body, imports)
				for _, call := range crossCalls {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   call.ToID,
						Type: types.CrossPackageDependency,
					})
					logger.Debug("パッケージ間呼び出し依存関係追加", "from", fromID, "to", call.ToID, "call", call.CallName)
				}
			}
		}
		return true
	})
	
	return dependencies, nil
}

// Name returns the name of the strategy
func (e *CrossPackageDependencyExtractor) Name() string {
	return "CrossPackageDependency"
}

// CrossPackageCall represents a cross-package call
type CrossPackageCall struct {
	CallName string
	ToID     types.NodeID
}

// extractImports extracts import mappings from the file
func (e *CrossPackageDependencyExtractor) extractImports(file *ast.File) map[string]string {
	imports := make(map[string]string)
	
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		var alias string
		
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			// Extract package name from path
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}
		
		imports[alias] = importPath
	}
	
	return imports
}

// extractCrossPackageCalls extracts cross-package calls from a function body
func (e *CrossPackageDependencyExtractor) extractCrossPackageCalls(body *ast.BlockStmt, imports map[string]string) []CrossPackageCall {
	var calls []CrossPackageCall
	
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					packageAlias := ident.Name
					funcName := selector.Sel.Name
					
					// Check if this is a cross-package call
					if importPath, exists := imports[packageAlias]; exists {
						// Extract target package name
						targetPkg := e.extractPackageAlias(importPath, packageAlias)
						if targetPkg != "" && utils.IsLocalPackage(importPath) {
							toID := types.NewNodeID(targetPkg, funcName)
							calls = append(calls, CrossPackageCall{
								CallName: packageAlias + "." + funcName,
								ToID:     toID,
							})
						}
					}
				}
			}
		}
		return true
	})
	
	return calls
}

// extractPackageAlias extracts package alias from import path
func (e *CrossPackageDependencyExtractor) extractPackageAlias(importPath, alias string) string {
	// Handle special cases
	if alias == "_" || alias == "." {
		return ""
	}
	
	// For local packages, use the last part of the path
	if strings.Contains(importPath, "/") {
		parts := strings.Split(importPath, "/")
		return parts[len(parts)-1]
	}
	
	// For standard library or simple packages
	return importPath
}
