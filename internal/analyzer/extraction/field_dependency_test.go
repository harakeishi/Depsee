package extraction

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/harakeishi/depsee/internal/types"
)

func TestFieldDependencyExtractor_ExtractDependencies(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []DependencyInfo
	}{
		{
			name: "基本的な構造体フィールド依存関係",
			code: `package test
type User struct {
	Profile *Profile
	Posts   []Post
}
type Profile struct {
	Name string
}
type Post struct {
	Title string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.FieldDependency,
				},
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "Post"),
					Type: types.FieldDependency,
				},
			},
		},
		{
			name: "基本型フィールドは除外",
			code: `package test
type User struct {
	Name string
	Age  int
	Active bool
}`,
			expected: []DependencyInfo{},
		},
		{
			name: "埋め込み構造体",
			code: `package test
type User struct {
	Profile
	Settings *UserSettings
}
type Profile struct {
	Name string
}
type UserSettings struct {
	Theme string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.FieldDependency,
				},
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "UserSettings"),
					Type: types.FieldDependency,
				},
			},
		},
		{
			name: "複雑なポインタとスライス",
			code: `package test
type User struct {
	Profile   *Profile
	Posts     []*Post
	Tags      []string
	Friends   []*User
}
type Profile struct {
	Name string
}
type Post struct {
	Title string
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "Profile"),
					Type: types.FieldDependency,
				},
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "*Post"),
					Type: types.FieldDependency,
				},
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "*User"),
					Type: types.FieldDependency,
				},
			},
		},
		{
			name: "インターフェース型",
			code: `package test
type User struct {
	Storage Storage
}
type Storage interface {
	Save() error
}`,
			expected: []DependencyInfo{
				{
					From: types.NewNodeID("test", "User"),
					To:   types.NewNodeID("test", "Storage"),
					Type: types.FieldDependency,
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
			extractor := NewFieldDependencyExtractor(ctx)
			
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

func TestFieldDependencyExtractor_Name(t *testing.T) {
	ctx := NewContext(token.NewFileSet(), "test")
	extractor := NewFieldDependencyExtractor(ctx)
	
	expected := "FieldDependency"
	if name := extractor.Name(); name != expected {
		t.Errorf("Name() = %v, want %v", name, expected)
	}
}

func TestFieldDependencyExtractor_ResolveType(t *testing.T) {
	tests := []struct {
		name        string
		typeStr     string
		currentPkg  string
		expected    string
	}{
		{
			name:       "基本型は除外",
			typeStr:    "string",
			currentPkg: "test",
			expected:   "",
		},
		{
			name:       "ポインタ型",
			typeStr:    "*User",
			currentPkg: "test",
			expected:   "User",
		},
		{
			name:       "スライス型",
			typeStr:    "[]User",
			currentPkg: "test",
			expected:   "User",
		},
		{
			name:       "スライスポインタ型",
			typeStr:    "[]*User",
			currentPkg: "test",
			expected:   "*User",
		},
		{
			name:       "同パッケージ型",
			typeStr:    "User",
			currentPkg: "test",
			expected:   "User",
		},
		{
			name:       "パッケージ修飾子付き型（対象外）",
			typeStr:    "pkg.User",
			currentPkg: "test",
			expected:   "",
		},
	}

	ctx := NewContext(token.NewFileSet(), "test")
	extractor := NewFieldDependencyExtractor(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.resolveType(tt.typeStr, tt.currentPkg)
			if result != tt.expected {
				t.Errorf("resolveType(%q, %q) = %q, want %q", tt.typeStr, tt.currentPkg, result, tt.expected)
			}
		})
	}
}
