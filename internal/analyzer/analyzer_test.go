package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/harakeishi/depsee/internal/logger"
)

func TestMain(m *testing.M) {
	// テスト用のログ設定
	logger.Init(logger.Config{
		Level:  logger.LevelError, // テスト時はエラーのみ
		Format: "text",
		Output: os.Stderr,
	})

	code := m.Run()
	os.Exit(code)
}

func TestAnalyzeDir(t *testing.T) {
	// テストデータディレクトリのパス
	testDataDir := "../../testdata/sample"

	// 相対パスを絶対パスに変換
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("テストデータディレクトリのパス解決失敗: %v", err)
	}

	// ディレクトリの存在確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("テストデータディレクトリが存在しません: %s", absPath)
	}

	result, err := AnalyzeDir(absPath)
	if err != nil {
		t.Fatalf("AnalyzeDir failed: %v", err)
	}

	if result == nil {
		t.Fatal("結果がnilです")
	}

	// 構造体の検証
	expectedStructs := []string{"User", "Profile", "Post", "UserSettings"}
	if len(result.Structs) < len(expectedStructs) {
		t.Errorf("期待される構造体数: %d, 実際: %d", len(expectedStructs), len(result.Structs))
	}

	structNames := make(map[string]bool)
	for _, s := range result.Structs {
		structNames[s.Name] = true

		// 構造体の基本情報検証
		if s.Package == "" {
			t.Errorf("構造体 %s のパッケージ名が空です", s.Name)
		}
		if s.File == "" {
			t.Errorf("構造体 %s のファイル名が空です", s.Name)
		}
	}

	for _, expected := range expectedStructs {
		if !structNames[expected] {
			t.Errorf("期待される構造体が見つかりません: %s", expected)
		}
	}

	// インターフェースの検証
	expectedInterfaces := []string{"UserService"}
	if len(result.Interfaces) < len(expectedInterfaces) {
		t.Errorf("期待されるインターフェース数: %d, 実際: %d", len(expectedInterfaces), len(result.Interfaces))
	}

	// 関数の検証
	expectedFunctions := []string{"CreateUser", "GetUserPosts"}
	if len(result.Functions) < len(expectedFunctions) {
		t.Errorf("期待される関数数: %d, 実際: %d", len(expectedFunctions), len(result.Functions))
	}

	functionNames := make(map[string]bool)
	for _, f := range result.Functions {
		functionNames[f.Name] = true
	}

	for _, expected := range expectedFunctions {
		if !functionNames[expected] {
			t.Errorf("期待される関数が見つかりません: %s", expected)
		}
	}
}

func TestAnalyzeDirNonExistent(t *testing.T) {
	_, err := AnalyzeDir("/non/existent/directory")
	if err == nil {
		t.Error("存在しないディレクトリでエラーが発生しませんでした")
	}
}

func TestStructFieldAnalysis(t *testing.T) {
	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("テストデータディレクトリのパス解決失敗: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("テストデータディレクトリが存在しません: %s", absPath)
	}

	result, err := AnalyzeDir(absPath)
	if err != nil {
		t.Fatalf("AnalyzeDir failed: %v", err)
	}

	// User構造体のフィールド検証
	var userStruct *StructInfo
	for _, s := range result.Structs {
		if s.Name == "User" {
			userStruct = &s
			break
		}
	}

	if userStruct == nil {
		t.Fatal("User構造体が見つかりません")
	}

	expectedFields := map[string]string{
		"ID":       "int",
		"Name":     "string",
		"Email":    "string",
		"Profile":  "*Profile",
		"Posts":    "[]Post",
		"Settings": "UserSettings",
	}

	if len(userStruct.Fields) != len(expectedFields) {
		t.Errorf("User構造体のフィールド数が期待値と異なります。期待: %d, 実際: %d",
			len(expectedFields), len(userStruct.Fields))
	}

	for _, field := range userStruct.Fields {
		expectedType, exists := expectedFields[field.Name]
		if !exists {
			t.Errorf("予期しないフィールドが見つかりました: %s", field.Name)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("フィールド %s の型が期待値と異なります。期待: %s, 実際: %s",
				field.Name, expectedType, field.Type)
		}
	}
}

func TestMethodAnalysis(t *testing.T) {
	testDataDir := "../../testdata/sample"
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("テストデータディレクトリのパス解決失敗: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("テストデータディレクトリが存在しません: %s", absPath)
	}

	result, err := AnalyzeDir(absPath)
	if err != nil {
		t.Fatalf("AnalyzeDir failed: %v", err)
	}

	// User構造体のメソッド検証
	var userStruct *StructInfo
	for _, s := range result.Structs {
		if s.Name == "User" {
			userStruct = &s
			break
		}
	}

	if userStruct == nil {
		t.Fatal("User構造体が見つかりません")
	}

	expectedMethods := []string{"UpdateProfile", "AddPost"}
	if len(userStruct.Methods) < len(expectedMethods) {
		t.Errorf("User構造体のメソッド数が期待値より少ないです。期待: %d以上, 実際: %d",
			len(expectedMethods), len(userStruct.Methods))
	}

	methodNames := make(map[string]bool)
	for _, method := range userStruct.Methods {
		methodNames[method.Name] = true

		// メソッドのレシーバ検証
		if method.Receiver != "User" {
			t.Errorf("メソッド %s のレシーバが期待値と異なります。期待: User, 実際: %s",
				method.Name, method.Receiver)
		}
	}

	for _, expected := range expectedMethods {
		if !methodNames[expected] {
			t.Errorf("期待されるメソッドが見つかりません: %s", expected)
		}
	}
}

