package analyzer

import (
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/types"
	"github.com/harakeishi/depsee/internal/utils"
)

// DependencyInfo は依存関係情報を表す構造体です。
// 依存元ノード、依存先ノード、依存関係の種類を定義します。
type DependencyInfo struct {
	From types.NodeID         // 依存元のノードID
	To   types.NodeID         // 依存先のノードID
	Type types.DependencyType // 依存関係の種類
}

// DependencyExtractor は依存関係抽出の戦略インターフェースです。
// Strategyパターンを使用して、異なる種類の依存関係抽出ロジックを実装します。
type DependencyExtractor interface {
	// Extract は解析結果から特定の種類の依存関係を抽出します
	Extract(result *Result) []DependencyInfo
}

// FieldDependencyExtractor は構造体フィールドの依存関係を抽出する実装です。
// 構造体のフィールドが他の型を参照している場合の依存関係を検出します。
type FieldDependencyExtractor struct {
	typeResolver *TypeResolver // 型情報の解決に使用するリゾルバ
}

// NewFieldDependencyExtractor は新しいFieldDependencyExtractorを作成します。
// 型解決に使用するTypeResolverを指定してインスタンスを初期化します。
func NewFieldDependencyExtractor(typeResolver *TypeResolver) *FieldDependencyExtractor {
	return &FieldDependencyExtractor{
		typeResolver: typeResolver,
	}
}

// Extract は構造体のフィールドによる依存関係を抽出します。
// 各構造体のフィールドを調べ、他の型を参照しているフィールドがある場合に
// 依存関係情報を作成します。
func (e *FieldDependencyExtractor) Extract(result *Result) []DependencyInfo {
	logger.Debug("フィールド依存関係抽出開始")
	var dependencies []DependencyInfo

	// ノードマップを作成（依存先の存在確認用）
	nodeMap := result.CreateNodeMap()

	for _, s := range result.Structs {
		fromID := types.NewNodeID(s.Package, s.Name)
		logger.Debug("構造体フィールド解析", "struct", s.Name, "package", s.Package)

		for _, field := range s.Fields {
			if toID := e.parseTypeToNodeID(field.Type, s.Package); toID != "" {
				if _, ok := nodeMap[toID]; ok {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   toID,
						Type: types.FieldDependency,
					})
					logger.Debug("フィールド依存関係追加", "from", fromID, "to", toID, "field", field.Name)
				} else {
					logger.Debug("依存先ノード未発見", "from", fromID, "to", toID, "field", field.Name)
				}
			}
		}
	}

	return dependencies
}

// parseTypeToNodeID は型名からノードIDを生成します。
// ポインタやスライスのプレフィックスを取り除き、パッケージ名と結合して
// 一意のノードIDを生成します。
func (e *FieldDependencyExtractor) parseTypeToNodeID(typeName, pkg string) types.NodeID {
	// より安全な型解析ロジック
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return types.NewNodeID(pkg, cleaned)
}


// SignatureDependencyExtractor は関数シグネチャの依存関係を抽出する実装です。
// 関数やメソッドの引数や戻り値の型が他の型を参照している場合の
// 依存関係を検出します。
type SignatureDependencyExtractor struct{}

// Extract は関数・メソッドのシグネチャによる依存関係を抽出します。
// 引数や戻り値の型を調べ、他の型を参照している場合に
// 依存関係情報を作成します。
func (e *SignatureDependencyExtractor) Extract(result *Result) []DependencyInfo {
	var dependencies []DependencyInfo
	nodeMap := result.CreateNodeMap()

	// 関数の引数・戻り値の依存関係抽出
	for _, f := range result.Functions {
		fromID := types.NewNodeID(f.Package, f.Name)
		dependencies = append(dependencies, e.extractFromParams(f.Params, fromID, f.Package, nodeMap)...)
		dependencies = append(dependencies, e.extractFromParams(f.Results, fromID, f.Package, nodeMap)...)
	}

	// メソッドの引数・戻り値の依存関係抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := types.NewNodeID(s.Package, m.Name)
			dependencies = append(dependencies, e.extractFromParams(m.Params, fromID, s.Package, nodeMap)...)
			dependencies = append(dependencies, e.extractFromParams(m.Results, fromID, s.Package, nodeMap)...)
		}
	}

	return dependencies
}

