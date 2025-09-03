package extraction

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harakeishi/depsee/internal/types"
)

func TestStrategyBasedExtractor_AddStrategy(t *testing.T) {
	extractor := NewStrategyBasedExtractor(".")
	
	if len(extractor.strategies) != 0 {
		t.Errorf("初期状態のストラテジー数が0ではありません: %d", len(extractor.strategies))
	}

	ctx := NewContext(nil, "test")
	fieldExtractor := NewFieldDependencyExtractor(ctx)
	extractor.AddStrategy(fieldExtractor)
	
	if len(extractor.strategies) != 1 {
		t.Errorf("ストラテジー追加後の数が1ではありません: %d", len(extractor.strategies))
	}
	
	if extractor.strategies[0].Name() != "FieldDependency" {
		t.Errorf("追加されたストラテジーの名前が期待値と異なります: %s", extractor.strategies[0].Name())
	}
}

func TestDefaultStrategyBasedExtractor(t *testing.T) {
	extractor := DefaultStrategyBasedExtractor(".")
	
	expectedStrategies := []string{
		"FieldDependency",
		"SignatureDependency", 
		"BodyCallDependency",
		"PackageDependency",
		"CrossPackageDependency",
	}
	
	if len(extractor.strategies) != len(expectedStrategies) {
		t.Errorf("デフォルトストラテジー数が期待値と異なります。期待値: %d, 実際: %d", 
			len(expectedStrategies), len(extractor.strategies))
	}
	
	strategyNames := make(map[string]bool)
	for _, strategy := range extractor.strategies {
		strategyNames[strategy.Name()] = true
	}
	
	for _, expected := range expectedStrategies {
		if !strategyNames[expected] {
			t.Errorf("期待されたストラテジーが見つかりません: %s", expected)
		}
	}
}

func TestStrategyBasedExtractor_ExtractFromFiles(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "extraction_test")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用のGoファイルを作成
	testCode := `package test
type User struct {
	Profile Profile
}
type Profile struct {
	Name string
}
func CreateUser(p Profile) User {
	return User{Profile: p}
}`

	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("テストファイル作成エラー: %v", err)
	}

	extractor := DefaultStrategyBasedExtractor(tmpDir)
	result, err := extractor.ExtractFromFiles([]string{testFile})
	if err != nil {
		t.Fatalf("依存関係抽出エラー: %v", err)
	}

	// 最低限の依存関係が抽出されることを確認
	if len(result) == 0 {
		t.Error("依存関係が抽出されませんでした")
	}

	// FieldDependencyによる依存関係が含まれることを確認
	hasFieldDep := false
	for _, dep := range result {
		if dep.Type == types.FieldDependency {
			hasFieldDep = true
			break
		}
	}
	
	if !hasFieldDep {
		t.Error("FieldDependencyによる依存関係が抽出されませんでした")
	}
}

func TestStrategyBasedExtractor_ExtractFromFiles_NonExistentFile(t *testing.T) {
	extractor := DefaultStrategyBasedExtractor(".")
	
	result, err := extractor.ExtractFromFiles([]string{"nonexistent.go"})
	if err != nil {
		t.Fatalf("存在しないファイルでエラーが発生しました: %v", err)
	}
	
	// エラーがあっても他のファイルは処理を継続するため、空の結果が返される
	if len(result) != 0 {
		t.Errorf("存在しないファイルで依存関係が抽出されました: %d", len(result))
	}
}

func TestStrategyBasedExtractor_ExtractFromFiles_EmptyList(t *testing.T) {
	extractor := DefaultStrategyBasedExtractor(".")
	
	result, err := extractor.ExtractFromFiles([]string{})
	if err != nil {
		t.Fatalf("空のファイルリストでエラーが発生しました: %v", err)
	}
	
	if len(result) != 0 {
		t.Errorf("空のファイルリストで依存関係が抽出されました: %d", len(result))
	}
}

func TestStrategyBasedExtractor_ExtractImportMap(t *testing.T) {
	extractor := NewStrategyBasedExtractor(".")
	
	code := `package test
import (
	"fmt"
	"os"
	alias "path/filepath"
	. "strings"
)`

	// パースしてファイルを作成
	tmpDir, err := os.MkdirTemp("", "extraction_test")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("テストファイル作成エラー: %v", err)
	}

	// extractFromFileを呼んでimportMapを間接的にテスト
	_, err = extractor.extractFromFile(testFile)
	if err != nil {
		t.Fatalf("ファイル抽出エラー: %v", err)
	}
}

func TestConvertToDependencyInfo(t *testing.T) {
	deps := []DependencyInfo{
		{
			From: types.NewNodeID("test", "User"),
			To:   types.NewNodeID("test", "Profile"),
			Type: types.FieldDependency,
		},
		{
			From: types.NewNodeID("test", "CreateUser"),
			To:   types.NewNodeID("test", "User"),
			Type: types.SignatureDependency,
		},
	}
	
	result := ConvertToDependencyInfo(deps)
	
	if len(result) != len(deps) {
		t.Errorf("変換後の依存関係数が一致しません。期待値: %d, 実際: %d", len(deps), len(result))
	}
	
	// 型変換の確認のため、interface{}として返されることを確認
	if result == nil {
		t.Error("変換結果がnilです")
	}
}

func TestStrategyBasedExtractor_ErrorHandling(t *testing.T) {
	// エラーを発生させるストラテジーを作成
	extractor := NewStrategyBasedExtractor(".")
	errorStrat := &errorStrategy{}
	extractor.AddStrategy(errorStrat)

	// テスト用ファイルを作成
	tmpDir, err := os.MkdirTemp("", "extraction_test")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCode := `package test
type User struct {
	Name string
}`

	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("テストファイル作成エラー: %v", err)
	}

	result, err := extractor.ExtractFromFiles([]string{testFile})
	if err != nil {
		t.Fatalf("エラーストラテジーでファイル抽出が失敗しました: %v", err)
	}
	
	// エラーストラテジーのエラーは継続処理で無視されるため、結果は空
	if len(result) != 0 {
		t.Errorf("エラーストラテジーで依存関係が抽出されました: %d", len(result))
	}
}

// テスト用のファイルシステムヘルパー
func createTestFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("テストファイル作成エラー (%s): %v", filename, err)
	}
	return filePath
}

func TestStrategyBasedExtractor_MultipleFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extraction_multi_test")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 複数のファイルを作成
	file1 := createTestFile(t, tmpDir, "user.go", `package test
type User struct {
	Profile Profile
}`)

	file2 := createTestFile(t, tmpDir, "profile.go", `package test
type Profile struct {
	Name string
}
func GetProfile() Profile {
	return Profile{}
}`)

	extractor := DefaultStrategyBasedExtractor(tmpDir)
	result, err := extractor.ExtractFromFiles([]string{file1, file2})
	if err != nil {
		t.Fatalf("複数ファイル抽出エラー: %v", err)
	}

	if len(result) == 0 {
		t.Error("複数ファイルから依存関係が抽出されませんでした")
	}

	// 両方のファイルからの依存関係が含まれることを確認
	// (詳細な検証は個別のストラテジーテストで行うため、ここでは基本的な動作確認のみ)
	t.Logf("抽出された依存関係数: %d", len(result))
	for i, dep := range result {
		t.Logf("依存関係[%d]: %s -> %s (%s)", i, dep.From, dep.To, dep.Type)
	}
}
