package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/graph"
	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/output"
)

const version = "v0.1.0"

// Config はCLIの設定を表す
type Config struct {
	ShowVersion        bool
	LogLevel           string
	LogFormat          string
	TargetDir          string
	IncludePackageDeps bool
}

// CLI はCLIアプリケーションを表す
type CLI struct {
	analyzer  analyzer.Analyzer
	grapher   graph.GraphBuilder
	outputter output.OutputGenerator
	logger    logger.Logger
}

// NewCLI は新しいCLIインスタンスを作成（デフォルト依存関係）
func NewCLI() *CLI {
	return NewCLIWithDependencies(
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

// NewCLIWithDependencies は依存関係を注入してCLIインスタンスを作成
func NewCLIWithDependencies(
	analyzer analyzer.Analyzer,
	grapher graph.GraphBuilder,
	outputter output.OutputGenerator,
	logger logger.Logger,
) *CLI {
	return &CLI{
		analyzer:  analyzer,
		grapher:   grapher,
		outputter: outputter,
		logger:    logger,
	}
}

// Run はCLIアプリケーションを実行
func (c *CLI) Run(args []string) error {
	config, err := c.parseFlags(args)
	if err != nil {
		return err
	}

	// ログ設定の初期化
	logger.Init(logger.Config{
		Level:  logger.LogLevel(config.LogLevel),
		Format: config.LogFormat,
		Output: os.Stderr,
	})

	if config.ShowVersion {
		fmt.Println("depsee", version)
		return nil
	}

	return c.execute(config)
}

// parseFlags はコマンドライン引数を解析
func (c *CLI) parseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("depsee", flag.ContinueOnError)

	config := &Config{}
	fs.BoolVar(&config.ShowVersion, "version", false, "バージョン情報を表示")
	fs.StringVar(&config.LogLevel, "log-level", "info", "ログレベル (debug, info, warn, error)")
	fs.StringVar(&config.LogFormat, "log-format", "text", "ログフォーマット (text, json)")
	fs.BoolVar(&config.IncludePackageDeps, "include-package-deps", false, "同リポジトリ内のパッケージ間依存関係を解析")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `depsee: Goコード依存可視化ツール

Usage: depsee [options] analyze <target_dir>

Options:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if config.ShowVersion {
		return config, nil
	}

	parsedArgs := fs.Args()
	if len(parsedArgs) < 2 || parsedArgs[0] != "analyze" {
		fs.Usage()
		return nil, fmt.Errorf("invalid arguments")
	}

	config.TargetDir = parsedArgs[1]
	if _, err := os.Stat(config.TargetDir); err != nil {
		return nil, fmt.Errorf("ディレクトリが存在しません: %s", config.TargetDir)
	}

	return config, nil
}

// execute は実際の処理を実行
func (c *CLI) execute(config *Config) error {
	c.logger.Info("解析開始", "target_dir", config.TargetDir)

	// 解析実行
	result, err := c.analyzer.AnalyzeDir(config.TargetDir)
	if err != nil {
		c.logger.Error("解析失敗", "error", err, "target_dir", config.TargetDir)
		return err
	}

	// 結果表示
	c.displayResults(result)

	// 依存グラフ構築
	var dependencyGraph *graph.DependencyGraph
	if config.IncludePackageDeps {
		c.logger.Info("パッケージ間依存関係を含む依存グラフ構築", "include_package_deps", config.IncludePackageDeps)
		dependencyGraph = c.grapher.BuildDependencyGraphWithPackages(result, config.TargetDir)
	} else {
		c.logger.Info("通常の依存グラフ構築", "include_package_deps", config.IncludePackageDeps)
		dependencyGraph = c.grapher.BuildDependencyGraph(result)
	}
	c.displayGraph(dependencyGraph)

	// 不安定度算出
	stability := graph.CalculateStability(dependencyGraph)
	c.displayStability(stability)

	// Mermaid記法の相関図出力
	mermaid := c.outputter.GenerateMermaid(dependencyGraph, stability)
	fmt.Println("[info] Mermaid相関図:")
	fmt.Println(mermaid)

	return nil
}

// displayResults は解析結果を表示
func (c *CLI) displayResults(result *analyzer.AnalysisResult) {
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
func (c *CLI) displayGraph(g *graph.DependencyGraph) {
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
func (c *CLI) displayStability(stability *graph.StabilityResult) {
	fmt.Println("[info] ノード不安定度:")
	for id, s := range stability.NodeStabilities {
		fmt.Printf("  %s: 依存数=%d, 非依存数=%d, 不安定度=%.2f\n", id, s.OutDegree, s.InDegree, s.Instability)
	}
}
