package extraction

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/harakeishi/depsee/internal/types"
)

func TestSignatureDependencyExtractor_ExtractDependencies(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []DependencyInfo
	}{
		{
			name: "関数引数の依存関係",
			code: `package test
func CreateUser(profile Profile, settings UserSettings) {
}
type Profile struct {
	Name string
}
type UserSettings struct {
	Theme string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "CreateUser"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "CreateUser"),
					To:   types.NewNodeID("test", "UserSettings"),
					Type: types.SignatureDependency,
				},
			},
		},
		{
			name: "関数戻り値の依存関係",
			code: `package test
func GetUser() (User, error) {
	return User{}, nil
}
func GetProfile() *Profile {
	return nil
}
type User struct {
	Name string
}
type Profile struct {
	Name string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "GetUser"),
					To:   types.NewNodeID("test", "User"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "GetProfile"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.SignatureDependency,
				},
			},
		},
		{
			name: "メソッドのレシーバー依存関係",
			code: `package test
func (u *User) GetProfile() Profile {
	return u.Profile
}
func (u User) GetName() string {
	return u.Name
}
type User struct {
	Name    string
	Profile Profile
}
type Profile struct {
	Bio string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "GetProfile"),
					To:   types.NewNodeID("test", "User"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "GetProfile"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "GetName"),
					To:   types.NewNodeID("test", "User"),
					Type: types.SignatureDependency,
				},
			},
		},
		{
			name: "複雑なシグネチャ",
			code: `package test
func ProcessUsers(users []User, filter *Filter) ([]*Result, error) {
	return nil, nil
}
type User struct {
	Name string
}
type Filter struct {
	Active bool
}
type Result struct {
	Data string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "ProcessUsers"),
					To:   types.NewNodeID("test", "User"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "ProcessUsers"),
					To:   types.NewNodeID("test", "Filter"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "ProcessUsers"),
					To:   types.NewNodeID("test", "*Result"),
					Type: types.SignatureDependency,
				},
			},
		},
		{
			name: "基本型のみの関数",
			code: `package test
func Calculate(a int, b string) (float64, bool) {
	return 0.0, true
}`,
			expected: []DependencyInfo{},
		},
		{
			name: "無名引数・戻り値",
			code: `package test
func Process(User, *Profile) (Result, error) {
	return Result{}, nil
}
type User struct {
	Name string
}
type Profile struct {
	Bio string
}
type Result struct {
	Data string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "Process"),
					To:   types.NewNodeID("test", "User"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "Process"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "Process"),
					To:   types.NewNodeID("test", "Result"),
					Type: types.SignatureDependency,
				},
			},
		},
		{
			name: "インターフェース型のシグネチャ",
			code: `package test
func UseStorage(s Storage) error {
	return nil
}
func GetStorage() Storage {
	return nil
}
type Storage interface {
	Save() error
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "UseStorage"),
					To:   types.NewNodeID("test", "Storage"),
					Type: types.SignatureDependency,
				},
				{
					From: types.NewNodeID("test", "GetStorage"),
					To:   types.NewNodeID("test", "Storage"),
					Type: types.SignatureDependency,
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
			extractor := NewSignatureDependencyExtractor(ctx)
			
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

func TestSignatureDependencyExtractor_Name(t *testing.T) {
	ctx := NewContext(token.NewFileSet(), "test")
	extractor := NewSignatureDependencyExtractor(ctx)
	
	expected := "SignatureDependency"
	if name := extractor.Name(); name != expected {
		t.Errorf("Name() = %v, want %v", name, expected)
	}
}
