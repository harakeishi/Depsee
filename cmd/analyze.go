package cmd

import (
	"github.com/harakeishi/depsee/pkg/depsee"
	"github.com/spf13/cobra"
)

var (
	// analyzeコマンド専用フラグ
	includePackageDeps     bool
	highlightSDPViolations bool
	targetPackages         string
)

// analyzeCmd はanalyzeサブコマンドを表します
var analyzeCmd = &cobra.Command{
	Use:   "analyze [target_dir]",
	Short: "指定されたディレクトリのGoコードを解析",
	Long: `指定されたディレクトリのGoコードを解析し、依存関係を可視化します。

構造体、インターフェース、関数間の依存関係を分析し、
不安定度を計算してMermaid記法での相関図を生成します。

例:
  depsee analyze ./src
  depsee analyze --include-package-deps ./src
  depsee analyze --target-packages main,cmd ./src`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// analyzeコマンド専用フラグ
	analyzeCmd.Flags().BoolVar(&includePackageDeps, "include-package-deps", false, "同リポジトリ内のパッケージ間依存関係を解析")
	analyzeCmd.Flags().BoolVar(&highlightSDPViolations, "highlight-sdp-violations", false, "SDP（Stable Dependencies Principle）違反のエッジを赤色でハイライト")
	analyzeCmd.Flags().StringVar(&targetPackages, "target-packages", "", "解析対象とするパッケージ名をカンマ区切りで指定（例: main,cmd）。指定しない場合は全パッケージが対象")
}

// runAnalyze はanalyzeコマンドの実行ロジック
func runAnalyze(cmd *cobra.Command, args []string) error {
	// 設定を構築
	config := depsee.Config{
		TargetDir:              args[0],
		IncludePackageDeps:     includePackageDeps,
		HighlightSDPViolations: highlightSDPViolations,
		TargetPackages:         targetPackages,
		LogLevel:               GetLogLevel(),
		LogFormat:              GetLogFormat(),
	}

	// Depseeインスタンスを作成して実行
	app := depsee.New()
	return app.Analyze(config)
}
