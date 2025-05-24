package analyzer

import (
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
