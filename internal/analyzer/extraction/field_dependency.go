package extraction

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/harakeishi/depsee/internal/types"
)

// FieldDependencyExtractor extracts dependencies from struct fields
type FieldDependencyExtractor struct {
	ctx *Context
}

// NewFieldDependencyExtractor creates a new field dependency extractor
func NewFieldDependencyExtractor(ctx *Context) *FieldDependencyExtractor {
	return &FieldDependencyExtractor{ctx: ctx}
}

// ExtractDependencies extracts field-based dependencies
func (e *FieldDependencyExtractor) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo
	
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := node.Type.(*ast.StructType); ok {
				structName := node.Name.Name
				fromID := types.NewNodeID(packageName, structName)
				
				for _, field := range structType.Fields.List {
					fieldType := extractTypeString(field.Type)
					if targetType := e.resolveType(fieldType, packageName); targetType != "" {
						toID := types.NewNodeID(packageName, targetType)
						dependencies = append(dependencies, DependencyInfo{
							From: fromID,
							To:   toID,
							Type: types.FieldDependency,
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
func (e *FieldDependencyExtractor) Name() string {
	return "FieldDependency"
}

// resolveType resolves a type string to a node name
func (e *FieldDependencyExtractor) resolveType(typeStr string, currentPkg string) string {
	// Remove pointer and slice indicators
	typeStr = strings.TrimPrefix(typeStr, "*")
	typeStr = strings.TrimPrefix(typeStr, "[]")
	typeStr = strings.TrimPrefix(typeStr, "[]*")
	
	// Skip basic types
	if isBasicType(typeStr) {
		return ""
	}
	
	// Handle package-qualified types
	if strings.Contains(typeStr, ".") {
		parts := strings.Split(typeStr, ".")
		if len(parts) == 2 {
			// Check if it's from an imported package
			if pkgPath, ok := e.ctx.ImportMap[parts[0]]; ok {
				// For now, we only handle same-package dependencies
				if pkgPath == currentPkg {
					return parts[1]
				}
			}
		}
		return ""
	}
	
	// Same package type
	return typeStr
}

// extractTypeString extracts type as string from ast.Expr
func extractTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + extractTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + extractTypeString(t.Elt)
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	}
	return ""
}

// isBasicType checks if a type is a basic Go type
func isBasicType(t string) bool {
	basicTypes := map[string]bool{
		"bool": true, "string": true, "error": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"byte": true, "rune": true,
		"float32": true, "float64": true,
		"complex64": true, "complex128": true,
		"interface{}": true, "any": true,
	}
	return basicTypes[t]
}