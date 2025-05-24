package depsee

import (
	"fmt"
	"os"
	"strings"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/graph"
	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/output"
)

// Config は解析の設定を表します
type Config struct {
	TargetDir              string
	IncludePackageDeps     bool
	HighlightSDPViolations bool
	TargetPackages         string
	LogLevel               string
	LogFormat              string
}

// Depsee はメインのアプリケーションロジックを表します
type Depsee struct {
	analyzer  analyzer.Analyzer
	grapher   graph.GraphBuilder
	outputter output.OutputGenerator
	logger    logger.Logger
}

// New は新しいDepseeインスタンスを作成します
func New() *Depsee {
	return NewWithDependencies(
		analyzer.New(),
		graph.NewBuilder(),
		output.NewGenerator(),
		logger.NewLogger(logger.Config{
			Level:  logger.LevelInfo,
			Format: "text",
			Output: os.Stderr,
		}),
	)
}

// NewWithDependencies は依存関係を注入してDepseeインスタンスを作成します
func NewWithDependencies(
	analyzer analyzer.Analyzer,
	grapher graph.GraphBuilder,
	outputter output.OutputGenerator,
	logger logger.Logger,
) *Depsee {
	return &Depsee{
		analyzer:  analyzer,
		grapher:   grapher,
		outputter: outputter,
		logger:    logger,
	}
}

// Analyze は指定された設定でコード解析を実行します
func (d *Depsee) Analyze(config Config) error {
	// ディレクトリの存在確認
	if _, err := os.Stat(config.TargetDir); err != nil {
		return fmt.Errorf("ディレクトリが存在しません: %s", config.TargetDir)
	}

	d.logger.Info("解析開始", "target_dir", config.TargetDir)

	// 解析実行
	var result *analyzer.AnalysisResult
	var err error

	// パッケージフィルタリングが指定されている場合
	if config.TargetPackages != "" {
		targetPackages := parseTargetPackages(config.TargetPackages)
		d.logger.Info("パッケージフィルタリング有効", "target_packages", targetPackages)
		result, err = d.analyzer.AnalyzeDirWithPackageFilter(config.TargetDir, targetPackages)
	} else {
		result, err = d.analyzer.AnalyzeDir(config.TargetDir)
	}

	if err != nil {
		d.logger.Error("解析失敗", "error", err, "target_dir", config.TargetDir)
		return err
	}

	// 結果表示
	d.displayResults(result)

	// 依存グラフ構築
	var dependencyGraph *graph.DependencyGraph
	if config.IncludePackageDeps {
		d.logger.Info("パッケージ間依存関係を含む依存グラフ構築", "include_package_deps", config.IncludePackageDeps)
		dependencyGraph = d.grapher.BuildDependencyGraphWithPackages(result, config.TargetDir)
	} else {
		d.logger.Info("通常の依存グラフ構築", "include_package_deps", config.IncludePackageDeps)
		dependencyGraph = d.grapher.BuildDependencyGraph(result)
	}
	d.displayGraph(dependencyGraph)

	// 不安定度算出
	stability := graph.CalculateStability(dependencyGraph)
	d.displayStability(stability)

	// SDP違反の表示
	if len(stability.SDPViolations) > 0 {
		fmt.Println("[info] SDP違反:")
		for _, violation := range stability.SDPViolations {
			fmt.Printf("  %s (不安定度:%.2f) --> %s (不安定度:%.2f) [違反度:%.2f]\n",
				violation.From, violation.FromInstability,
				violation.To, violation.ToInstability,
				violation.ViolationSeverity)
		}
	} else {
		fmt.Println("[info] SDP違反: なし")
	}

	// Mermaid記法の相関図出力
	var mermaid string
	if config.HighlightSDPViolations {
		// SDP違反ハイライト機能を使用
		mermaid = d.outputter.GenerateMermaidWithOptions(dependencyGraph, stability, true)
	} else {
		mermaid = d.outputter.GenerateMermaid(dependencyGraph, stability)
	}
	fmt.Println("[info] Mermaid相関図:")
	fmt.Println(mermaid)

	return nil
}

// parseTargetPackages はカンマ区切りの文字列をパッケージ名のスライスに変換します
func parseTargetPackages(targetPackages string) []string {
	if targetPackages == "" {
		return nil
	}

	packages := strings.Split(targetPackages, ",")
	result := make([]string, 0, len(packages))

	for _, pkg := range packages {
		trimmed := strings.TrimSpace(pkg)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// displayResults は解析結果を表示
func (d *Depsee) displayResults(result *analyzer.AnalysisResult) {
	fmt.Println("[info] 構造体一覧:")
	for _, s := range result.Structs {
		fmt.Printf("  - %s (package: %s, file: %s)\n", s.Name, s.Package, s.File)
		for _, m := range s.Methods {
			fmt.Printf("      * メソッド: %s\n", m.Name)
		}
	}

	fmt.Println("[info] インターフェース一覧:")
	for _, i := range result.Interfaces {
		fmt.Printf("  - %s (package: %s, file: %s)\n", i.Name, i.Package, i.File)
	}

	fmt.Println("[info] 関数一覧:")
	for _, f := range result.Functions {
		fmt.Printf("  - %s (package: %s, file: %s)\n", f.Name, f.Package, f.File)
	}

	if len(result.Packages) > 0 {
		fmt.Println("[info] パッケージ一覧:")
		for _, p := range result.Packages {
			fmt.Printf("  - %s (file: %s)\n", p.Name, p.File)
			for _, imp := range p.Imports {
				alias := ""
				if imp.Alias != "" {
					alias = " as " + imp.Alias
				}
				fmt.Printf("      * import: %s%s\n", imp.Path, alias)
			}
		}
	}
}

// displayGraph は依存グラフを表示
func (d *Depsee) displayGraph(g *graph.DependencyGraph) {
	fmt.Println("[info] 依存グラフ ノード:")
	for _, n := range g.Nodes {
		fmt.Printf("  - %s (%s)\n", n.ID, n.Name)
	}

	fmt.Println("[info] 依存グラフ エッジ:")
	for from, tos := range g.Edges {
		for to := range tos {
			fmt.Printf("  %s --> %s\n", from, to)
		}
	}
}

// displayStability は不安定度を表示
func (d *Depsee) displayStability(stability *graph.StabilityResult) {
	fmt.Println("[info] ノード不安定度:")
	for id, s := range stability.NodeStabilities {
		fmt.Printf("  %s: 依存数=%d, 非依存数=%d, 不安定度=%.2f\n", id, s.OutDegree, s.InDegree, s.Instability)
	}

	if len(stability.PackageStabilities) > 0 {
		fmt.Println("[info] パッケージ不安定度:")
		for pkg, s := range stability.PackageStabilities {
			fmt.Printf("  %s: 依存数=%d, 非依存数=%d, 不安定度=%.2f\n", pkg, s.OutDegree, s.InDegree, s.Instability)
		}
	}
}
