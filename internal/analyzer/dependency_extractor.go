package analyzer

import (
	"strings"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/utils"
)

// DependencyInfo は依存関係情報を表す
type DependencyInfo struct {
	From NodeID
	To   NodeID
	Type DependencyType
}

// NodeID はノードの一意識別子
type NodeID string

// DependencyType は依存関係の種類
type DependencyType int

const (
	FieldDependency DependencyType = iota
	SignatureDependency
	BodyCallDependency
	CrossPackageDependency
	PackageDependency
	ConcreteInjectionDependency // 具象実装注入の依存関係タイプを追加
)

// DependencyExtractor は依存関係抽出の戦略インターフェース
type DependencyExtractor interface {
	Extract(result *Result) []DependencyInfo
}

// FieldDependencyExtractor は構造体フィールドの依存関係を抽出
type FieldDependencyExtractor struct {
	typeResolver *TypeResolver
}

// NewFieldDependencyExtractor は新しいFieldDependencyExtractorを作成
func NewFieldDependencyExtractor(typeResolver *TypeResolver) *FieldDependencyExtractor {
	return &FieldDependencyExtractor{
		typeResolver: typeResolver,
	}
}

func (e *FieldDependencyExtractor) Extract(result *Result) []DependencyInfo {
	logger.Debug("フィールド依存関係抽出開始")
	var dependencies []DependencyInfo

	// ノードマップを作成（依存先の存在確認用）
	nodeMap := e.createNodeMap(result)

	for _, s := range result.Structs {
		fromID := NodeID(s.Package + "." + s.Name)
		logger.Debug("構造体フィールド解析", "struct", s.Name, "package", s.Package)

		for _, field := range s.Fields {
			if toID := e.parseTypeToNodeID(field.Type, s.Package); toID != "" {
				if _, ok := nodeMap[toID]; ok {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   toID,
						Type: FieldDependency,
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

func (e *FieldDependencyExtractor) parseTypeToNodeID(typeName, pkg string) NodeID {
	// より安全な型解析ロジック
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return NodeID(pkg + "." + cleaned)
}

func (e *FieldDependencyExtractor) createNodeMap(result *Result) map[NodeID]struct{} {
	nodeMap := make(map[NodeID]struct{})
	
	// 構造体ノード登録
	for _, s := range result.Structs {
		id := NodeID(s.Package + "." + s.Name)
		nodeMap[id] = struct{}{}
	}

	// インターフェースノード登録
	for _, i := range result.Interfaces {
		id := NodeID(i.Package + "." + i.Name)
		nodeMap[id] = struct{}{}
	}

	// 関数ノード登録
	for _, f := range result.Functions {
		id := NodeID(f.Package + "." + f.Name)
		nodeMap[id] = struct{}{}
	}

	return nodeMap
}

// SignatureDependencyExtractor は関数シグネチャの依存関係を抽出
type SignatureDependencyExtractor struct{}

func (e *SignatureDependencyExtractor) Extract(result *Result) []DependencyInfo {
	var dependencies []DependencyInfo
	nodeMap := e.createNodeMap(result)

	// 関数の引数・戻り値の依存関係抽出
	for _, f := range result.Functions {
		fromID := NodeID(f.Package + "." + f.Name)
		dependencies = append(dependencies, e.extractFromParams(f.Params, fromID, f.Package, nodeMap)...)
		dependencies = append(dependencies, e.extractFromParams(f.Results, fromID, f.Package, nodeMap)...)
	}

	// メソッドの引数・戻り値の依存関係抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := NodeID(s.Package + "." + m.Name)
			dependencies = append(dependencies, e.extractFromParams(m.Params, fromID, s.Package, nodeMap)...)
			dependencies = append(dependencies, e.extractFromParams(m.Results, fromID, s.Package, nodeMap)...)
		}
	}

	return dependencies
}

func (e *SignatureDependencyExtractor) extractFromParams(params []FieldInfo, fromID NodeID, pkg string, nodeMap map[NodeID]struct{}) []DependencyInfo {
	var dependencies []DependencyInfo
	for _, param := range params {
		if toID := e.parseTypeToNodeID(param.Type, pkg); toID != "" {
			if _, ok := nodeMap[toID]; ok {
				dependencies = append(dependencies, DependencyInfo{
					From: fromID,
					To:   toID,
					Type: SignatureDependency,
				})
			}
		}
	}
	return dependencies
}

func (e *SignatureDependencyExtractor) parseTypeToNodeID(typeName, pkg string) NodeID {
	cleaned := strings.TrimLeft(typeName, "*[]")
	if cleaned == "" || strings.Contains(cleaned, "map[") {
		return ""
	}
	return NodeID(pkg + "." + cleaned)
}

func (e *SignatureDependencyExtractor) createNodeMap(result *Result) map[NodeID]struct{} {
	nodeMap := make(map[NodeID]struct{})
	
	for _, s := range result.Structs {
		id := NodeID(s.Package + "." + s.Name)
		nodeMap[id] = struct{}{}
	}

	for _, i := range result.Interfaces {
		id := NodeID(i.Package + "." + i.Name)
		nodeMap[id] = struct{}{}
	}

	for _, f := range result.Functions {
		id := NodeID(f.Package + "." + f.Name)
		nodeMap[id] = struct{}{}
	}

	return nodeMap
}

// BodyCallDependencyExtractor は関数本体の呼び出し依存関係を抽出
type BodyCallDependencyExtractor struct{}

func (e *BodyCallDependencyExtractor) Extract(result *Result) []DependencyInfo {
	var dependencies []DependencyInfo
	nodeMap := e.createNodeMap(result)

	// 関数本体の依存関係抽出
	for _, f := range result.Functions {
		fromID := NodeID(f.Package + "." + f.Name)
		for _, called := range f.BodyCalls {
			toID := NodeID(f.Package + "." + called)
			if _, ok := nodeMap[toID]; ok {
				dependencies = append(dependencies, DependencyInfo{
					From: fromID,
					To:   toID,
					Type: BodyCallDependency,
				})
			}
		}
	}

	// メソッド本体の依存関係抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := NodeID(s.Package + "." + m.Name)
			for _, called := range m.BodyCalls {
				toID := NodeID(s.Package + "." + called)
				if _, ok := nodeMap[toID]; ok {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   toID,
						Type: BodyCallDependency,
					})
				}
			}
		}
	}

	return dependencies
}

