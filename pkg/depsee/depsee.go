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
	ExcludePackages        string
	ExcludeDirs            string
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
	var err error

	// フィルタリング設定をパース :FIXME: cobraの機能でパースできるか確認する
	targetPackagesList := parseTargetPackages(config.TargetPackages)
	excludePackagesList := parseTargetPackages(config.ExcludePackages)
	excludeDirsList := parseTargetPackages(config.ExcludeDirs)
	filters := analyzer.Filters{
		TargetPackages:  targetPackagesList,
		ExcludePackages: excludePackagesList,
		ExcludeDirs:     excludeDirsList,
	}
	d.analyzer.SetFilters(filters)
	d.analyzer.ListTartgetFiles(config.TargetDir)
	err = d.analyzer.Analyze()
	if err != nil {
		d.logger.Error("解析失敗", "error", err, "target_dir", config.TargetDir)
		return err
	}

	// 解析結果をエクスポート
	result := d.analyzer.ExportResult()
	// ここまでリファクタ済み

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
