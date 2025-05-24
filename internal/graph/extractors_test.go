package graph

import (
	"testing"

	"github.com/harakeishi/depsee/internal/analyzer"
)

func TestPackageDependencyExtractor_Extract(t *testing.T) {
	// テスト用の解析結果を作成
	result := &analyzer.AnalysisResult{
		Packages: []analyzer.PackageInfo{
			{
				Name: "pkg1",
				File: "testdata/pkg1/models.go",
				Imports: []analyzer.ImportInfo{
					{Path: "time"},
					{Path: "github.com/example/project/pkg2"},
				},
			},
			{
				Name: "pkg2",
				File: "testdata/pkg2/profile.go",
				Imports: []analyzer.ImportInfo{
					{Path: "time"},
				},
			},
		},
	}

	// 依存グラフを作成
	g := NewDependencyGraph()

	// パッケージ依存関係抽出器を作成
	extractor := NewPackageDependencyExtractor("/test/project")

	// 抽出実行
	extractor.Extract(result, g)

	// パッケージノードが追加されているか確認
	expectedNodes := []NodeID{"package:pkg1", "package:pkg2"}
	for _, nodeID := range expectedNodes {
		if _, ok := g.Nodes[nodeID]; !ok {
			t.Errorf("Expected node %s not found", nodeID)
		}
	}

	// パッケージ間依存関係が追加されているか確認
	if _, ok := g.Edges["package:pkg1"]["package:pkg2"]; !ok {
		t.Error("Expected edge from package:pkg1 to package:pkg2 not found")
	}

	// 標準ライブラリへの依存関係は追加されていないことを確認
	if _, ok := g.Edges["package:pkg1"]["package:time"]; ok {
		t.Error("Unexpected edge to standard library package found")
	}
}

func TestPackageDependencyExtractor_isStandardLibrary(t *testing.T) {
	extractor := NewPackageDependencyExtractor("/test")

	tests := []struct {
		importPath string
		expected   bool
	}{
		{"fmt", true},
		{"time", true},
		{"net/http", true},
		{"github.com/user/repo", false},
		{"example.com/pkg", false},
		{"./relative", false},
		{"../relative", false},
	}

	for _, test := range tests {
		result := extractor.isStandardLibrary(test.importPath)
		if result != test.expected {
			t.Errorf("isStandardLibrary(%s) = %v, expected %v", test.importPath, result, test.expected)
		}
	}
}

func TestPackageDependencyExtractor_isLocalPackage(t *testing.T) {
	extractor := NewPackageDependencyExtractor("/test")

	tests := []struct {
		importPath string
		expected   bool
	}{
		{"fmt", false},
		{"time", false},
		{"net/http", false},
		{"github.com/user/repo", true},
		{"example.com/pkg", true},
		{"./relative", true},
		{"../relative", true},
	}

	for _, test := range tests {
		result := extractor.isLocalPackage(test.importPath)
		if result != test.expected {
			t.Errorf("isLocalPackage(%s) = %v, expected %v", test.importPath, result, test.expected)
		}
	}
}

func TestPackageDependencyExtractor_extractPackageName(t *testing.T) {
	extractor := NewPackageDependencyExtractor("/test")

	tests := []struct {
		importPath string
		expected   string
	}{
		{"github.com/user/repo/pkg", "pkg"},
		{"example.com/project/internal/service", "service"},
		{"pkg", "pkg"},
		{"./relative", "relative"},
		{"../parent", "parent"},
	}

	for _, test := range tests {
		result := extractor.extractPackageName(test.importPath)
		if result != test.expected {
			t.Errorf("extractPackageName(%s) = %s, expected %s", test.importPath, result, test.expected)
		}
	}
}

