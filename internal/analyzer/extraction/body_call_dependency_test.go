package extraction

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/harakeishi/depsee/internal/types"
)

func TestBodyCallDependencyExtractor_ExtractDependencies(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []DependencyInfo
	}{
		{
			name: "基本的な関数呼び出し",
			code: `package test
func Main() {
	CreateUser()
	DeleteUser()
}
func CreateUser() {}
func DeleteUser() {}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "Main"),
					To:   types.NewNodeID("test", "CreateUser"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "Main"),
					To:   types.NewNodeID("test", "DeleteUser"),
					Type: types.BodyCallDependency,
				},
			},
		},
		{
			name: "組み込み関数の呼び出し",
			code: `package test
func ProcessData(data []string) []string {
	result := make([]string, len(data))
	for i := range data {
		result = append(result, data[i])
	}
	return result
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "ProcessData"),
					To:   types.NewNodeID("test", "make"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "ProcessData"),
					To:   types.NewNodeID("test", "len"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "ProcessData"),
					To:   types.NewNodeID("test", "append"),
					Type: types.BodyCallDependency,
				},
			},
		},
		{
			name: "メソッド呼び出し（パッケージ修飾子付きは除外）",
			code: `package test
func ProcessUser(u User) {
	ValidateUser()
	u.Save()
	fmt.Println("saved")
}
func ValidateUser() {}
type User struct{}
func (u User) Save() {}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "ProcessUser"),
					To:   types.NewNodeID("test", "ValidateUser"),
					Type: types.BodyCallDependency,
				},
			},
		},
		{
			name: "ネストした関数呼び出し",
			code: `package test
func ComplexOperation() {
	if checkCondition() {
		for i := 0; i < getLimit(); i++ {
			processItem()
		}
	}
}
func checkCondition() bool { return true }
func getLimit() int { return 10 }
func processItem() {}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "ComplexOperation"),
					To:   types.NewNodeID("test", "checkCondition"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "ComplexOperation"),
					To:   types.NewNodeID("test", "getLimit"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "ComplexOperation"),
					To:   types.NewNodeID("test", "processItem"),
					Type: types.BodyCallDependency,
				},
			},
		},
		{
			name: "関数本体がない関数（インターフェースなど）",
			code: `package test
func NoBody()
type Handler interface {
	Handle()
}`,
			expected: []DependencyInfo{},
		},
		{
			name: "空の関数本体",
			code: `package test
func EmptyFunc() {
}`,
			expected: []DependencyInfo{},
		},
		{
			name: "変数代入での関数呼び出し",
			code: `package test
func AssignmentCalls() {
	result := calculate()
	data, err := fetchData()
	_ = result
	_ = data
	_ = err
}
func calculate() int { return 0 }
func fetchData() (string, error) { return "", nil }`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "AssignmentCalls"),
					To:   types.NewNodeID("test", "calculate"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "AssignmentCalls"),
					To:   types.NewNodeID("test", "fetchData"),
					Type: types.BodyCallDependency,
				},
			},
		},
		{
			name: "関数リテラル内の呼び出し",
			code: `package test
func WithClosures() {
	fn := func() {
		helper()
	}
	fn()
}
func helper() {}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "WithClosures"),
					To:   types.NewNodeID("test", "helper"),
					Type: types.BodyCallDependency,
				},
				{
					From: types.NewNodeID("test", "WithClosures"),
					To:   types.NewNodeID("test", "fn"),
					Type: types.BodyCallDependency,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("コード解析エラー: %v", err)
			}

			ctx := NewContext(fset, "test")
			extractor := NewBodyCallDependencyExtractor(ctx)
			
			result, err := extractor.ExtractDependencies(file, fset, "test")
			if err != nil {
				t.Fatalf("依存関係抽出エラー: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("依存関係数が一致しません。期待値: %d, 実際: %d", len(tt.expected), len(result))
				for i, dep := range result {
					t.Logf("実際の依存関係[%d]: %s -> %s (%s)", i, dep.From, dep.To, dep.Type)
				}
				return
			}

			// 依存関係の順序は保証されないので、存在チェックで検証
			expectedMap := make(map[string]bool)
			for _, exp := range tt.expected {
				key := string(exp.From) + "->" + string(exp.To) + ":" + exp.Type.String()
				expectedMap[key] = true
			}

			for _, dep := range result {
				key := string(dep.From) + "->" + string(dep.To) + ":" + dep.Type.String()
				if !expectedMap[key] {
					t.Errorf("予期しない依存関係: %s -> %s (%s)", dep.From, dep.To, dep.Type)
				}
				delete(expectedMap, key)
			}

			if len(expectedMap) > 0 {
				for key := range expectedMap {
					t.Errorf("期待していた依存関係が見つかりません: %s", key)
				}
			}
		})
	}
}

func TestBodyCallDependencyExtractor_Name(t *testing.T) {
	ctx := NewContext(token.NewFileSet(), "test")
	extractor := NewBodyCallDependencyExtractor(ctx)
	
	expected := "BodyCallDependency"
	if name := extractor.Name(); name != expected {
		t.Errorf("Name() = %v, want %v", name, expected)
	}
}

func TestBodyCallDependencyExtractor_ExtractCalls(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "単純な関数呼び出し",
			code: `package test
func test() {
	foo()
	bar()
}`,
			expected: []string{"foo", "bar"},
		},
		{
			name: "パッケージ修飾子付き呼び出し",
			code: `package test
func test() {
	fmt.Println("test")
	os.Exit(0)
}`,
			expected: []string{"fmt.Println", "os.Exit"},
		},
		{
			name: "混在パターン",
			code: `package test
func test() {
	localFunc()
	pkg.RemoteFunc()
	len(data)
}`,
			expected: []string{"localFunc", "pkg.RemoteFunc", "len"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("コード解析エラー: %v", err)
			}

			ctx := NewContext(fset, "test")
			extractor := NewBodyCallDependencyExtractor(ctx)

			// 関数本体を取得
			var funcBody *ast.BlockStmt
			for _, decl := range file.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "test" {
					funcBody = funcDecl.Body
					break
				}
			}

			if funcBody == nil {
				t.Fatal("test関数が見つかりません")
			}

			result := extractor.extractCalls(funcBody)

			if len(result) != len(tt.expected) {
				t.Errorf("呼び出し数が一致しません。期待値: %d, 実際: %d", len(tt.expected), len(result))
				t.Logf("実際の呼び出し: %v", result)
				return
			}

			// 順序は保証されないので、存在チェック
			expectedMap := make(map[string]bool)
			for _, exp := range tt.expected {
				expectedMap[exp] = true
			}

			for _, call := range result {
				if !expectedMap[call] {
					t.Errorf("予期しない呼び出し: %s", call)
				}
				delete(expectedMap, call)
			}

			if len(expectedMap) > 0 {
				for call := range expectedMap {
					t.Errorf("期待していた呼び出しが見つかりません: %s", call)
				}
			}
		})
	}
}

func TestBodyCallDependencyExtractor_ResolveCall(t *testing.T) {
	tests := []struct {
		name       string
		call       string
		currentPkg string
		expected   string
	}{
		{
			name:       "ローカル関数呼び出し",
			call:       "localFunc",
			currentPkg: "test",
			expected:   "localFunc",
		},
		{
			name:       "パッケージ修飾子付き呼び出し",
			call:       "fmt.Println",
			currentPkg: "test",
			expected:   "",
		},
		{
			name:       "組み込み関数",
			call:       "len",
			currentPkg: "test",
			expected:   "len",
		},
	}

	ctx := NewContext(token.NewFileSet(), "test")
	extractor := NewBodyCallDependencyExtractor(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.resolveCall(tt.call, tt.currentPkg)
			if result != tt.expected {
				t.Errorf("resolveCall(%q, %q) = %q, want %q", tt.call, tt.currentPkg, result, tt.expected)
			}
		})
	}
}