func (e *BodyCallDependencyExtractor) createNodeMap(result *Result) map[NodeID]struct{} {
	nodeMap := make(map[NodeID]struct{})
	
	for _, s := range result.Structs {
		id := NodeID(s.Package + "." + s.Name)
		nodeMap[id] = struct{}{}
	}

	for _, i := range result.Interfaces {
		id := NodeID(i.Package + "." + i.Name)
		nodeMap[id] = struct{}{}
	}

	for _, f := range result.Functions {
		id := NodeID(f.Package + "." + f.Name)
		nodeMap[id] = struct{}{}
	}

	return nodeMap
}

// PackageDependencyExtractor はパッケージ間の依存関係を抽出
type PackageDependencyExtractor struct {
	targetDir string // 解析対象のルートディレクトリ
}

// NewPackageDependencyExtractor は新しいPackageDependencyExtractorを作成
func NewPackageDependencyExtractor(targetDir string) *PackageDependencyExtractor {
	return &PackageDependencyExtractor{
		targetDir: targetDir,
	}
}

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
		fromID := NodeID("package:" + pkg.Name)

		for _, imp := range pkg.Imports {
			// 同リポジトリ内のパッケージかどうかを判定
			if utils.IsLocalPackage(imp.Path) {
				// パッケージ名を抽出（パスの最後の部分）
				targetPkgName := utils.ExtractPackageName(imp.Path)
				toID := NodeID("package:" + targetPkgName)

				// 依存先パッケージが存在する場合のみエッジを追加
				if _, ok := packageNodes[targetPkgName]; ok {
					dependencies = append(dependencies, DependencyInfo{
						From: fromID,
						To:   toID,
						Type: PackageDependency,
					})
					logger.Debug("パッケージ間依存関係追加", "from", fromID, "to", toID, "import", imp.Path)
				}
			}
		}
	}

	return dependencies
}

// CrossPackageDependencyExtractor はパッケージ間の関数呼び出しや型の使用を抽出
type CrossPackageDependencyExtractor struct {
	packageMap map[string]string // import alias -> package name のマッピング
}

// NewCrossPackageDependencyExtractor は新しいCrossPackageDependencyExtractorを作成
func NewCrossPackageDependencyExtractor() *CrossPackageDependencyExtractor {
	return &CrossPackageDependencyExtractor{
		packageMap: make(map[string]string),
	}
}

func (e *CrossPackageDependencyExtractor) Extract(result *Result) []DependencyInfo {
	logger.Debug("パッケージ間関数呼び出し依存関係抽出開始")
	var dependencies []DependencyInfo
	nodeMap := e.createNodeMap(result)

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
		fromID := NodeID(f.Package + "." + f.Name)
		dependencies = append(dependencies, e.extractCrossPackageCalls(f.BodyCalls, f.Package, packageImports, fromID, nodeMap)...)
	}

	// メソッドの本体から他パッケージの関数呼び出しを抽出
	for _, s := range result.Structs {
		for _, m := range s.Methods {
			fromID := NodeID(s.Package + "." + m.Name)
			dependencies = append(dependencies, e.extractCrossPackageCalls(m.BodyCalls, s.Package, packageImports, fromID, nodeMap)...)
		}
	}

	return dependencies
}

func (e *CrossPackageDependencyExtractor) extractCrossPackageCalls(bodyCalls []string, currentPkg string, packageImports map[string]map[string]string, fromID NodeID, nodeMap map[NodeID]struct{}) []DependencyInfo {
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
						toID := NodeID(targetPkgName + "." + funcOrTypeName)

						// 依存先ノードが存在する場合のみエッジを追加
						if _, ok := nodeMap[toID]; ok {
							dependencies = append(dependencies, DependencyInfo{
								From: fromID,
								To:   toID,
								Type: CrossPackageDependency,
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

func (e *CrossPackageDependencyExtractor) extractPackageAlias(importPath, alias string) string {
	if alias != "" {
		return alias // "."や"_"も含めて、指定されたエイリアスをそのまま返す
	}
	// エイリアスがない場合はパッケージ名を使用
	return utils.ExtractPackageName(importPath)
}

func (e *CrossPackageDependencyExtractor) createNodeMap(result *Result) map[NodeID]struct{} {
	nodeMap := make(map[NodeID]struct{})
	
	for _, s := range result.Structs {
		id := NodeID(s.Package + "." + s.Name)
		nodeMap[id] = struct{}{}
	}

	for _, i := range result.Interfaces {
		id := NodeID(i.Package + "." + i.Name)
		nodeMap[id] = struct{}{}
	}

	for _, f := range result.Functions {
		id := NodeID(f.Package + "." + f.Name)
		nodeMap[id] = struct{}{}
	}

	return nodeMap
}
