package extraction

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/types"
	"github.com/harakeishi/depsee/internal/utils"
)

// PackageDependencyExtractor extracts package-level dependencies
type PackageDependencyExtractor struct {
	ctx       *Context
	targetDir string
}

// NewPackageDependencyExtractor creates a new package dependency extractor
func NewPackageDependencyExtractor(ctx *Context, targetDir string) *PackageDependencyExtractor {
	return &PackageDependencyExtractor{
		ctx:       ctx,
		targetDir: targetDir,
	}
}

// ExtractDependencies extracts package-level dependencies
func (e *PackageDependencyExtractor) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo
	
	// パッケージ間の依存関係を抽出
	// これは単一ファイルではなく、プロジェクト全体を解析する必要がある
	packageDeps, err := e.extractPackageDependencies()
	if err != nil {
		logger.Error("パッケージ依存関係抽出エラー", "error", err)
		return dependencies, err
	}
	
	// 現在のパッケージから他のパッケージへの依存関係を作成
	for _, dep := range packageDeps {
		if dep.From == packageName {
			fromID := types.NewPackageNodeID(dep.From)
			toID := types.NewPackageNodeID(dep.To)
			dependencies = append(dependencies, DependencyInfo{
				From: fromID,
				To:   toID,
				Type: types.PackageDependency,
			})
		}
	}
	
	return dependencies, nil
}

// Name returns the name of the strategy
func (e *PackageDependencyExtractor) Name() string {
	return "PackageDependency"
}

// PackageDep represents a package dependency
type PackageDep struct {
	From string
	To   string
}

// extractPackageDependencies extracts dependencies between packages
func (e *PackageDependencyExtractor) extractPackageDependencies() ([]PackageDep, error) {
	var dependencies []PackageDep
	packageImports := make(map[string][]string)
	
	// プロジェクト内のすべてのGoファイルを走査
	err := filepath.Walk(e.targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// ファイルを解析してインポートを抽出
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			logger.Debug("ファイル解析エラー", "path", path, "error", err)
			return nil
		}
		
		packageName := file.Name.Name
		var imports []string
		
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if utils.IsLocalPackage(importPath) {
				// ローカルパッケージの場合、パッケージ名を抽出
				if pkgName := utils.ExtractPackageName(importPath); pkgName != "" {
					imports = append(imports, pkgName)
				}
			}
		}
		
		if len(imports) > 0 {
			packageImports[packageName] = imports
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// パッケージ間依存関係を構築
	for fromPkg, imports := range packageImports {
		for _, toPkg := range imports {
			dependencies = append(dependencies, PackageDep{
				From: fromPkg,
				To:   toPkg,
			})
		}
	}
	
	return dependencies, nil
}
