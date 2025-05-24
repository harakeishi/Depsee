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
