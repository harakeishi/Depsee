package cli

import (
	"fmt"
	"os"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/graph"
	"github.com/harakeishi/depsee/internal/logger"
	"github.com/harakeishi/depsee/internal/output"
	"github.com/spf13/cobra"
)

const version = "v0.1.0"

// CLI はCLIアプリケーションを表す
type CLI struct {
	analyzer  analyzer.Analyzer
	grapher   graph.GraphBuilder
	outputter output.OutputGenerator
	logger    logger.Logger
	rootCmd   *cobra.Command
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
	cli := &CLI{
		analyzer:  analyzer,
		grapher:   grapher,
		outputter: outputter,
		logger:    logger,
	}

	cli.setupCommands()
	return cli
}

// setupCommands はCobraコマンドを設定
func (c *CLI) setupCommands() {
	c.rootCmd = &cobra.Command{
		Use:   "depsee",
		Short: "Goコード依存可視化ツール",
		Long:  "Goコードの依存関係を解析し、可視化するツールです。",
	}

	// グローバルフラグ
	c.rootCmd.PersistentFlags().String("log-level", "info", "ログレベル (debug, info, warn, error)")
	c.rootCmd.PersistentFlags().String("log-format", "text", "ログフォーマット (text, json)")

	// versionコマンド
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("depsee %s\n", version)
		},
	}

	// analyzeコマンド
	analyzeCmd := &cobra.Command{
		Use:   "analyze [target_dir]",
		Short: "指定されたディレクトリのGoコードを解析",
		Long:  "指定されたディレクトリのGoコードを解析し、依存関係を可視化します。",
		Args:  cobra.ExactArgs(1),
		RunE:  c.runAnalyze,
	}

	// analyzeコマンド専用フラグ
	analyzeCmd.Flags().Bool("include-package-deps", false, "同リポジトリ内のパッケージ間依存関係を解析")

	c.rootCmd.AddCommand(versionCmd)
	c.rootCmd.AddCommand(analyzeCmd)
}

// Run はCLIアプリケーションを実行
func (c *CLI) Run(args []string) error {
	c.rootCmd.SetArgs(args)
	return c.rootCmd.Execute()
}

// runAnalyze はanalyzeコマンドの実行
func (c *CLI) runAnalyze(cmd *cobra.Command, args []string) error {
	// グローバルフラグの取得
	logLevel, _ := cmd.Root().PersistentFlags().GetString("log-level")
	logFormat, _ := cmd.Root().PersistentFlags().GetString("log-format")

	// ローカルフラグの取得
	includePackageDeps, _ := cmd.Flags().GetBool("include-package-deps")

	targetDir := args[0]

	// ログ設定の初期化
	logger.Init(logger.Config{
		Level:  logger.LogLevel(logLevel),
		Format: logFormat,
		Output: os.Stderr,
	})

	// ディレクトリの存在確認
	if _, err := os.Stat(targetDir); err != nil {
		return fmt.Errorf("ディレクトリが存在しません: %s", targetDir)
	}

	return c.execute(targetDir, includePackageDeps)
}

// execute は実際の処理を実行
func (c *CLI) execute(targetDir string, includePackageDeps bool) error {
	c.logger.Info("解析開始", "target_dir", targetDir)

	// 解析実行
	result, err := c.analyzer.AnalyzeDir(targetDir)
	if err != nil {
		c.logger.Error("解析失敗", "error", err, "target_dir", targetDir)
		return err
	}

	// 結果表示
	c.displayResults(result)

	// 依存グラフ構築
	var dependencyGraph *graph.DependencyGraph
	if includePackageDeps {
		c.logger.Info("パッケージ間依存関係を含む依存グラフ構築", "include_package_deps", includePackageDeps)
		dependencyGraph = c.grapher.BuildDependencyGraphWithPackages(result, targetDir)
	} else {
		c.logger.Info("通常の依存グラフ構築", "include_package_deps", includePackageDeps)
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