func TestCrossPackageDependencyExtractor_Extract(t *testing.T) {
	// テスト用の解析結果を作成
	result := &analyzer.AnalysisResult{
		Packages: []analyzer.PackageInfo{
			{
				Name: "main",
				File: "main.go",
				Imports: []analyzer.ImportInfo{
					{Path: "github.com/example/project/pkg/service", Alias: "service"},
					{Path: "fmt", Alias: "fmt"},
				},
			},
			{
				Name: "service",
				File: "pkg/service/service.go",
				Imports: []analyzer.ImportInfo{
					{Path: "github.com/example/project/pkg/model", Alias: "model"},
				},
			},
			{
				Name:    "model",
				File:    "pkg/model/user.go",
				Imports: []analyzer.ImportInfo{},
			},
		},
		Functions: []analyzer.FuncInfo{
			{
				Name:      "main",
				Package:   "main",
				BodyCalls: []string{"service.New", "fmt.Println"},
			},
		},
		Structs: []analyzer.StructInfo{
			{
				Name:    "Service",
				Package: "service",
				Methods: []analyzer.FuncInfo{
					{
						Name:      "Process",
						Package:   "service",
						BodyCalls: []string{"model.User"},
					},
				},
			},
			{
				Name:    "User",
				Package: "model",
			},
		},
	}

	// 依存グラフを作成
	g := NewDependencyGraph()

	// ノードを登録
	g.AddNode(&Node{ID: "main.main", Kind: NodeFunc, Name: "main", Package: "main"})
	g.AddNode(&Node{ID: "service.Service", Kind: NodeStruct, Name: "Service", Package: "service"})
	g.AddNode(&Node{ID: "service.Process", Kind: NodeFunc, Name: "Process", Package: "service"})
	g.AddNode(&Node{ID: "service.New", Kind: NodeFunc, Name: "New", Package: "service"})
	g.AddNode(&Node{ID: "model.User", Kind: NodeStruct, Name: "User", Package: "model"})

	// CrossPackageDependencyExtractorを作成
	extractor := NewCrossPackageDependencyExtractor()

	// 抽出実行
	extractor.Extract(result, g)

	// パッケージ間依存関係が追加されているか確認
	// main.main -> service.New
	if _, ok := g.Edges["main.main"]["service.New"]; !ok {
		t.Error("Expected edge from main.main to service.New not found")
	}

	// service.Process -> model.User
	if _, ok := g.Edges["service.Process"]["model.User"]; !ok {
		t.Error("Expected edge from service.Process to model.User not found")
	}

	// 標準ライブラリへの依存関係は追加されていないことを確認
	if _, ok := g.Edges["main.main"]["fmt.Println"]; ok {
		t.Error("Unexpected edge to standard library function found")
	}
}

func TestCrossPackageDependencyExtractor_isStandardLibrary(t *testing.T) {
	extractor := NewCrossPackageDependencyExtractor()

	tests := []struct {
		importPath string
		expected   bool
	}{
		// 標準ライブラリ
		{"fmt", true},
		{"time", true},
		{"net/http", true},
		{"encoding/json", true},
		{"crypto/sha256", true},
		{"go/ast", true},

		// 外部パッケージ
		{"github.com/user/repo", false},
		{"example.com/pkg", false},
		{"golang.org/x/tools", false},

		// 相対パス
		{"./relative", false},
		{"../relative", false},

		// エッジケース
		{"", false},
		{".", false},

		// 単一パッケージ名（標準ライブラリではない）
		{"mypackage", false},
		{"customlib", false},
	}

	for _, test := range tests {
		result := extractor.isStandardLibrary(test.importPath)
		if result != test.expected {
			t.Errorf("isStandardLibrary(%s) = %v, expected %v", test.importPath, result, test.expected)
		}
	}
}

func TestCrossPackageDependencyExtractor_isLocalPackage(t *testing.T) {
	extractor := NewCrossPackageDependencyExtractor()

	tests := []struct {
		importPath string
		expected   bool
	}{
		// 標準ライブラリ（ローカルではない）
		{"fmt", false},
		{"time", false},
		{"net/http", false},

		// 外部パッケージ（ローカル）
		{"github.com/user/repo", true},
		{"example.com/pkg", true},

		// 相対パス（ローカル）
		{"./relative", true},
		{"../relative", true},

		// エッジケース
		{"", true}, // 標準ライブラリではないのでローカル扱い
		{".", true},

		// 単一パッケージ名（標準ライブラリではないのでローカル扱い）
		{"mypackage", true},
		{"customlib", true},
	}

	for _, test := range tests {
		result := extractor.isLocalPackage(test.importPath)
		if result != test.expected {
			t.Errorf("isLocalPackage(%s) = %v, expected %v", test.importPath, result, test.expected)
		}
	}
}

func TestCrossPackageDependencyExtractor_extractPackageAlias(t *testing.T) {
	extractor := NewCrossPackageDependencyExtractor()

	tests := []struct {
		importPath string
		alias      string
		expected   string
	}{
		{"fmt", "", "fmt"},
		{"github.com/user/repo/pkg/service", "", "service"},
		{"github.com/user/repo/pkg/service", "svc", "svc"},
		{"fmt", ".", "."},
		{"fmt", "_", "_"},
		{"example.com/project/internal/model", "m", "m"},
	}

	for _, test := range tests {
		result := extractor.extractPackageAlias(test.importPath, test.alias)
		if result != test.expected {
			t.Errorf("extractPackageAlias(%s, %s) = %s, expected %s",
				test.importPath, test.alias, result, test.expected)
		}
	}
}
