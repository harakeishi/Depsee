package extraction

import (
	"go/token"
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/types"
	"github.com/harakeishi/depsee/internal/utils"
)

// LegacyResult は解析結果を格納する構造体です（analyzer.Resultの複製）
type LegacyResult struct {
	Structs      []LegacyStructInfo     // 抽出された構造体の一覧
	Interfaces   []LegacyInterfaceInfo  // 抽出されたインターフェースの一覧
	Functions    []LegacyFuncInfo       // 抽出された関数の一覧
	Packages     []LegacyPackageInfo    // 抽出されたパッケージの一覧
	Dependencies []LegacyDependencyInfo // 抽出された依存関係の一覧
}

// LegacyStructInfo は構造体の情報を表します（analyzer.StructInfoの複製）
type LegacyStructInfo struct {
	Name     string              // 構造体名
	Package  string              // 所属パッケージ名
	File     string              // 定義されているファイルパス
	Position token.Position      // ファイル内での位置情報
	Fields   []LegacyFieldInfo   // 構造体が持つフィールドの一覧
	Methods  []LegacyFuncInfo    // 構造体に関連付けられたメソッドの一覧
}

// LegacyInterfaceInfo はインターフェースの情報を表します（analyzer.InterfaceInfoの複製）
type LegacyInterfaceInfo struct {
	Name     string              // インターフェース名
	Package  string              // 所属パッケージ名
	File     string              // 定義されているファイルパス
	Position token.Position      // ファイル内での位置情報
	Methods  []LegacyFuncInfo    // インターフェースで定義されているメソッドの一覧
}

// LegacyFuncInfo は関数・メソッドの情報を表します（analyzer.FuncInfoの複製）
type LegacyFuncInfo struct {
	Name      string              // 関数・メソッド名
	Package   string              // 所属パッケージ名
	File      string              // 定義されているファイルパス
	Position  token.Position      // ファイル内での位置情報
	Receiver  string              // レシーバ型名（メソッドの場合のみ設定される）
	Params    []LegacyFieldInfo   // 引数の一覧
	Results   []LegacyFieldInfo   // 戻り値の一覧
	BodyCalls []string            // 関数本体内で呼び出している関数名の一覧
}

// LegacyFieldInfo はフィールドの情報を表します（analyzer.FieldInfoの複製）
type LegacyFieldInfo struct {
	Name string // フィールド名
	Type string // フィールドの型名
}

// LegacyPackageInfo はパッケージの情報を表します（analyzer.PackageInfoの複製）
type LegacyPackageInfo struct {
	Name    string              // パッケージ名
	Path    string              // パッケージのパス
	Files   []string            // パッケージに含まれるファイルの一覧
	Imports []LegacyImportInfo  // パッケージのimport情報の一覧
}

// LegacyImportInfo はimport情報を表します（analyzer.ImportInfoの複製）
type LegacyImportInfo struct {
	Path  string // importするパッケージのパス
	Alias string // importエイリアス（省略可能）
}

// LegacyTypeResolver は型情報の解決を行います（analyzer.TypeResolverの複製）
type LegacyTypeResolver struct {
	// 実装は必要に応じて追加
}

// LegacyDependencyInfo は依存関係情報を表す構造体です。
// 依存元ノード、依存先ノード、依存関係の種類を定義します。
type LegacyDependencyInfo struct {
	From types.NodeID         // 依存元のノードID
	To   types.NodeID         // 依存先のノードID
	Type types.DependencyType // 依存関係の種類
}

// CreateNodeMap creates a map for checking node existence
func (r *LegacyResult) CreateNodeMap() map[types.NodeID]struct{} {
	nodeMap := make(map[types.NodeID]struct{})
	
	// Add struct nodes
	for _, s := range r.Structs {
		nodeID := types.NewNodeID(s.Package, s.Name)
		nodeMap[nodeID] = struct{}{}
	}
	
	// Add interface nodes
	for _, i := range r.Interfaces {
		nodeID := types.NewNodeID(i.Package, i.Name)
		nodeMap[nodeID] = struct{}{}
	}
	
	// Add function nodes
	for _, f := range r.Functions {
		nodeID := types.NewNodeID(f.Package, f.Name)
		nodeMap[nodeID] = struct{}{}
	}
	
	return nodeMap
}

