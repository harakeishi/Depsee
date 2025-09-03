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
			wantFilesPath: []string{},
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
			want:    false,
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
