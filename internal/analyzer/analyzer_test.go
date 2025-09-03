package analyzer

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
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

func TestGoAnalyzer_SetFilters(t *testing.T) {
	type fields struct {
		Filters   Filters
		filesPath []string
		Result    *Result
	}
	type args struct {
		filters Filters
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetFilters_SingleTargetPackage",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{TargetPackages: []string{"test"}}},
		},
		{
			name: "SetFilters_MultipleTargetPackages",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{TargetPackages: []string{"test1", "test2", "test3"}}},
		},
		{
			name: "SetFilters_ExcludePackages",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{
				TargetPackages:  []string{"test"},
				ExcludePackages: []string{"exclude1", "exclude2"},
			}},
		},
		{
			name: "SetFilters_ExcludeDirs",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{
				TargetPackages: []string{"test"},
				ExcludeDirs:    []string{"/tmp", "/test"},
			}},
		},
		{
			name: "SetFilters_AllFilters",
			fields: fields{
				Filters: Filters{},
			},
			args: args{filters: Filters{
				TargetPackages:  []string{"main", "utils", "config"},
				ExcludePackages: []string{"test", "mock"},
				ExcludeDirs:     []string{"/vendor", "/node_modules", "/tmp"},
			}},
		},
		{
			name: "SetFilters_EmptyFilters",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"existing"}},
			},
			args: args{filters: Filters{}},
		},
		{
			name: "SetFilters_OverwriteExistingFilters",
			fields: fields{
				Filters: Filters{
					TargetPackages:  []string{"old1", "old2"},
					ExcludePackages: []string{"oldExclude"},
					ExcludeDirs:     []string{"/oldDir"},
				},
			},
			args: args{filters: Filters{
				TargetPackages:  []string{"new1", "new2"},
				ExcludePackages: []string{"newExclude"},
				ExcludeDirs:     []string{"/newDir"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ga := &GoAnalyzer{
				Filters:   tt.fields.Filters,
				filesPath: tt.fields.filesPath,
				Result:    tt.fields.Result,
			}
			ga.SetFilters(tt.args.filters)

			// TargetPackagesの確認
			if !slices.Equal(ga.Filters.TargetPackages, tt.args.filters.TargetPackages) {
				t.Errorf("Filters.TargetPackages = %v, want %v", ga.Filters.TargetPackages, tt.args.filters.TargetPackages)
			}

			// ExcludePackagesの確認
			if !slices.Equal(ga.Filters.ExcludePackages, tt.args.filters.ExcludePackages) {
				t.Errorf("Filters.ExcludePackages = %v, want %v", ga.Filters.ExcludePackages, tt.args.filters.ExcludePackages)
			}

			// ExcludeDirsの確認
			if !slices.Equal(ga.Filters.ExcludeDirs, tt.args.filters.ExcludeDirs) {
				t.Errorf("Filters.ExcludeDirs = %v, want %v", ga.Filters.ExcludeDirs, tt.args.filters.ExcludeDirs)
			}

		})
	}
}

