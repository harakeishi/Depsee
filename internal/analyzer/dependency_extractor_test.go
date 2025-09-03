package analyzer

import (
	"testing"
)

// テスト用のヘルパー関数
func createTestResult() *Result {
	return &Result{
		Structs: []StructInfo{
			{
				Name:    "User",
				Package: "test",
				File:    "test.go",
				Fields: []FieldInfo{
					{Name: "ID", Type: "int"},
					{Name: "Name", Type: "string"},
					{Name: "Profile", Type: "*Profile"},
					{Name: "Posts", Type: "[]Post"},
					{Name: "Settings", Type: "UserSettings"},
				},
			},
			{
				Name:    "Profile",
				Package: "test",
				File:    "test.go",
				Fields: []FieldInfo{
					{Name: "Bio", Type: "string"},
					{Name: "Avatar", Type: "string"},
					{Name: "User", Type: "*User"}, // 循環依存
				},
			},
			{
				Name:    "Post",
				Package: "test",
				File:    "test.go",
				Fields: []FieldInfo{
					{Name: "ID", Type: "int"},
					{Name: "Title", Type: "string"},
					{Name: "Author", Type: "*User"},
				},
			},
			{
				Name:    "UserSettings",
				Package: "test",
				File:    "test.go",
				Fields: []FieldInfo{
					{Name: "Theme", Type: "string"},
					{Name: "Notifications", Type: "bool"},
				},
			},
		},
		Interfaces: []InterfaceInfo{
			{
				Name:    "UserService",
				Package: "test",
				File:    "test.go",
			},
		},
		Functions: []FuncInfo{
			{
				Name:    "CreateUser",
				Package: "test",
				File:    "test.go",
				Params: []FieldInfo{
					{Name: "name", Type: "string"},
					{Name: "email", Type: "string"},
				},
				Results: []FieldInfo{
					{Name: "", Type: "*User"},
					{Name: "", Type: "error"},
				},
				BodyCalls: []string{"User", "Profile", "UserSettings"},
			},
			{
				Name:    "GetUserPosts",
				Package: "test",
				File:    "test.go",
				Params: []FieldInfo{
					{Name: "user", Type: "*User"},
				},
				Results: []FieldInfo{
					{Name: "", Type: "[]Post"},
				},
				BodyCalls: []string{},
			},
		},
		Packages: []PackageInfo{
			{
				Name: "test",
				Path: "github.com/example/test",
				File: "test.go",
				Imports: []ImportInfo{
					{Path: "fmt", Alias: "fmt"},
					{Path: "github.com/example/other", Alias: "other"},
				},
			},
		},
	}
}

func TestFieldDependencyExtractor(t *testing.T) {
	tests := []struct {
		name                 string
		result               *Result
		expectedDependencies []DependencyInfo
	}{
		{
			name:   "基本的なフィールド依存関係",
			result: createTestResult(),
			expectedDependencies: []DependencyInfo{
				{From: "test.User", To: "test.Profile", Type: FieldDependency},
				{From: "test.User", To: "test.Post", Type: FieldDependency},
				{From: "test.User", To: "test.UserSettings", Type: FieldDependency},
				{From: "test.Profile", To: "test.User", Type: FieldDependency},
				{From: "test.Post", To: "test.User", Type: FieldDependency},
			},
		},
		{
			name: "フィールドがない構造体",
			result: &Result{
				Structs: []StructInfo{
					{
						Name:    "EmptyStruct",
						Package: "test",
						File:    "test.go",
						Fields:  []FieldInfo{},
					},
				},
			},
			expectedDependencies: []DependencyInfo{},
		},
		{
			name: "プリミティブ型のみのフィールド",
			result: &Result{
				Structs: []StructInfo{
					{
						Name:    "SimpleStruct",
						Package: "test",
						File:    "test.go",
						Fields: []FieldInfo{
							{Name: "ID", Type: "int"},
							{Name: "Name", Type: "string"},
							{Name: "Active", Type: "bool"},
						},
					},
				},
			},
			expectedDependencies: []DependencyInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeResolver := NewTypeResolver()
			extractor := NewFieldDependencyExtractor(typeResolver)

			dependencies := extractor.Extract(tt.result)

			if len(dependencies) != len(tt.expectedDependencies) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDependencies), len(dependencies))
			}

			// 依存関係の存在確認
			for _, expected := range tt.expectedDependencies {
				found := false
				for _, actual := range dependencies {
					if actual.From == expected.From && actual.To == expected.To && actual.Type == expected.Type {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency not found: %+v", expected)
				}
			}
		})
	}
}

