package extraction

import (
	"go/ast"
	"go/token"

	"github.com/harakeishi/depsee/internal/types"
)

// SignatureDependencyExtractor extracts dependencies from function signatures
type SignatureDependencyExtractor struct {
	ctx *Context
}

// NewSignatureDependencyExtractor creates a new signature dependency extractor
func NewSignatureDependencyExtractor(ctx *Context) *SignatureDependencyExtractor {
	return &SignatureDependencyExtractor{ctx: ctx}
}

// ExtractDependencies extracts signature-based dependencies
func (e *SignatureDependencyExtractor) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo
	
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			funcName := node.Name.Name
			fromID := types.NewNodeID(packageName, funcName)
			
			// Extract dependencies from parameters
			if node.Type.Params != nil {
				for _, field := range node.Type.Params.List {
					typeStr := extractTypeString(field.Type)
					if targetType := e.resolveType(typeStr, packageName); targetType != "" {
						toID := types.NewNodeID(packageName, targetType)
						dependencies = append(dependencies, DependencyInfo{
							From: fromID,
							To:   toID,
							Type: types.SignatureDependency,
						})
					}
				}
			}
			
			// Extract dependencies from return types
			if node.Type.Results != nil {
				for _, field := range node.Type.Results.List {
					typeStr := extractTypeString(field.Type)
					if targetType := e.resolveType(typeStr, packageName); targetType != "" {
						toID := types.NewNodeID(packageName, targetType)
						dependencies = append(dependencies, DependencyInfo{
							From: fromID,
							To:   toID,
							Type: types.SignatureDependency,
						})
					}
				}
			}
			
			// Extract dependencies from receiver (methods)
			if node.Recv != nil {
				for _, field := range node.Recv.List {
					typeStr := extractTypeString(field.Type)
					if targetType := e.resolveType(typeStr, packageName); targetType != "" {
						toID := types.NewNodeID(packageName, targetType)
						dependencies = append(dependencies, DependencyInfo{
							From: fromID,
							To:   toID,
							Type: types.SignatureDependency,
						})
					}
				}
			}
		}
		return true
	})
	
	return dependencies, nil
}

// Name returns the strategy name
func (e *SignatureDependencyExtractor) Name() string {
	return "SignatureDependency"
}

// resolveType resolves a type string to a node name
func (e *SignatureDependencyExtractor) resolveType(typeStr string, currentPkg string) string {
	// Reuse the same logic from FieldDependencyExtractor
	fieldExtractor := &FieldDependencyExtractor{ctx: e.ctx}
	return fieldExtractor.resolveType(typeStr, currentPkg)
}