// extractFromParams はパラメータ一覧から依存関係を抽出します。
// 各パラメータの型を解析し、他のノードへの依存関係を検出します。
func (e *SignatureDependencyExtractor) extractFromParams(params []FieldInfo, fromID types.NodeID, pkg string, nodeMap map[types.NodeID]struct{}) []DependencyInfo {
	var dependencies []DependencyInfo
	for _, param := range params {
		if toID := e.parseTypeToNodeID(param.Type, pkg); toID != "" {
			if _, ok := nodeMap[toID]; ok {
				dependencies = append(dependencies, DependencyInfo{
					From: fromID,
					To:   toID,
					Type: types.SignatureDependency,
				})
			}
		}
	}
	return dependencies
}

// parseTypeToNodeID は型名からノードIDを生成します。
// FieldDependencyExtractorの同名メソッドと同じロジックで型名を解析します。
func (e *SignatureDependencyExtractor) parseTypeToNodeID(typeName, pkg string) types.NodeID {
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return types.NewNodeID(pkg, cleaned)
}


// BodyCallDependencyExtractor は関数本体の呼び出し依存関係を抽出する実装です。
// 関数やメソッドの本体内で他の関数や構造体を呼び出している場合の
// 依存関係を検出します。
type BodyCallDependencyExtractor struct{}

// Extract は関数・メソッド本体の呼び出しによる依存関係を抽出します。
// 関数やメソッドの本体で使用されている他の関数や型への呼び出しを
// 検出して依存関係情報を作成します。
func (e *BodyCallDependencyExtractor) Extract(result *Result) []DependencyInfo {
	var dependencies []DependencyInfo
	nodeMap := result.CreateNodeMap()

	// 関数本体の依存関係抽出
	for _, f := range result.Functions {
		fromID := types.NewNodeID(f.Package, f.Name)
		for _, called := range f.BodyCalls {
			toID := types.NewNodeID(f.Package, called)
			if _, ok := nodeMap[toID]; ok {
				dependencies = append(dependencies, DependencyInfo{
					From: fromID,
					To:   toID,
					Type: types.BodyCallDependency,
				})
			}
		}
	}

	// メソッド本体の依存関係抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := types.NewNodeID(s.Package, m.Name)
			for _, called := range m.BodyCalls {
				toID := types.NewNodeID(s.Package, called)
				if _, ok := nodeMap[toID]; ok {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   toID,
						Type: types.BodyCallDependency,
					})
				}
			}
		}
	}

	return dependencies
}


// PackageDependencyExtractor はパッケージ間の依存関係を抽出する実装です。
// import文に基づいてパッケージ間の依存関係を検出します。
// 同リポジトリ内のパッケージ間依存のみを対象とします。
type PackageDependencyExtractor struct {
	targetDir string // 解析対象のルートディレクトリパス
}

// NewPackageDependencyExtractor は新しいPackageDependencyExtractorを作成します。
// 解析対象のルートディレクトリを指定してインスタンスを初期化します。
func NewPackageDependencyExtractor(targetDir string) *PackageDependencyExtractor {
	return &PackageDependencyExtractor{
		targetDir: targetDir,
	}
}

// Extract はimport文に基づいたパッケージ間依存関係を抽出します。
// 各パッケージのimport文を調べ、同リポジトリ内の他パッケージを参照している場合に
// パッケージ間依存関係情報を作成します。
func (e *PackageDependencyExtractor) Extract(result *Result) []DependencyInfo {
	logger.Debug("パッケージ間依存関係抽出開始")
	var dependencies []DependencyInfo

	// パッケージノードマップを作成
	packageNodes := make(map[string]struct{})
	for _, pkg := range result.Packages {
		packageNodes[pkg.Name] = struct{}{}
		logger.Debug("パッケージノード追加", "package", pkg.Name)
	}

	// パッケージ間の依存関係を抽出
	for _, pkg := range result.Packages {
		fromID := types.NewPackageNodeID(pkg.Name)

		for _, imp := range pkg.Imports {
			// 同リポジトリ内のパッケージかどうかを判定
			if utils.IsLocalPackage(imp.Path) {
				// パッケージ名を抽出（パスの最後の部分）
				targetPkgName := utils.ExtractPackageName(imp.Path)
				toID := types.NewPackageNodeID(targetPkgName)

				// 依存先パッケージが存在する場合のみエッジを追加
				if _, ok := packageNodes[targetPkgName]; ok {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   toID,
						Type: types.PackageDependency,
					})
					logger.Debug("パッケージ間依存関係追加", "from", fromID, "to", toID, "import", imp.Path)
				}
			}
		}
	}

	return dependencies
}

// CrossPackageDependencyExtractor はパッケージ間の関数呼び出しや型の使用を抽出する実装です。
// import文で取り込まれたパッケージの関数や型を使用している箇所を検出し、
// パッケージ間の具体的な依存関係を分析します。
type CrossPackageDependencyExtractor struct {
	packageMap map[string]string // importエイリアスからパッケージ名へのマッピング（未使用）
}