func TestFieldDependencyExtractor_parseTypeToNodeID(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		pkg      string
		expected NodeID
	}{
		{
			name:     "シンプルな型",
			typeName: "User",
			pkg:      "test",
			expected: "test.User",
		},
		{
			name:     "ポインタ型",
			typeName: "*User",
			pkg:      "test",
			expected: "test.User",
		},
		{
			name:     "スライス型",
			typeName: "[]User",
			pkg:      "test",
			expected: "test.User",
		},
		{
			name:     "ポインタスライス型",
			typeName: "*[]User",
			pkg:      "test",
			expected: "test.User",
		},
		{
			name:     "マップ型（除外される）",
			typeName: "map[string]User",
			pkg:      "test",
			expected: "",
		},
		{
			name:     "空の型名",
			typeName: "",
			pkg:      "test",
			expected: "",
		},
	}

	typeResolver := NewTypeResolver()
	extractor := NewFieldDependencyExtractor(typeResolver)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.parseTypeToNodeID(tt.typeName, tt.pkg)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSignatureDependencyExtractor(t *testing.T) {
	tests := []struct {
		name                 string
		result               *Result
		expectedDependencies []DependencyInfo
	}{
		{
			name:   "関数のシグネチャ依存関係",
			result: createTestResult(),
			expectedDependencies: []DependencyInfo{
				{From: "test.CreateUser", To: "test.User", Type: SignatureDependency},
				{From: "test.GetUserPosts", To: "test.User", Type: SignatureDependency},
				{From: "test.GetUserPosts", To: "test.Post", Type: SignatureDependency},
			},
		},
		{
			name: "メソッドのシグネチャ依存関係",
			result: &Result{
				Structs: []StructInfo{
					{
						Name:    "User",
						Package: "test",
						File:    "test.go",
						Methods: []FuncInfo{
							{
								Name:     "UpdateProfile",
								Package:  "test",
								File:     "test.go",
								Receiver: "User",
								Params: []FieldInfo{
									{Name: "profile", Type: "*Profile"},
								},
								Results: []FieldInfo{
									{Name: "", Type: "error"},
								},
							},
						},
					},
					{
						Name:    "Profile",
						Package: "test",
						File:    "test.go",
					},
				},
			},
			expectedDependencies: []DependencyInfo{
				{From: "test.UpdateProfile", To: "test.Profile", Type: SignatureDependency},
			},
		},
		{
			name: "プリミティブ型のみの関数",
			result: &Result{
				Functions: []FuncInfo{
					{
						Name:    "Add",
						Package: "test",
						File:    "test.go",
						Params: []FieldInfo{
							{Name: "a", Type: "int"},
							{Name: "b", Type: "int"},
						},
						Results: []FieldInfo{
							{Name: "", Type: "int"},
						},
					},
				},
			},
			expectedDependencies: []DependencyInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := &SignatureDependencyExtractor{}

			dependencies := extractor.Extract(tt.result)

			if len(dependencies) != len(tt.expectedDependencies) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDependencies), len(dependencies))
			}

			// 依存関係の存在確認
			for _, expected := range tt.expectedDependencies {
				found := false
				for _, actual := range dependencies {
					if actual.From == expected.From && actual.To == expected.To && actual.Type == expected.Type {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency not found: %+v", expected)
				}
			}
		})
	}
}

