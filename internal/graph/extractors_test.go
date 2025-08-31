package graph

import (
	"testing"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/utils"
)

func TestPackageDependencyExtractor_Extract(t *testing.T) {
	// テスト用の解析結果を作成
	result := &analyzer.Result{
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

func TestCrossPackageDependencyExtractor_Extract(t *testing.T) {
	// テスト用の解析結果を作成
	result := &analyzer.Result{
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

// 共通ユーティリティ関数のテストは utils パッケージで実行されるため、
// ここでは簡単な統合テストのみを行う
func TestUtilsIntegration(t *testing.T) {
	// 標準ライブラリの判定テスト
	if !utils.IsStandardLibrary("fmt") {
		t.Error("Expected fmt to be identified as standard library")
	}

	if utils.IsStandardLibrary("github.com/user/repo") {
		t.Error("Expected github.com/user/repo to not be identified as standard library")
	}

	// パッケージ名抽出テスト
	if utils.ExtractPackageName("github.com/user/repo/pkg") != "pkg" {
		t.Error("Expected package name extraction to work correctly")
	}

	// ローカルパッケージ判定テスト
	if utils.IsLocalPackage("fmt") {
		t.Error("Expected fmt to not be identified as local package")
	}

	if !utils.IsLocalPackage("github.com/user/repo") {
		t.Error("Expected github.com/user/repo to be identified as local package")
	}
}
