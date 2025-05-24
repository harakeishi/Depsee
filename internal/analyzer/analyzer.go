package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/harakeishi/depsee/internal/errors"
	"github.com/harakeishi/depsee/internal/logger"
)

type AnalysisResult struct {
	Structs    []StructInfo
	Interfaces []InterfaceInfo
	Functions  []FuncInfo
	Packages   []PackageInfo
}

// GoAnalyzer はGo言語の静的解析を行う具象実装
type GoAnalyzer struct {
	// 将来的に設定やオプションを追加可能
}

// New は新しいAnalyzerを作成
func New() Analyzer {
	return &GoAnalyzer{}
}

// NewGoAnalyzer は新しいGoAnalyzerを作成
func NewGoAnalyzer() Analyzer {
	return &GoAnalyzer{}
}

// AnalyzeDir は指定ディレクトリ配下のGoファイルを再帰的に探索し、構造体・インターフェース・関数を抽出する
func (ga *GoAnalyzer) AnalyzeDir(dir string) (*AnalysisResult, error) {
	logger.Debug("ディレクトリ解析開始", "dir", dir)

	// ディレクトリの存在確認
	if _, err := os.Stat(dir); err != nil {
		logger.Error("ディレクトリが存在しません", "dir", dir, "error", err)
		return nil, errors.NewAnalysisError(dir, err)
	}

	var files []string
	errorCollector := errors.NewErrorCollector()

	// ディレクトリ再帰探索
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Warn("ファイル読み込みエラー", "path", path, "error", err)
			errorCollector.Add(errors.NewAnalysisError(path, err))
			return nil // エラーを収集して処理を続行
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
			logger.Debug("Goファイル発見", "file", path)
		}
		return nil
	})
	if err != nil {
		logger.Error("ディレクトリ探索失敗", "dir", dir, "error", err)
		return nil, errors.NewAnalysisError(dir, err)
	}

	logger.Info("Goファイル発見完了", "count", len(files))

	fset := token.NewFileSet()
	result := &AnalysisResult{}

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			logger.Warn("ファイルパース失敗", "file", file, "error", err)
			errorCollector.Add(errors.NewAnalysisError(file, err))
			continue // パースエラーがあっても他のファイルは処理を続行
		}
		analyzeFile(f, fset, file, result)
	}

	// 重大なエラーがある場合は失敗とする
	if errorCollector.HasErrors() {
		logger.Warn("解析中にエラーが発生しました", "error_count", len(errorCollector.Errors()))
		// 部分的な結果でも返す（警告として扱う）
	}

	return result, nil
}

// 後方互換性のためのラッパー関数
func AnalyzeDir(dir string) (*AnalysisResult, error) {
	analyzer := NewGoAnalyzer()
	return analyzer.AnalyzeDir(dir)
}

// AnalyzeDirWithPackageFilter は指定されたパッケージのみを解析対象とする
func (ga *GoAnalyzer) AnalyzeDirWithPackageFilter(dir string, targetPackages []string) (*AnalysisResult, error) {
	logger.Debug("パッケージフィルタ付きディレクトリ解析開始", "dir", dir, "target_packages", targetPackages)

	// 全体解析を実行
	result, err := ga.AnalyzeDir(dir)
	if err != nil {
		return nil, err
	}

	// パッケージフィルタリングが指定されていない場合は全結果を返す
	if len(targetPackages) == 0 {
		return result, nil
	}

	// 対象パッケージのセットを作成
	targetSet := make(map[string]bool)
	for _, pkg := range targetPackages {
		targetSet[pkg] = true
	}

	// フィルタリング実行
	filteredResult := &AnalysisResult{}

	// 構造体のフィルタリング
	for _, s := range result.Structs {
		if targetSet[s.Package] {
			filteredResult.Structs = append(filteredResult.Structs, s)
		}
	}

	// インターフェースのフィルタリング
	for _, i := range result.Interfaces {
		if targetSet[i.Package] {
			filteredResult.Interfaces = append(filteredResult.Interfaces, i)
		}
	}

	// 関数のフィルタリング
	for _, f := range result.Functions {
		if targetSet[f.Package] {
			filteredResult.Functions = append(filteredResult.Functions, f)
		}
	}

	// パッケージのフィルタリング
	for _, p := range result.Packages {
		if targetSet[p.Name] {
			filteredResult.Packages = append(filteredResult.Packages, p)
		}
	}

	logger.Info("パッケージフィルタリング完了",
		"original_structs", len(result.Structs), "filtered_structs", len(filteredResult.Structs),
		"original_interfaces", len(result.Interfaces), "filtered_interfaces", len(filteredResult.Interfaces),
		"original_functions", len(result.Functions), "filtered_functions", len(filteredResult.Functions),
		"original_packages", len(result.Packages), "filtered_packages", len(filteredResult.Packages))

	return filteredResult, nil
}