func TestBodyCallDependencyExtractor(t *testing.T) {
	tests := []struct {
		name                 string
		result               *Result
		expectedDependencies []DependencyInfo
	}{
		{
			name:   "関数本体の呼び出し依存関係",
			result: createTestResult(),
			expectedDependencies: []DependencyInfo{
				{From: "test.CreateUser", To: "test.User", Type: BodyCallDependency},
				{From: "test.CreateUser", To: "test.Profile", Type: BodyCallDependency},
				{From: "test.CreateUser", To: "test.UserSettings", Type: BodyCallDependency},
			},
		},
		{
			name: "メソッド本体の呼び出し依存関係",
			result: &Result{
				Structs: []StructInfo{
					{
						Name:    "User",
						Package: "test",
						File:    "test.go",
						Methods: []FuncInfo{
							{
								Name:      "Save",
								Package:   "test",
								File:      "test.go",
								Receiver:  "User",
								BodyCalls: []string{"Profile", "UserSettings"},
							},
						},
					},
					{
						Name:    "Profile",
						Package: "test",
						File:    "test.go",
					},
					{
						Name:    "UserSettings",
						Package: "test",
						File:    "test.go",
					},
				},
			},
			expectedDependencies: []DependencyInfo{
				{From: "test.Save", To: "test.Profile", Type: BodyCallDependency},
				{From: "test.Save", To: "test.UserSettings", Type: BodyCallDependency},
			},
		},
		{
			name: "呼び出しがない関数",
			result: &Result{
				Functions: []FuncInfo{
					{
						Name:      "SimpleFunc",
						Package:   "test",
						File:      "test.go",
						BodyCalls: []string{},
					},
				},
			},
			expectedDependencies: []DependencyInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := &BodyCallDependencyExtractor{}

			dependencies := extractor.Extract(tt.result)

			if len(dependencies) != len(tt.expectedDependencies) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDependencies), len(dependencies))
			}

			// 依存関係の存在確認
			for _, expected := range tt.expectedDependencies {
				found := false
				for _, actual := range dependencies {
					if actual.From == expected.From && actual.To == expected.To && actual.Type == expected.Type {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency not found: %+v", expected)
				}
			}
		})
	}
}

func TestPackageDependencyExtractor(t *testing.T) {
	tests := []struct {
		name                 string
		result               *Result
		targetDir            string
		expectedDependencies []DependencyInfo
	}{
		{
			name:                 "パッケージ間依存関係",
			result:               createTestResult(),
			targetDir:            "/test",
			expectedDependencies: []DependencyInfo{
				// 実際のローカルパッケージが存在しないため、依存関係は生成されない
			},
		},
		{
			name: "複数パッケージ間依存関係",
			result: &Result{
				Packages: []PackageInfo{
					{
						Name: "pkg1",
						Path: "github.com/example/pkg1",
						File: "pkg1.go",
						Imports: []ImportInfo{
							{Path: "github.com/example/pkg2", Alias: "pkg2"},
						},
					},
					{
						Name:    "pkg2",
						Path:    "github.com/example/pkg2",
						File:    "pkg2.go",
						Imports: []ImportInfo{},
					},
				},
			},
			targetDir: "/test",
			expectedDependencies: []DependencyInfo{
				{From: "package:pkg1", To: "package:pkg2", Type: PackageDependency},
			},
		},
		{
			name: "外部パッケージのみ（依存関係なし）",
			result: &Result{
				Packages: []PackageInfo{
					{
						Name: "test",
						Path: "github.com/example/test",
						File: "test.go",
						Imports: []ImportInfo{
							{Path: "fmt", Alias: "fmt"},
							{Path: "net/http", Alias: "http"},
						},
					},
				},
			},
			targetDir:            "/test",
			expectedDependencies: []DependencyInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewPackageDependencyExtractor(tt.targetDir)

			dependencies := extractor.Extract(tt.result)

			if len(dependencies) != len(tt.expectedDependencies) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDependencies), len(dependencies))
			}

			// 依存関係の存在確認
			for _, expected := range tt.expectedDependencies {
				found := false
				for _, actual := range dependencies {
					if actual.From == expected.From && actual.To == expected.To && actual.Type == expected.Type {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency not found: %+v", expected)
				}
			}
		})
	}
}

