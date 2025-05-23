package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type AnalysisResult struct {
	Structs    []StructInfo
	Interfaces []InterfaceInfo
	Functions  []FuncInfo
}

// AnalyzeDir は指定ディレクトリ配下のGoファイルを再帰的に探索し、構造体・インターフェース・関数を抽出する
func AnalyzeDir(dir string) (*AnalysisResult, error) {
	var files []string
	// ディレクトリ再帰探索
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	result := &AnalysisResult{}

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		analyzeFile(f, fset, file, result)
	}

	return result, nil
}

// analyzeFile: ASTを走査し、構造体・インターフェース・関数・メソッドを抽出
func analyzeFile(f *ast.File, fset *token.FileSet, file string, result *AnalysisResult) {
	pkgName := f.Name.Name
	structMap := map[string]*StructInfo{}

	// パッケージのエイリアスマップを作成
	aliasMap := make(map[string]string)
	importMap := make(map[string]string)
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		if imp.Name != nil {
			// エイリアスがある場合
			aliasMap[imp.Name.Name] = path
		} else {
			// エイリアスがない場合は、パッケージ名を抽出
			parts := strings.Split(path, "/")
			pkg := parts[len(parts)-1]
			importMap[pkg] = path
		}
	}

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
			fi.BodyCalls = extractBodyCallsWithAlias(funcDecl.Body, aliasMap, importMap)
			result.Functions = append(result.Functions, fi)
		}
	}
	// 構造体メソッドにもBodyCallsを追加
	for _, s := range structMap {
		for i, m := range s.Methods {
			if m.Position.IsValid() {
				s.Methods[i].BodyCalls = extractBodyCallsWithAlias(findFuncDeclByName(f, m.Name), aliasMap, importMap)
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

// 関数本体から呼び出している関数名リストを抽出（エイリアス対応版）
func extractBodyCallsWithAlias(body *ast.BlockStmt, aliasMap, importMap map[string]string) []string {
	calls := []string{}
	if body == nil {
		return calls
	}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			// ローカル関数呼び出し
			calls = append(calls, fun.Name)
		case *ast.SelectorExpr:
			// パッケージ名を含む関数呼び出し
			if pkg, ok := fun.X.(*ast.Ident); ok {
				// エイリアスを解決
				if alias, ok := aliasMap[pkg.Name]; ok {
					// エイリアスが存在する場合は、パッケージパスを使用
					parts := strings.Split(alias, "/")
					pkgName := parts[len(parts)-1]
					calls = append(calls, pkgName+"."+fun.Sel.Name)
				} else if imp, ok := importMap[pkg.Name]; ok {
					// インポートマップに存在する場合
					parts := strings.Split(imp, "/")
					pkgName := parts[len(parts)-1]
					calls = append(calls, pkgName+"."+fun.Sel.Name)
				} else {
					// エイリアスもインポートマップにもない場合は、そのまま使用
					calls = append(calls, pkg.Name+"."+fun.Sel.Name)
				}
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

// AnalyzeDirWithOption は指定ディレクトリ配下のGoファイルを再帰的に探索し、
// withLocalImportsがtrueの場合は同一リポジトリ内のimport先パッケージも再帰的に解析する
func AnalyzeDirWithOption(dir string, withLocalImports bool) (*AnalysisResult, error) {
	result := &AnalysisResult{}
	processedDirs := make(map[string]bool)

	var analyzeDirRecursive func(dir string) error
	analyzeDirRecursive = func(dir string) error {
		if processedDirs[dir] {
			return nil
		}
		processedDirs[dir] = true

		// 現在のディレクトリを解析
		currentResult, err := AnalyzeDir(dir)
		if err != nil {
			return err
		}

		// 結果をマージ
		result.Structs = append(result.Structs, currentResult.Structs...)
		result.Interfaces = append(result.Interfaces, currentResult.Interfaces...)
		result.Functions = append(result.Functions, currentResult.Functions...)

		if !withLocalImports {
			return nil
		}

		// ローカルインポートの解析
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					impPath := strings.Trim(imp.Path.Value, "\"")
					// ローカルインポートの場合のみ処理
					if strings.HasPrefix(impPath, "./") || strings.HasPrefix(impPath, "../") {
						absPath := filepath.Join(dir, impPath)
						if err := analyzeDirRecursive(absPath); err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	}

	if err := analyzeDirRecursive(dir); err != nil {
		return nil, err
	}

	return result, nil
}