// analyzeFile: ASTを走査し、構造体・インターフェース・関数・メソッドを抽出
func analyzeFile(f *ast.File, fset *token.FileSet, file string, result *AnalysisResult) {
	pkgName := f.Name.Name
	structMap := map[string]*StructInfo{}

	// 0th pass: import文の解析
	imports := []ImportInfo{}
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			// エイリアスが指定されていない場合、importパスから自動的にパッケージ名を抽出
			parts := strings.Split(importPath, "/")
			if len(parts) > 0 {
				alias = parts[len(parts)-1]
			} else {
				alias = importPath // フォールバック
			}
		}
		imports = append(imports, ImportInfo{
			Path:  importPath,
			Alias: alias,
		})
	}

	// パッケージ情報を追加
	pos := fset.Position(f.Name.Pos())
	packageInfo := PackageInfo{
		Name:     pkgName,
		Path:     "", // TODO: パッケージパスの取得
		File:     file,
		Position: pos,
		Imports:  imports,
	}
	result.Packages = append(result.Packages, packageInfo)

	// 1st pass: type宣言（構造体・インターフェース）
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			pos := fset.Position(typeSpec.Pos())
			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				fields := []FieldInfo{}
				for _, field := range t.Fields.List {
					// フィールド名（複数可）
					for _, name := range field.Names {
						fields = append(fields, FieldInfo{
							Name: name.Name,
							Type: exprToTypeString(field.Type),
						})
					}
					// 無名フィールド（埋め込み）
					if len(field.Names) == 0 {
						fields = append(fields, FieldInfo{
							Name: "",
							Type: exprToTypeString(field.Type),
						})
					}
				}
				si := StructInfo{
					Name:     typeSpec.Name.Name,
					Package:  pkgName,
					File:     file,
					Position: pos,
					Fields:   fields,
				}
				structMap[si.Name] = &si
				result.Structs = append(result.Structs, si)
			case *ast.InterfaceType:
				ii := InterfaceInfo{
					Name:     typeSpec.Name.Name,
					Package:  pkgName,
					File:     file,
					Position: pos,
				}
				result.Interfaces = append(result.Interfaces, ii)
			}
		}
	}

	// 2nd pass: 関数・メソッド
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		pos := fset.Position(funcDecl.Pos())
		params := []FieldInfo{}
		if funcDecl.Type.Params != nil {
			for _, field := range funcDecl.Type.Params.List {
				typeStr := exprToTypeString(field.Type)
				for _, name := range field.Names {
					params = append(params, FieldInfo{Name: name.Name, Type: typeStr})
				}
				if len(field.Names) == 0 {
					params = append(params, FieldInfo{Name: "", Type: typeStr})
				}
			}
		}
		results := []FieldInfo{}
		if funcDecl.Type.Results != nil {
			for _, field := range funcDecl.Type.Results.List {
				typeStr := exprToTypeString(field.Type)
				for _, name := range field.Names {
					results = append(results, FieldInfo{Name: name.Name, Type: typeStr})
				}
				if len(field.Names) == 0 {
					results = append(results, FieldInfo{Name: "", Type: typeStr})
				}
			}
		}
		fi := FuncInfo{
			Name:     funcDecl.Name.Name,
			Package:  pkgName,
			File:     file,
			Position: pos,
			Params:   params,
			Results:  results,
		}
		// メソッドの場合はStructInfoに内包
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			recvType := ""
			switch t := funcDecl.Recv.List[0].Type.(type) {
			case *ast.Ident:
				recvType = t.Name
			case *ast.StarExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					recvType = ident.Name
				}
			}
			fi.Receiver = recvType
			if s, ok := structMap[recvType]; ok {
				s.Methods = append(s.Methods, fi)
				// 構造体リストも更新
				for i := range result.Structs {
					if result.Structs[i].Name == recvType {
						result.Structs[i] = *s
					}
				}
			}
		} else {
			// 通常の関数
			// --- 関数本体の呼び出し関数名抽出 ---
			fi.BodyCalls = extractBodyCalls(funcDecl.Body)
			result.Functions = append(result.Functions, fi)
		}
	}
	// 構造体メソッドにもBodyCallsを追加
	for _, s := range structMap {
		for i, m := range s.Methods {
			if m.Position.IsValid() {
				s.Methods[i].BodyCalls = extractBodyCalls(findFuncDeclByName(f, m.Name))
			}
		}
	}
}

// 型表現を文字列化するユーティリティ
func exprToTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToTypeString(t.X)
	case *ast.SelectorExpr:
		return exprToTypeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToTypeString(t.Key) + "]" + exprToTypeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// 関数本体から呼び出している関数名リストを抽出
func extractBodyCalls(body *ast.BlockStmt) []string {
	calls := []string{}
	if body == nil {
		return calls
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// 関数呼び出し
			switch fun := node.Fun.(type) {
			case *ast.Ident:
				// 同一パッケージ内の関数呼び出し（例：New）
				calls = append(calls, fun.Name)
			case *ast.SelectorExpr:
				// パッケージ修飾子付きの関数呼び出し（例：depsee.New）
				if ident, ok := fun.X.(*ast.Ident); ok {
					// パッケージ名.関数名の形式で保存
					calls = append(calls, ident.Name+"."+fun.Sel.Name)
				} else {
					// その他のセレクタ（例：obj.Method()）
					calls = append(calls, fun.Sel.Name)
				}
			}
		case *ast.CompositeLit:
			// 構造体リテラル（例：depsee.Config{...}）
			if selectorExpr, ok := node.Type.(*ast.SelectorExpr); ok {
				if ident, ok := selectorExpr.X.(*ast.Ident); ok {
					// パッケージ名.型名の形式で保存
					calls = append(calls, ident.Name+"."+selectorExpr.Sel.Name)
				}
			} else if ident, ok := node.Type.(*ast.Ident); ok {
				// 同一パッケージ内の型（例：Config{...}）
				calls = append(calls, ident.Name)
			}
		}
		return true
	})
	return calls
}

// 名前からFuncDeclを探す（同一ファイル内のみ）
func findFuncDeclByName(f *ast.File, name string) *ast.BlockStmt {
	for _, decl := range f.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == name {
				return funcDecl.Body
			}
		}
	}
	return nil
}