func TestCrossPackageDependencyExtractor(t *testing.T) {
	tests := []struct {
		name                 string
		result               *Result
		expectedDependencies []DependencyInfo
	}{
		{
			name: "パッケージ間関数呼び出し",
			result: &Result{
				Functions: []FuncInfo{
					{
						Name:      "UseOtherPackage",
						Package:   "test",
						File:      "test.go",
						BodyCalls: []string{"other.SomeFunc", "fmt.Println"},
					},
					{
						Name:    "SomeFunc",
						Package: "other",
						File:    "other.go",
					},
				},
				Packages: []PackageInfo{
					{
						Name: "test",
						Path: "github.com/example/test",
						File: "test.go",
						Imports: []ImportInfo{
							{Path: "github.com/example/other", Alias: "other"},
							{Path: "fmt", Alias: "fmt"},
						},
					},
				},
			},
			expectedDependencies: []DependencyInfo{
				{From: "test.UseOtherPackage", To: "other.SomeFunc", Type: CrossPackageDependency},
			},
		},
		{
			name: "メソッドの他パッケージ呼び出し",
			result: &Result{
				Structs: []StructInfo{
					{
						Name:    "User",
						Package: "test",
						File:    "test.go",
						Methods: []FuncInfo{
							{
								Name:      "Process",
								Package:   "test",
								File:      "test.go",
								Receiver:  "User",
								BodyCalls: []string{"utils.Validate", "fmt.Printf"},
							},
						},
					},
					{
						Name:    "Validate",
						Package: "utils",
						File:    "utils.go",
					},
				},
				Packages: []PackageInfo{
					{
						Name: "test",
						Path: "github.com/example/test",
						File: "test.go",
						Imports: []ImportInfo{
							{Path: "github.com/example/utils", Alias: "utils"},
							{Path: "fmt", Alias: "fmt"},
						},
					},
				},
			},
			expectedDependencies: []DependencyInfo{
				{From: "test.Process", To: "utils.Validate", Type: CrossPackageDependency},
			},
		},
		{
			name: "パッケージ修飾子なしの呼び出し",
			result: &Result{
				Functions: []FuncInfo{
					{
						Name:      "LocalFunc",
						Package:   "test",
						File:      "test.go",
						BodyCalls: []string{"LocalHelper", "len", "append"},
					},
				},
				Packages: []PackageInfo{
					{
						Name:    "test",
						Path:    "github.com/example/test",
						File:    "test.go",
						Imports: []ImportInfo{},
					},
				},
			},
			expectedDependencies: []DependencyInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewCrossPackageDependencyExtractor()

			dependencies := extractor.Extract(tt.result)

			if len(dependencies) != len(tt.expectedDependencies) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDependencies), len(dependencies))
			}

			// 依存関係の存在確認
			for _, expected := range tt.expectedDependencies {
				found := false
				for _, actual := range dependencies {
					if actual.From == expected.From && actual.To == expected.To && actual.Type == expected.Type {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency not found: %+v", expected)
				}
			}
		})
	}
}

func TestCrossPackageDependencyExtractor_extractPackageAlias(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		alias      string
		expected   string
	}{
		{
			name:       "エイリアスあり",
			importPath: "github.com/example/utils",
			alias:      "myutils",
			expected:   "myutils",
		},
		{
			name:       "エイリアスなし",
			importPath: "github.com/example/utils",
			alias:      "",
			expected:   "utils",
		},
		{
			name:       "ドットインポート",
			importPath: "github.com/example/utils",
			alias:      ".",
			expected:   ".",
		},
		{
			name:       "アンダースコアインポート",
			importPath: "github.com/example/utils",
			alias:      "_",
			expected:   "_",
		},
		{
			name:       "標準ライブラリ",
			importPath: "fmt",
			alias:      "",
			expected:   "fmt",
		},
	}

	extractor := NewCrossPackageDependencyExtractor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractPackageAlias(tt.importPath, tt.alias)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDependencyType_String(t *testing.T) {
	tests := []struct {
		name     string
		depType  DependencyType
		expected string
	}{
		{
			name:     "FieldDependency",
			depType:  FieldDependency,
			expected: "0", // iota value
		},
		{
			name:     "SignatureDependency",
			depType:  SignatureDependency,
			expected: "1", // iota value
		},
		{
			name:     "BodyCallDependency",
			depType:  BodyCallDependency,
			expected: "2", // iota value
		},
		{
			name:     "CrossPackageDependency",
			depType:  CrossPackageDependency,
			expected: "3", // iota value
		},
		{
			name:     "PackageDependency",
			depType:  PackageDependency,
			expected: "4", // iota value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(rune(tt.depType + '0')) // int to string conversion
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
