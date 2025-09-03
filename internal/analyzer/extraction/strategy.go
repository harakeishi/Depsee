package extraction

import (
	"go/ast"
	"go/token"

	"github.com/harakeishi/depsee/internal/types"
)

// DependencyInfo represents a dependency between two nodes
type DependencyInfo = types.DependencyInfo

// ExtractionStrategy defines the interface for dependency extraction strategies
type ExtractionStrategy interface {
	// ExtractDependencies extracts dependencies from the AST
	ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error)
	
	// Name returns the name of the strategy
	Name() string
}

// Context holds shared state for extraction strategies
type Context struct {
	FileSet     *token.FileSet
	PackageName string
	ImportMap   map[string]string // maps import aliases to package paths
}

// NewContext creates a new extraction context
func NewContext(fset *token.FileSet, packageName string) *Context {
	return &Context{
		FileSet:     fset,
		PackageName: packageName,
		ImportMap:   make(map[string]string),
	}
}