func TestGoAnalyzer_ListTartgetFiles(t *testing.T) {
	type fields struct {
		Filters   Filters
		filesPath []string
		Result    *Result
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantFilesPath []string
	}{
		{
			name: "ListTargetFiles_ValidDirectory",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"sample"}},
			},
			args:          args{dir: "../../testdata/sample"},
			wantErr:       false,
			wantFilesPath: []string{"../../testdata/sample/user.go"},
		},
		{
			name: "ListTargetFiles_MultiplePackages",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"pkg1", "pkg2"}},
			},
			args:    args{dir: "../../testdata/multi-package"},
			wantErr: false,
			wantFilesPath: []string{
				"../../testdata/multi-package/pkg1/models.go",
				"../../testdata/multi-package/pkg2/profile.go",
			},
		},
		{
			name: "ListTargetFiles_WithExcludePackages",
			fields: fields{
				Filters: Filters{
					TargetPackages:  []string{"pkg1", "pkg2"},
					ExcludePackages: []string{"pkg2"},
				},
			},
			args:          args{dir: "../../testdata/multi-package"},
			wantErr:       false,
			wantFilesPath: []string{"../../testdata/multi-package/pkg1/models.go"},
		},
		{
			name: "ListTargetFiles_AllPackages",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"sample", "pkg1", "pkg2", "reserved"}},
			},
			args:    args{dir: "../../testdata"},
			wantErr: false,
			wantFilesPath: []string{
				"../../testdata/multi-package/pkg1/models.go",
				"../../testdata/multi-package/pkg2/profile.go",
				"../../testdata/reserved_words/test.go",
				"../../testdata/sample/user.go",
			},
		},
		{
			name: "ListTargetFiles_NonExistentDirectory",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"sample"}},
			},
			args:          args{dir: "../../testdata/nonexistent"},
			wantErr:       true,
			wantFilesPath: []string{},
		},
		{
			name: "ListTargetFiles_NoMatchingPackages",
			fields: fields{
				Filters: Filters{TargetPackages: []string{"nonexistent"}},
			},
			args:          args{dir: "../../testdata/sample"},
			wantErr:       false,
			wantFilesPath: []string{},
		},
		{
			name: "ListTargetFiles_EmptyTargetPackages",
			fields: fields{
				Filters: Filters{TargetPackages: []string{}},
			},
			args:          args{dir: "../../testdata/sample"},
			wantErr:       false,
			wantFilesPath: []string{"../../testdata/sample/user.go"}, // 空の場合はすべてのファイルが対象
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ga := &GoAnalyzer{
				Filters:   tt.fields.Filters,
				filesPath: tt.fields.filesPath,
				Result:    tt.fields.Result,
			}
			err := ga.ListTartgetFiles(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoAnalyzer.ListTartgetFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// パスを正規化して比較（順序は問わない）
				normalizedFilesPath := normalizeAndSortPaths(ga.filesPath)
				expectedPaths := normalizeAndSortPaths(tt.wantFilesPath)

				if !slices.Equal(normalizedFilesPath, expectedPaths) {
					t.Errorf("GoAnalyzer.ListTartgetFiles() filesPath = %v, want %v", normalizedFilesPath, expectedPaths)
				}
			}
		})
	}
}

// normalizeAndSortPaths はパスを正規化してソートする
func normalizeAndSortPaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for i, path := range paths {
		// パスを絶対パスに変換
		absPath, err := filepath.Abs(path)
		if err != nil {
			normalized[i] = path // エラーの場合は元のパスを使用
		} else {
			normalized[i] = absPath
		}
	}
	slices.Sort(normalized)
	return normalized
}

