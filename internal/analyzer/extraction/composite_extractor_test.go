package extraction

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/harakeishi/depsee/internal/types"
)

func TestCompositeExtractor_ExtractDependencies(t *testing.T) {
	code := `package test
type User struct {
	Profile Profile
}
type Profile struct {
	Name string
}
func CreateUser(p Profile) User {
	ValidateProfile()
	return User{Profile: p}
}
func ValidateProfile() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("コード解析エラー: %v", err)
	}

	ctx := NewContext(fset, "test")

	// 個別のストラテジーを作成
	fieldExtractor := NewFieldDependencyExtractor(ctx)
	signatureExtractor := NewSignatureDependencyExtractor(ctx)
	bodyCallExtractor := NewBodyCallDependencyExtractor(ctx)

	// CompositeExtractorを作成
	composite := NewCompositeExtractor(fieldExtractor, signatureExtractor, bodyCallExtractor)

	result, err := composite.ExtractDependencies(file, fset, "test")
	if err != nil {
		t.Fatalf("依存関係抽出エラー: %v", err)
	}

	// 期待される依存関係
	expected := []DependencyInfo{
		// FieldDependency
		{
			From: types.NewNodeID("test", "User"),
			To:   types.NewNodeID("test", "Profile"),
			Type: types.FieldDependency,
		},
		// SignatureDependency
		{
			From: types.NewNodeID("test", "CreateUser"),
			To:   types.NewNodeID("test", "Profile"),
			Type: types.SignatureDependency,
		},
		{
			From: types.NewNodeID("test", "CreateUser"),
			To:   types.NewNodeID("test", "User"),
			Type: types.SignatureDependency,
		},
		// BodyCallDependency
		{
			From: types.NewNodeID("test", "CreateUser"),
			To:   types.NewNodeID("test", "ValidateProfile"),
			Type: types.BodyCallDependency,
		},
	}

	if len(result) != len(expected) {
		t.Errorf("依存関係数が一致しません。期待値: %d, 実際: %d", len(expected), len(result))
		for i, dep := range result {
			t.Logf("実際の依存関係[%d]: %s -> %s (%s)", i, dep.From, dep.To, dep.Type)
		}
		return
	}

	// 依存関係の存在チェック
	expectedMap := make(map[string]bool)
	for _, exp := range expected {
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
}

func TestCompositeExtractor_Name(t *testing.T) {
	composite := NewCompositeExtractor()
	
	expected := "CompositeExtractor"
	if name := composite.Name(); name != expected {
		t.Errorf("Name() = %v, want %v", name, expected)
	}
}

func TestCompositeExtractor_AddStrategy(t *testing.T) {
	composite := NewCompositeExtractor()
	
	// 初期状態では空
	if len(composite.strategies) != 0 {
		t.Errorf("初期状態のストラテジー数が0ではありません: %d", len(composite.strategies))
	}

	ctx := NewContext(token.NewFileSet(), "test")
	fieldExtractor := NewFieldDependencyExtractor(ctx)
	
	composite.AddStrategy(fieldExtractor)
	
	if len(composite.strategies) != 1 {
		t.Errorf("ストラテジー追加後の数が1ではありません: %d", len(composite.strategies))
	}
	
	if composite.strategies[0].Name() != "FieldDependency" {
		t.Errorf("追加されたストラテジーの名前が期待値と異なります: %s", composite.strategies[0].Name())
	}
}

func TestCompositeExtractor_RemoveStrategy(t *testing.T) {
	ctx := NewContext(token.NewFileSet(), "test")
	fieldExtractor := NewFieldDependencyExtractor(ctx)
	signatureExtractor := NewSignatureDependencyExtractor(ctx)
	
	composite := NewCompositeExtractor(fieldExtractor, signatureExtractor)
	
	// 初期状態で2つのストラテジー
	if len(composite.strategies) != 2 {
		t.Errorf("初期ストラテジー数が2ではありません: %d", len(composite.strategies))
	}
	
	// FieldDependencyを削除
	composite.RemoveStrategy("FieldDependency")
	
	if len(composite.strategies) != 1 {
		t.Errorf("削除後のストラテジー数が1ではありません: %d", len(composite.strategies))
	}
	
	if composite.strategies[0].Name() != "SignatureDependency" {
		t.Errorf("残ったストラテジーの名前が期待値と異なります: %s", composite.strategies[0].Name())
	}
	
	// 存在しないストラテジーを削除
	composite.RemoveStrategy("NonExistent")
	
	if len(composite.strategies) != 1 {
		t.Errorf("存在しないストラテジー削除後の数が変わってしまいました: %d", len(composite.strategies))
	}
}

func TestCompositeExtractor_EmptyStrategies(t *testing.T) {
	composite := NewCompositeExtractor()
	
	code := `package test
type User struct {
	Name string
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("コード解析エラー: %v", err)
	}

	result, err := composite.ExtractDependencies(file, fset, "test")
	if err != nil {
		t.Fatalf("依存関係抽出エラー: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("空のストラテジーで依存関係が抽出されました: %d", len(result))
	}
}

// Mock strategy for testing error handling
type errorStrategy struct{}

func (e *errorStrategy) ExtractDependencies(file *ast.File, fset *token.FileSet, packageName string) ([]DependencyInfo, error) {
	return nil, &mockError{"test error"}
}

func (e *errorStrategy) Name() string {
	return "ErrorStrategy"
}

type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestCompositeExtractor_ErrorHandling(t *testing.T) {
	ctx := NewContext(token.NewFileSet(), "test")
	fieldExtractor := NewFieldDependencyExtractor(ctx)
	errorStrat := &errorStrategy{}
	
	composite := NewCompositeExtractor(fieldExtractor, errorStrat)

	code := `package test
type User struct {
	Name string
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("コード解析エラー: %v", err)
	}

	_, err = composite.ExtractDependencies(file, fset, "test")
	if err == nil {
		t.Error("エラーが発生すべきですが、nilが返されました")
	}
	
	if err.Error() != "test error" {
		t.Errorf("期待したエラーメッセージが返されませんでした: %s", err.Error())
	}
}
