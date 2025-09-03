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

// TypeResolver は型情報を解決するための構造体です。
// Go言語の型チェッカーを使用して正確な型情報を取得し、
// 型名の解決やパッケージ情報の管理を行います。
type TypeResolver struct {
	fset     *token.FileSet             // ファイル位置情報の管理
	packages map[string]*types.Package  // パッケージ名から型情報へのマッピング
	info     *types.Info               // 型チェック結果の詳細情報
}

// NewTypeResolver は新しいTypeResolverを作成します。
// 型解決に必要な内部構造を初期化します。
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

// ResolvePackage は指定されたディレクトリのパッケージの型情報を解決します。
// パッケージ内のGoファイルをパースし、型チェッカーで型情報を取得します。
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

// ResolveType は型表現から正確な型名を解決します。
// 型チェック情報が利用可能な場合はそれを使用し、
// 利用できない場合はフォールバック処理を行います。
func (tr *TypeResolver) ResolveType(expr ast.Expr) string {
	if tr.info == nil {
		return tr.fallbackTypeResolution(expr)
	}

	if typeAndValue, ok := tr.info.Types[expr]; ok {
		return tr.formatType(typeAndValue.Type)
	}

	return tr.fallbackTypeResolution(expr)
}

// formatType は型情報を文字列に変換します。
// 名前付き型、ポインタ、スライス、配列、マップ、インターフェース、基本型を
// 適切な文字列表現に変換します。
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

// fallbackTypeResolution はgo/typesが使用できない場合のフォールバック処理です。
// 型チェック情報が利用できない場合に、ASTから直接型情報を抽出します。
func (tr *TypeResolver) fallbackTypeResolution(expr ast.Expr) string {
	return exprToTypeString(expr)
}

// GetPackageInfo は指定されたパッケージ名のパッケージ情報を取得します。
// 型解決済みのパッケージ情報が存在する場合はそれを返し、
// 存在しない場合はnilを返します。
func (tr *TypeResolver) GetPackageInfo(pkgName string) *types.Package {
	return tr.packages[pkgName]
}

// IsExternalType は指定された型名が外部パッケージの型かどうかを判定します。
// 型名にパッケージ修飾子（ドット）が含まれている場合はtrueを返します。
func (tr *TypeResolver) IsExternalType(typeName string) bool {
	return strings.Contains(typeName, ".")
}

// ExtractPackageName は型名からパッケージ名を抽出します。
// "package.Type"形式の型名から"package"部分を取得します。
// パッケージ修飾子がない場合は空文字を返します。
func (tr *TypeResolver) ExtractPackageName(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[:idx]
	}
	return ""
}

// ExtractTypeName は型名から型名部分のみを抽出します。
// "package.Type"形式の型名から"Type"部分を取得します。
// パッケージ修飾子がない場合は元の型名をそのまま返します。
func (tr *TypeResolver) ExtractTypeName(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}