func TestExtractBodyCalls(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "同一パッケージ内の関数呼び出し",
			code: `
package test
func TestFunc() {
	New()
	CreateUser()
}`,
			expected: []string{"New", "CreateUser"},
		},
		{
			name: "パッケージ修飾子付きの関数呼び出し",
			code: `
package test
import "fmt"
func TestFunc() {
	fmt.Println("test")
	depsee.New()
}`,
			expected: []string{"fmt.Println", "depsee.New"},
		},
		{
			name: "構造体リテラル（パッケージ修飾子付き）",
			code: `
package test
func TestFunc() {
	config := depsee.Config{Name: "test"}
	_ = config
}`,
			expected: []string{"depsee.Config"},
		},
		{
			name: "構造体リテラル（同一パッケージ）",
			code: `
package test
func TestFunc() {
	user := User{Name: "test"}
	_ = user
}`,
			expected: []string{"User"},
		},
		{
			name: "メソッド呼び出し",
			code: `
package test
func TestFunc() {
	obj.Method()
	user.UpdateProfile()
}`,
			expected: []string{"obj.Method", "user.UpdateProfile"},
		},
		{
			name: "複合的なケース",
			code: `
package test
import "fmt"
func TestFunc() {
	fmt.Println("start")
	user := User{Name: "test"}
	config := depsee.Config{}
	depsee.New(config)
	user.UpdateProfile()
}`,
			expected: []string{"fmt.Println", "User", "depsee.Config", "depsee.New", "user.UpdateProfile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストコードをパース
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("コードのパースに失敗: %v", err)
			}

			// 関数を見つけてBodyCallsを抽出
			var funcBody *ast.BlockStmt
			for _, decl := range f.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "TestFunc" {
					funcBody = funcDecl.Body
					break
				}
			}

			if funcBody == nil {
				t.Fatal("TestFunc関数が見つかりません")
			}

			// extractBodyCallsを実行
			calls := extractBodyCalls(funcBody)

			// 結果を検証
			if len(calls) != len(tt.expected) {
				t.Errorf("呼び出し数が期待値と異なります。期待: %d, 実際: %d\n期待: %v\n実際: %v",
					len(tt.expected), len(calls), tt.expected, calls)
				return
			}

			// 順序は保証されないので、セットとして比較
			callSet := make(map[string]bool)
			for _, call := range calls {
				callSet[call] = true
			}

			for _, expected := range tt.expected {
				if !callSet[expected] {
					t.Errorf("期待される呼び出しが見つかりません: %s\n実際の呼び出し: %v", expected, calls)
				}
			}
		})
	}
}

func TestImportAliasExtraction(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedPath  string
		expectedAlias string
	}{
		{
			name: "エイリアス指定あり",
			code: `
package test
import f "fmt"
`,
			expectedPath:  "fmt",
			expectedAlias: "f",
		},
		{
			name: "エイリアス指定なし",
			code: `
package test
import "fmt"
`,
			expectedPath:  "fmt",
			expectedAlias: "fmt",
		},
		{
			name: "パッケージパスからの自動抽出",
			code: `
package test
import "github.com/user/repo/pkg/service"
`,
			expectedPath:  "github.com/user/repo/pkg/service",
			expectedAlias: "service",
		},
		{
			name: "ドット記法",
			code: `
package test
import . "fmt"
`,
			expectedPath:  "fmt",
			expectedAlias: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストコードをパース
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("コードのパースに失敗: %v", err)
			}

			// analyzeFileを実行
			result := &AnalysisResult{}
			analyzeFile(f, fset, "test.go", result)

			// パッケージ情報を検証
			if len(result.Packages) != 1 {
				t.Fatalf("パッケージ数が期待値と異なります。期待: 1, 実際: %d", len(result.Packages))
			}

			pkg := result.Packages[0]
			if len(pkg.Imports) != 1 {
				t.Fatalf("import数が期待値と異なります。期待: 1, 実際: %d", len(pkg.Imports))
			}

			imp := pkg.Imports[0]
			if imp.Path != tt.expectedPath {
				t.Errorf("importパスが期待値と異なります。期待: %s, 実際: %s", tt.expectedPath, imp.Path)
			}

			if imp.Alias != tt.expectedAlias {
				t.Errorf("importエイリアスが期待値と異なります。期待: %s, 実際: %s", tt.expectedAlias, imp.Alias)
			}
		})
	}
}
