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
	ShowVersion bool
	LogLevel    string
	LogFormat   string
	TargetDir   string
}

// CLI はCLIアプリケーションを表す
type CLI struct {
	analyzer  *analyzer.Analyzer
	grapher   *graph.Builder
	outputter *output.Generator
}

// NewCLI は新しいCLIインスタンスを作成
func NewCLI() *CLI {
	return &CLI{
		analyzer:  analyzer.New(),
		grapher:   graph.NewBuilder(),
		outputter: output.NewGenerator(),
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

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `depsee: Goコード依存可視化ツール

Usage: depsee analyze <target_dir>

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
	logger.Info("解析開始", "target_dir", config.TargetDir)

	// 解析実行
	result, err := c.analyzer.AnalyzeDir(config.TargetDir)
	if err != nil {
		logger.Error("解析失敗", "error", err, "target_dir", config.TargetDir)
		return err
	}

	// 結果表示
	c.displayResults(result)

	// 依存グラフ構築
	dependencyGraph := c.grapher.BuildDependencyGraph(result)
	c.displayGraph(dependencyGraph)

	// 安定度算出
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

// displayStability は安定度を表示
func (c *CLI) displayStability(stability *graph.StabilityResult) {
	fmt.Println("[info] ノード安定度:")
	for id, s := range stability.NodeStabilities {
		fmt.Printf("  %s: 依存数=%d, 非依存数=%d, 安定度=%.2f\n", id, s.OutDegree, s.InDegree, s.Instability)
	}
}
