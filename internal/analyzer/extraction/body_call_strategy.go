package extraction

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/types"
)

// BodyCallDependencyExtractor extracts dependencies from function body calls
type BodyCallDependencyExtractor struct {
	ctx *Context
}

// NewBodyCallDependencyExtractor creates a new body call dependency extractor
func NewBodyCallDependencyExtractor(ctx *Context) *BodyCallDependencyExtractor {
	return &BodyCallDependencyExtractor{ctx: ctx}
}

// ExtractDependencies extracts body call-based dependencies
func (e *BodyCallDependencyExtractor) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo
	
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				funcName := node.Name.Name
				fromID := types.NewNodeID(packageName, funcName)
				
				// Extract function calls from body
				calls := e.extractCalls(node.Body)
				for _, call := range calls {
					if targetFunc := e.resolveCall(call, packageName); targetFunc != "" {
						toID := types.NewNodeID(packageName, targetFunc)
						dependencies = append(dependencies, DependencyInfo{
							From: fromID,
							To:   toID,
							Type: types.BodyCallDependency,
						})
						logger.Debug("関数呼び出し依存関係追加", "from", fromID, "to", toID, "call", call)
					}
				}
			}
		}
		return true
	})
	
	return dependencies, nil
}

// Name returns the name of the strategy
func (e *BodyCallDependencyExtractor) Name() string {
	return "BodyCallDependency"
}

// extractCalls extracts function calls from a function body
func (e *BodyCallDependencyExtractor) extractCalls(body *ast.BlockStmt) []string {
	var calls []string
	
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if ident, ok := node.Fun.(*ast.Ident); ok {
				calls = append(calls, ident.Name)
			} else if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					calls = append(calls, ident.Name+"."+selector.Sel.Name)
				}
			}
		}
		return true
	})
	
	return calls
}

// resolveCall resolves a function call to a target function name
func (e *BodyCallDependencyExtractor) resolveCall(call, currentPkg string) string {
	// Simple resolution - just return the call name if it doesn't contain a package qualifier
	if !strings.Contains(call, ".") {
		return call
	}
	
	// For qualified calls, return empty string for now (cross-package calls are handled separately)
	return ""
}