func TestFilters_shouldIncludeFile(t *testing.T) {
	type fields struct {
		TargetPackages  []string
		ExcludePackages []string
		ExcludeDirs     []string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "shouldIncludeFile_TargetPackageMatch",
			fields: fields{
				TargetPackages: []string{"sample"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_TargetPackageNoMatch",
			fields: fields{
				TargetPackages: []string{"nonexistent"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    false,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_MultipleTargetPackages",
			fields: fields{
				TargetPackages: []string{"pkg1", "pkg2", "sample"},
			},
			args:    args{path: "../../testdata/multi-package/pkg1/models.go"},
			want:    true,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_ExcludePackage",
			fields: fields{
				TargetPackages:  []string{"sample"},
				ExcludePackages: []string{"sample"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    false,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_ExcludePackageNoMatch",
			fields: fields{
				TargetPackages:  []string{"sample"},
				ExcludePackages: []string{"other"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_ExcludeDirectory",
			fields: fields{
				TargetPackages: []string{"sample"},
				ExcludeDirs:    []string{"../../testdata/sample"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    false,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_ExcludeDirectoryNoMatch",
			fields: fields{
				TargetPackages: []string{"sample"},
				ExcludeDirs:    []string{"../../testdata/other"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_MultipleExcludeDirs",
			fields: fields{
				TargetPackages: []string{"pkg1", "pkg2"},
				ExcludeDirs:    []string{"../../testdata/multi-package/pkg1", "../../testdata/sample"},
			},
			args:    args{path: "../../testdata/multi-package/pkg1/models.go"},
			want:    false,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_ComplexFilter",
			fields: fields{
				TargetPackages:  []string{"pkg1", "pkg2", "sample"},
				ExcludePackages: []string{"pkg2"},
				ExcludeDirs:     []string{"../../testdata/sample"},
			},
			args:    args{path: "../../testdata/multi-package/pkg1/models.go"},
			want:    true,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_ComplexFilterExcluded",
			fields: fields{
				TargetPackages:  []string{"pkg1", "pkg2", "sample"},
				ExcludePackages: []string{"pkg1"},
				ExcludeDirs:     []string{"../../testdata/sample"},
			},
			args:    args{path: "../../testdata/multi-package/pkg1/models.go"},
			want:    false,
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_NonExistentFile",
			fields: fields{
				TargetPackages: []string{"sample"},
			},
			args:    args{path: "../../testdata/nonexistent.go"},
			want:    false,
			wantErr: true,
		},
		{
			name: "shouldIncludeFile_EmptyTargetPackages",
			fields: fields{
				TargetPackages: []string{},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // 空の場合はすべてのパッケージを対象とする
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_EmptyExcludePackages",
			fields: fields{
				TargetPackages:  []string{"sample"},
				ExcludePackages: []string{},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // 空の除外リストは除外しない
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_EmptyExcludeDirs",
			fields: fields{
				TargetPackages: []string{"sample"},
				ExcludeDirs:    []string{},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // 空の除外ディレクトリは除外しない
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_AllFiltersEmpty",
			fields: fields{
				TargetPackages:  []string{},
				ExcludePackages: []string{},
				ExcludeDirs:     []string{},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // すべて空の場合はすべてを対象とする
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_NilTargetPackages",
			fields: fields{
				TargetPackages: nil,
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // nilの場合はすべてのパッケージを対象とする
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_NilExcludePackages",
			fields: fields{
				TargetPackages:  []string{"sample"},
				ExcludePackages: nil,
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // nilの除外リストは除外しない
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_NilExcludeDirs",
			fields: fields{
				TargetPackages: []string{"sample"},
				ExcludeDirs:    nil,
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    true, // nilの除外ディレクトリは除外しない
			wantErr: false,
		},
		{
			name: "shouldIncludeFile_EmptyTargetWithExclude",
			fields: fields{
				TargetPackages:  []string{},
				ExcludePackages: []string{"sample"},
			},
			args:    args{path: "../../testdata/sample/user.go"},
			want:    false, // 空のターゲットでも除外は有効
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filters{
				TargetPackages:  tt.fields.TargetPackages,
				ExcludePackages: tt.fields.ExcludePackages,
				ExcludeDirs:     tt.fields.ExcludeDirs,
			}
			got, err := f.shouldIncludeFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filters.shouldIncludeFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Filters.shouldIncludeFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []ImportInfo
	}{
		{
			name: "extractImports_StandardLibrary",
			content: `package test

import "fmt"

func main() {}`,
			expected: []ImportInfo{
				{Path: "fmt", Alias: "fmt"},
			},
		},
		{
			name: "extractImports_MultipleImports",
			content: `package test

import (
	"fmt"
	"strings"
	"os"
)

func main() {}`,
			expected: []ImportInfo{
				{Path: "fmt", Alias: "fmt"},
				{Path: "strings", Alias: "strings"},
				{Path: "os", Alias: "os"},
			},
		},
		{
			name: "extractImports_WithAlias",
			content: `package test

import (
	"fmt"
	str "strings"
	. "os"
	_ "log"
)

func main() {}`,
			expected: []ImportInfo{
				{Path: "fmt", Alias: "fmt"},
				{Path: "strings", Alias: "str"},
				{Path: "os", Alias: "."},
				{Path: "log", Alias: "_"},
			},
		},
		{
			name: "extractImports_ExternalPackages",
			content: `package test

import (
	"fmt"
	"github.com/harakeishi/depsee/internal/logger"
	"github.com/some/other/package"
)

func main() {}`,
			expected: []ImportInfo{
				{Path: "fmt", Alias: "fmt"},
				{Path: "github.com/harakeishi/depsee/internal/logger", Alias: "logger"},
				{Path: "github.com/some/other/package", Alias: "package"},
			},
		},
		{
			name: "extractImports_NoImports",
			content: `package test

func main() {}`,
			expected: []ImportInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストコードをパースしてASTを作成
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.content, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse test content: %v", err)
			}

			// extractImports関数をテスト
			result := extractImports(f)

			// 結果の検証
			if len(result) != len(tt.expected) {
				t.Errorf("extractImports() returned %d imports, expected %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("extractImports() missing import at index %d", i)
					continue
				}
				if result[i].Path != expected.Path {
					t.Errorf("extractImports()[%d].Path = %q, expected %q", i, result[i].Path, expected.Path)
				}
				if result[i].Alias != expected.Alias {
					t.Errorf("extractImports()[%d].Alias = %q, expected %q", i, result[i].Alias, expected.Alias)
				}
			}
		})
	}
}

func TestExtractTypes(t *testing.T) {
	tests := []struct {
		name               string
		content            string
		expectedStructs    []StructInfo
		expectedInterfaces []InterfaceInfo
		expectedStructMap  map[string]string // 簡略化：構造体名のみを検証
	}{
		{
			name: "extractTypes_SimpleStruct",
			content: `package test

type User struct {
	ID   int
	Name string
}`,
			expectedStructs: []StructInfo{
				{
					Name:    "User",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "ID", Type: "int"},
						{Name: "Name", Type: "string"},
					},
				},
			},
			expectedInterfaces: []InterfaceInfo{},
			expectedStructMap:  map[string]string{"User": "User"},
		},
		{
			name: "extractTypes_SimpleInterface",
			content: `package test

type Reader interface {
	Read([]byte) (int, error)
}`,
			expectedStructs: []StructInfo{},
			expectedInterfaces: []InterfaceInfo{
				{
					Name:    "Reader",
					Package: "test",
				},
			},
			expectedStructMap: map[string]string{},
		},
		{
			name: "extractTypes_StructWithEmbedding",
			content: `package test

type Base struct {
	ID int
}

type User struct {
	Base
	Name string
}`,
			expectedStructs: []StructInfo{
				{
					Name:    "Base",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "ID", Type: "int"},
					},
				},
				{
					Name:    "User",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "", Type: "Base"},
						{Name: "Name", Type: "string"},
					},
				},
			},
			expectedInterfaces: []InterfaceInfo{},
			expectedStructMap:  map[string]string{"Base": "Base", "User": "User"},
		},
		{
			name: "extractTypes_MultipleTypes",
			content: `package test

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	GetUser(id int) (*User, error)
}

type Product struct {
	ID    int
	Price float64
}`,
			expectedStructs: []StructInfo{
				{
					Name:    "User",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "ID", Type: "int"},
						{Name: "Name", Type: "string"},
					},
				},
				{
					Name:    "Product",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "ID", Type: "int"},
						{Name: "Price", Type: "float64"},
					},
				},
			},
			expectedInterfaces: []InterfaceInfo{
				{
					Name:    "UserRepository",
					Package: "test",
				},
			},
			expectedStructMap: map[string]string{"User": "User", "Product": "Product"},
		},
		{
			name: "extractTypes_NoTypes",
			content: `package test

func main() {}`,
			expectedStructs:    []StructInfo{},
			expectedInterfaces: []InterfaceInfo{},
			expectedStructMap:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストコードをパースしてASTを作成
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.content, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse test content: %v", err)
			}

			// extractTypes関数をテスト
			structs, interfaces, structMap := extractTypes(f, fset, "test.go", "test")

			// 構造体の検証
			if len(structs) != len(tt.expectedStructs) {
				t.Errorf("extractTypes() returned %d structs, expected %d", len(structs), len(tt.expectedStructs))
			} else {
				for i, expected := range tt.expectedStructs {
					if structs[i].Name != expected.Name {
						t.Errorf("structs[%d].Name = %q, expected %q", i, structs[i].Name, expected.Name)
					}
					if structs[i].Package != expected.Package {
						t.Errorf("structs[%d].Package = %q, expected %q", i, structs[i].Package, expected.Package)
					}
					if len(structs[i].Fields) != len(expected.Fields) {
						t.Errorf("structs[%d] has %d fields, expected %d", i, len(structs[i].Fields), len(expected.Fields))
					} else {
						for j, expectedField := range expected.Fields {
							if structs[i].Fields[j].Name != expectedField.Name {
								t.Errorf("structs[%d].Fields[%d].Name = %q, expected %q", i, j, structs[i].Fields[j].Name, expectedField.Name)
							}
							if structs[i].Fields[j].Type != expectedField.Type {
								t.Errorf("structs[%d].Fields[%d].Type = %q, expected %q", i, j, structs[i].Fields[j].Type, expectedField.Type)
							}
						}
					}
				}
			}

			// インターフェースの検証
			if len(interfaces) != len(tt.expectedInterfaces) {
				t.Errorf("extractTypes() returned %d interfaces, expected %d", len(interfaces), len(tt.expectedInterfaces))
			} else {
				for i, expected := range tt.expectedInterfaces {
					if interfaces[i].Name != expected.Name {
						t.Errorf("interfaces[%d].Name = %q, expected %q", i, interfaces[i].Name, expected.Name)
					}
					if interfaces[i].Package != expected.Package {
						t.Errorf("interfaces[%d].Package = %q, expected %q", i, interfaces[i].Package, expected.Package)
					}
				}
			}

			// structMapの検証
			if len(structMap) != len(tt.expectedStructMap) {
				t.Errorf("extractTypes() returned structMap with %d entries, expected %d", len(structMap), len(tt.expectedStructMap))
			} else {
				for expectedName := range tt.expectedStructMap {
					if _, exists := structMap[expectedName]; !exists {
						t.Errorf("structMap missing key %q", expectedName)
					} else if structMap[expectedName].Name != expectedName {
						t.Errorf("structMap[%q].Name = %q, expected %q", expectedName, structMap[expectedName].Name, expectedName)
					}
				}
			}
		})
	}
}

func TestExtractFunctions(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		expectedFunctions []FuncInfo
		structMap         map[string]*StructInfo
		expectedMethods   map[string][]string // 構造体名 -> メソッド名のリスト
	}{
		{
			name: "extractFunctions_SimpleFunction",
			content: `package test

func Add(a, b int) int {
	return a + b
}`,
			expectedFunctions: []FuncInfo{
				{
					Name:    "Add",
					Package: "test",
					Params: []FieldInfo{
						{Name: "a", Type: "int"},
						{Name: "b", Type: "int"},
					},
					Results: []FieldInfo{
						{Name: "", Type: "int"},
					},
				},
			},
			structMap:       map[string]*StructInfo{},
			expectedMethods: map[string][]string{},
		},
		{
			name: "extractFunctions_Method",
			content: `package test

type User struct {
	Name string
}

func (u *User) GetName() string {
	return u.Name
}

func (u User) SetName(name string) {
	u.Name = name
}`,
			expectedFunctions: []FuncInfo{},
			structMap: map[string]*StructInfo{
				"User": {
					Name:    "User",
					Package: "test",
					Fields: []FieldInfo{
						{Name: "Name", Type: "string"},
					},
					Methods: []FuncInfo{},
				},
			},
			expectedMethods: map[string][]string{
				"User": {"GetName", "SetName"},
			},
		},
		{
			name: "extractFunctions_MixedFunctionsAndMethods",
			content: `package test

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}`,
			expectedFunctions: []FuncInfo{
				{
					Name:    "NewCalculator",
					Package: "test",
					Params:  []FieldInfo{},
					Results: []FieldInfo{
						{Name: "", Type: "*Calculator"},
					},
				},
				{
					Name:    "Multiply",
					Package: "test",
					Params: []FieldInfo{
						{Name: "a", Type: "int"},
						{Name: "b", Type: "int"},
					},
					Results: []FieldInfo{
						{Name: "", Type: "int"},
					},
				},
			},
			structMap: map[string]*StructInfo{
				"Calculator": {
					Name:    "Calculator",
					Package: "test",
					Fields:  []FieldInfo{},
					Methods: []FuncInfo{},
				},
			},
			expectedMethods: map[string][]string{
				"Calculator": {"Add"},
			},
		},
		{
			name: "extractFunctions_NoFunctions",
			content: `package test

type User struct {
	Name string
}`,
			expectedFunctions: []FuncInfo{},
			structMap:         map[string]*StructInfo{},
			expectedMethods:   map[string][]string{},
		},
		{
			name: "extractFunctions_FunctionWithNoParams",
			content: `package test

func GetCurrentTime() time.Time {
	return time.Now()
}`,
			expectedFunctions: []FuncInfo{
				{
					Name:    "GetCurrentTime",
					Package: "test",
					Params:  []FieldInfo{},
					Results: []FieldInfo{
						{Name: "", Type: "time.Time"},
					},
				},
			},
			structMap:       map[string]*StructInfo{},
			expectedMethods: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストコードをパースしてASTを作成
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.content, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse test content: %v", err)
			}

			// extractFunctions関数をテスト
			functions := extractFunctions(f, fset, "test.go", "test", tt.structMap)

			// 関数の検証
			if len(functions) != len(tt.expectedFunctions) {
				t.Errorf("extractFunctions() returned %d functions, expected %d", len(functions), len(tt.expectedFunctions))
			} else {
				for i, expected := range tt.expectedFunctions {
					if functions[i].Name != expected.Name {
						t.Errorf("functions[%d].Name = %q, expected %q", i, functions[i].Name, expected.Name)
					}
					if functions[i].Package != expected.Package {
						t.Errorf("functions[%d].Package = %q, expected %q", i, functions[i].Package, expected.Package)
					}
					// パラメータの検証
					if len(functions[i].Params) != len(expected.Params) {
						t.Errorf("functions[%d] has %d params, expected %d", i, len(functions[i].Params), len(expected.Params))
					} else {
						for j, expectedParam := range expected.Params {
							if functions[i].Params[j].Name != expectedParam.Name {
								t.Errorf("functions[%d].Params[%d].Name = %q, expected %q", i, j, functions[i].Params[j].Name, expectedParam.Name)
							}
							if functions[i].Params[j].Type != expectedParam.Type {
								t.Errorf("functions[%d].Params[%d].Type = %q, expected %q", i, j, functions[i].Params[j].Type, expectedParam.Type)
							}
						}
					}
					// 戻り値の検証
					if len(functions[i].Results) != len(expected.Results) {
						t.Errorf("functions[%d] has %d results, expected %d", i, len(functions[i].Results), len(expected.Results))
					} else {
						for j, expectedResult := range expected.Results {
							if functions[i].Results[j].Type != expectedResult.Type {
								t.Errorf("functions[%d].Results[%d].Type = %q, expected %q", i, j, functions[i].Results[j].Type, expectedResult.Type)
							}
						}
					}
				}
			}

			// メソッドの検証
			for structName, expectedMethodNames := range tt.expectedMethods {
				if s, exists := tt.structMap[structName]; exists {
					if len(s.Methods) != len(expectedMethodNames) {
						t.Errorf("struct %q has %d methods, expected %d", structName, len(s.Methods), len(expectedMethodNames))
					} else {
						for i, expectedMethodName := range expectedMethodNames {
							if s.Methods[i].Name != expectedMethodName {
								t.Errorf("struct %q method[%d].Name = %q, expected %q", structName, i, s.Methods[i].Name, expectedMethodName)
							}
							if s.Methods[i].Receiver != structName {
								t.Errorf("struct %q method[%d].Receiver = %q, expected %q", structName, i, s.Methods[i].Receiver, structName)
							}
						}
					}
				}
			}
		})
	}
}