// NewCrossPackageDependencyExtractor は新しいCrossPackageDependencyExtractorを作成します。
// パッケージ間の関数呼び出しや型使用の分析に必要な内部状態を初期化します。
func NewCrossPackageDependencyExtractor() *CrossPackageDependencyExtractor {
	return &CrossPackageDependencyExtractor{
		packageMap: make(map[string]string),
	}
}

// Extract はパッケージ間の関数呼び出しや型使用による依存関係を抽出します。
// 関数やメソッドの本体で使用されている他パッケージの関数や型を検出し、
// 具体的なパッケージ間依存関係情報を作成します。
func (e *CrossPackageDependencyExtractor) Extract(result *Result) []DependencyInfo {
	logger.Debug("パッケージ間関数呼び出し依存関係抽出開始")
	var dependencies []DependencyInfo
	nodeMap := result.CreateNodeMap()

	// パッケージごとのimport情報を構築（同じパッケージの複数ファイルをマージ）
	packageImports := make(map[string]map[string]string) // package -> (alias -> import_path)
	for _, pkg := range result.Packages {
		if packageImports[pkg.Name] == nil {
			packageImports[pkg.Name] = make(map[string]string)
		}
		for _, imp := range pkg.Imports {
			alias := e.extractPackageAlias(imp.Path, imp.Alias)
			packageImports[pkg.Name][alias] = imp.Path
		}
	}

	// 関数の本体から他パッケージの関数呼び出しを抽出
	for _, f := range result.Functions {
		fromID := types.NewNodeID(f.Package, f.Name)
		dependencies = append(dependencies, e.extractCrossPackageCalls(f.BodyCalls, f.Package, packageImports, fromID, nodeMap)...)
	}

	// メソッドの本体から他パッケージの関数呼び出しを抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := types.NewNodeID(s.Package, m.Name)
			dependencies = append(dependencies, e.extractCrossPackageCalls(m.BodyCalls, s.Package, packageImports, fromID, nodeMap)...)
		}
	}

	return dependencies
}

// extractCrossPackageCalls は関数本体の呼び出しからパッケージ間依存を抽出します。
// パッケージ修飾子付きの呼び出しを検出し、import情報と照合して
// 同リポジトリ内の他パッケージへの依存関係を特定します。
func (e *CrossPackageDependencyExtractor) extractCrossPackageCalls(bodyCalls []string, currentPkg string, packageImports map[string]map[string]string, fromID types.NodeID, nodeMap map[types.NodeID]struct{}) []DependencyInfo {
	var dependencies []DependencyInfo
	imports := packageImports[currentPkg]
	if imports == nil {
		return dependencies
	}

	for _, call := range bodyCalls {
		// パッケージ修飾子付きの呼び出しを検出（例：depsee.New, depsee.Config）
		if strings.Contains(call, ".") {
			parts := strings.Split(call, ".")
			if len(parts) >= 2 {
				pkgAlias := parts[0]
				funcOrTypeName := parts[1]

				// import情報からパッケージパスを取得
				if importPath, ok := imports[pkgAlias]; ok {
					// 同リポジトリ内のパッケージかどうかを判定
					if utils.IsLocalPackage(importPath) {
						targetPkgName := utils.ExtractPackageName(importPath)
						toID := types.NewNodeID(targetPkgName, funcOrTypeName)

						// 依存先ノードが存在する場合のみエッジを追加
						if _, ok := nodeMap[toID]; ok {
							dependencies = append(dependencies, DependencyInfo{
								From: fromID,
								To:   toID,
								Type: types.CrossPackageDependency,
							})
							logger.Debug("パッケージ間関数呼び出し依存関係追加", "from", fromID, "to", toID, "call", call)
						} else {
							logger.Debug("パッケージ間依存先ノード未発見", "from", fromID, "to", toID, "call", call)
						}
					}
				}
			}
		}
	}

	return dependencies
}

// extractPackageAlias はimport文からパッケージのエイリアスを抽出します。
// エイリアスが明示的に指定されている場合はそれを使用し、
// 指定されていない場合はimportパスからパッケージ名を抽出します。
func (e *CrossPackageDependencyExtractor) extractPackageAlias(importPath, alias string) string {
	if alias != "" {
		return alias // "."や"_"も含めて、指定されたエイリアスをそのまま返す
	}
	// エイリアスがない場合はパッケージ名を使用
	return utils.ExtractPackageName(importPath)
}