// LegacyDependencyExtractor は依存関係抽出の戦略インターフェースです。
// Strategyパターンを使用して、異なる種類の依存関係抽出ロジックを実装します。
type LegacyDependencyExtractor interface {
	// Extract は解析結果から特定の種類の依存関係を抽出します
	Extract(result *LegacyResult) []LegacyDependencyInfo
}

// LegacyFieldDependencyExtractor は構造体フィールドの依存関係を抽出する実装です。
// 構造体のフィールドが他の型を参照している場合の依存関係を検出します。
type LegacyFieldDependencyExtractor struct {
	typeResolver *LegacyTypeResolver // 型情報の解決に使用するリゾルバ
}

// NewLegacyFieldDependencyExtractor は新しいLegacyFieldDependencyExtractorを作成します。
// 型解決に使用するTypeResolverを指定してインスタンスを初期化します。
func NewLegacyFieldDependencyExtractor(typeResolver *LegacyTypeResolver) *LegacyFieldDependencyExtractor {
	return &LegacyFieldDependencyExtractor{
		typeResolver: typeResolver,
	}
}

// Extract は構造体のフィールドによる依存関係を抽出します。
// 各構造体のフィールドを調べ、他の型を参照しているフィールドがある場合に
// 依存関係情報を作成します。
func (e *LegacyFieldDependencyExtractor) Extract(result *LegacyResult) []LegacyDependencyInfo {
	logger.Debug("フィールド依存関係抽出開始")
	var dependencies []LegacyDependencyInfo

	// ノードマップを作成（依存先の存在確認用）
	nodeMap := result.CreateNodeMap()

	for _, s := range result.Structs {
		fromID := types.NewNodeID(s.Package, s.Name)
		logger.Debug("構造体フィールド解析", "struct", s.Name, "package", s.Package)

		for _, field := range s.Fields {
			if toID := e.parseTypeToNodeID(field.Type, s.Package); toID != "" {
				if _, ok := nodeMap[toID]; ok {
					dependencies = append(dependencies, LegacyDependencyInfo{
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
func (e *LegacyFieldDependencyExtractor) parseTypeToNodeID(typeName, pkg string) types.NodeID {
	// より安全な型解析ロジック
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return types.NewNodeID(pkg, cleaned)
}


// LegacySignatureDependencyExtractor は関数シグネチャの依存関係を抽出する実装です。
// 関数やメソッドの引数や戻り値の型が他の型を参照している場合の
// 依存関係を検出します。
type LegacySignatureDependencyExtractor struct{}

// Extract は関数・メソッドのシグネチャによる依存関係を抽出します。
// 引数や戻り値の型を調べ、他の型を参照している場合に
// 依存関係情報を作成します。
func (e *LegacySignatureDependencyExtractor) Extract(result *LegacyResult) []LegacyDependencyInfo {
	var dependencies []LegacyDependencyInfo
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
func (e *LegacySignatureDependencyExtractor) extractFromParams(params []LegacyFieldInfo, fromID types.NodeID, pkg string, nodeMap map[types.NodeID]struct{}) []LegacyDependencyInfo {
	var dependencies []LegacyDependencyInfo
	for _, param := range params {
		if toID := e.parseTypeToNodeID(param.Type, pkg); toID != "" {
			if _, ok := nodeMap[toID]; ok {
				dependencies = append(dependencies, LegacyDependencyInfo{
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
// LegacyFieldDependencyExtractorの同名メソッドと同じロジックで型名を解析します。
func (e *LegacySignatureDependencyExtractor) parseTypeToNodeID(typeName, pkg string) types.NodeID {
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return types.NewNodeID(pkg, cleaned)
}


// LegacyBodyCallDependencyExtractor は関数本体の呼び出し依存関係を抽出する実装です。
// 関数やメソッドの本体内で他の関数や構造体を呼び出している場合の
// 依存関係を検出します。
type LegacyBodyCallDependencyExtractor struct{}

// Extract は関数・メソッド本体の呼び出しによる依存関係を抽出します。
// 関数やメソッドの本体で使用されている他の関数や型への呼び出しを
// 検出して依存関係情報を作成します。
func (e *LegacyBodyCallDependencyExtractor) Extract(result *LegacyResult) []LegacyDependencyInfo {
	var dependencies []LegacyDependencyInfo
	nodeMap := result.CreateNodeMap()

	// 関数本体の依存関係抽出
	for _, f := range result.Functions {
		fromID := types.NewNodeID(f.Package, f.Name)
		for _, called := range f.BodyCalls {
			toID := types.NewNodeID(f.Package, called)
			if _, ok := nodeMap[toID]; ok {
				dependencies = append(dependencies, LegacyDependencyInfo{
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
					dependencies = append(dependencies, LegacyDependencyInfo{
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


// LegacyPackageDependencyExtractor はパッケージ間の依存関係を抽出する実装です。
// import文に基づいてパッケージ間の依存関係を検出します。
// 同リポジトリ内のパッケージ間依存のみを対象とします。
type LegacyPackageDependencyExtractor struct {
	targetDir string // 解析対象のルートディレクトリパス
}

// NewLegacyPackageDependencyExtractor は新しいLegacyPackageDependencyExtractorを作成します。
// 解析対象のルートディレクトリを指定してインスタンスを初期化します。
func NewLegacyPackageDependencyExtractor(targetDir string) *LegacyPackageDependencyExtractor {
	return &LegacyPackageDependencyExtractor{
		targetDir: targetDir,
	}
}

// Extract はimport文に基づいたパッケージ間依存関係を抽出します。
// 各パッケージのimport文を調べ、同リポジトリ内の他パッケージを参照している場合に
// パッケージ間依存関係情報を作成します。
func (e *LegacyPackageDependencyExtractor) Extract(result *LegacyResult) []LegacyDependencyInfo {
	logger.Debug("パッケージ間依存関係抽出開始")
	var dependencies []LegacyDependencyInfo

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
					dependencies = append(dependencies, LegacyDependencyInfo{
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

// LegacyCrossPackageDependencyExtractor はパッケージ間の関数呼び出しや型の使用を抽出する実装です。
// import文で取り込まれたパッケージの関数や型を使用している箇所を検出し、
// パッケージ間の具体的な依存関係を分析します。
type LegacyCrossPackageDependencyExtractor struct {
	packageMap map[string]string // importエイリアスからパッケージ名へのマッピング（未使用）
}

// NewLegacyCrossPackageDependencyExtractor は新しいLegacyCrossPackageDependencyExtractorを作成します。
// パッケージ間の関数呼び出しや型使用の分析に必要な内部状態を初期化します。
func NewLegacyCrossPackageDependencyExtractor() *LegacyCrossPackageDependencyExtractor {
	return &LegacyCrossPackageDependencyExtractor{
		packageMap: make(map[string]string),
	}
}

// Extract はパッケージ間の関数呼び出しや型使用による依存関係を抽出します。
// 関数やメソッドの本体で使用されている他パッケージの関数や型を検出し、
// 具体的なパッケージ間依存関係情報を作成します。
func (e *LegacyCrossPackageDependencyExtractor) Extract(result *LegacyResult) []LegacyDependencyInfo {
	logger.Debug("パッケージ間関数呼び出し依存関係抽出開始")
	var dependencies []LegacyDependencyInfo
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
func (e *LegacyCrossPackageDependencyExtractor) extractCrossPackageCalls(bodyCalls []string, currentPkg string, packageImports map[string]map[string]string, fromID types.NodeID, nodeMap map[types.NodeID]struct{}) []LegacyDependencyInfo {
	var dependencies []LegacyDependencyInfo
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
							dependencies = append(dependencies, LegacyDependencyInfo{
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
func (e *LegacyCrossPackageDependencyExtractor) extractPackageAlias(importPath, alias string) string {
	if alias != "" {
		return alias // "."や"_"も含めて、指定されたエイリアスをそのまま返す
	}
	// エイリアスがない場合はパッケージ名を使用
	return utils.ExtractPackageName(importPath)
}
