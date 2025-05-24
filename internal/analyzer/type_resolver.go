package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
)

// TypeResolver は型情報を解決するための構造体
type TypeResolver struct {
	fset     *token.FileSet
	packages map[string]*types.Package
	info     *types.Info
}

// NewTypeResolver は新しいTypeResolverを作成
func NewTypeResolver() *TypeResolver {
	return &TypeResolver{
		fset:     token.NewFileSet(),
		packages: make(map[string]*types.Package),
		info: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		},
	}
}

// ResolvePackage はパッケージの型情報を解決
func (tr *TypeResolver) ResolvePackage(dir string) error {
	logger.Debug("パッケージ型解析開始", "dir", dir)

	// パッケージ内のGoファイルをパース
	pkgs, err := parser.ParseDir(tr.fset, dir, nil, parser.ParseComments)
	if err != nil {
		logger.Error("パッケージパース失敗", "dir", dir, "error", err)
		return err
	}

	for pkgName, pkg := range pkgs {
		if strings.HasSuffix(pkgName, "_test") {
			continue // テストパッケージはスキップ
		}

		logger.Debug("パッケージ型チェック開始", "package", pkgName)

		// ASTファイルのスライスを作成
		var files []*ast.File
		for _, file := range pkg.Files {
			files = append(files, file)
		}

		// 型チェッカーの設定
		conf := types.Config{
			Importer: importer.ForCompiler(tr.fset, "source", nil),
			Error: func(err error) {
				logger.Warn("型チェックエラー", "error", err)
			},
		}

		// 型チェック実行
		typePkg, err := conf.Check(pkgName, tr.fset, files, tr.info)
		if err != nil {
			logger.Warn("型チェック部分失敗", "package", pkgName, "error", err)
			// エラーがあっても部分的な情報は使用可能
		}

		if typePkg != nil {
			tr.packages[pkgName] = typePkg
			logger.Debug("パッケージ型情報取得完了", "package", pkgName)
		}
	}

	return nil
}

// ResolveType は型表現から正確な型名を解決
func (tr *TypeResolver) ResolveType(expr ast.Expr) string {
	if tr.info == nil {
		return tr.fallbackTypeResolution(expr)
	}

	if typeAndValue, ok := tr.info.Types[expr]; ok {
		return tr.formatType(typeAndValue.Type)
	}

	return tr.fallbackTypeResolution(expr)
}

// formatType は型情報を文字列に変換
func (tr *TypeResolver) formatType(t types.Type) string {
	switch typ := t.(type) {
	case *types.Named:
		obj := typ.Obj()
		if obj.Pkg() != nil {
			return obj.Pkg().Name() + "." + obj.Name()
		}
		return obj.Name()
	case *types.Pointer:
		return "*" + tr.formatType(typ.Elem())
	case *types.Slice:
		return "[]" + tr.formatType(typ.Elem())
	case *types.Array:
		return "[" + string(rune(typ.Len())) + "]" + tr.formatType(typ.Elem())
	case *types.Map:
		return "map[" + tr.formatType(typ.Key()) + "]" + tr.formatType(typ.Elem())
	case *types.Interface:
		if typ.Empty() {
			return "interface{}"
		}
		return "interface{...}"
	case *types.Basic:
		return typ.Name()
	default:
		return t.String()
	}
}

// fallbackTypeResolution はgo/typesが使用できない場合のフォールバック
func (tr *TypeResolver) fallbackTypeResolution(expr ast.Expr) string {
	return exprToTypeString(expr)
}

// GetPackageInfo はパッケージ情報を取得
func (tr *TypeResolver) GetPackageInfo(pkgName string) *types.Package {
	return tr.packages[pkgName]
}

// IsExternalType は外部パッケージの型かどうかを判定
func (tr *TypeResolver) IsExternalType(typeName string) bool {
	return strings.Contains(typeName, ".")
}

// ExtractPackageName は型名からパッケージ名を抽出
func (tr *TypeResolver) ExtractPackageName(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[:idx]
	}
	return ""
}

// ExtractTypeName は型名から型名部分のみを抽出
func (tr *TypeResolver) ExtractTypeName(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}
