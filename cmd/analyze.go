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
	excludePackages        string
	excludeDirs            string
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
  depsee analyze -p ./src                           # パッケージ間依存関係を含む
  depsee analyze --include-package-deps ./src       # 上記と同じ
  depsee analyze -t main,cmd ./src                  # 特定パッケージのみ解析
  depsee analyze -e test,mock ./src                 # 特定パッケージを除外
  depsee analyze -d testdata,vendor ./src           # 特定ディレクトリを除外
  depsee analyze -s ./src                           # SDP違反をハイライト
  depsee analyze -p -s -e test ./src                # 複数オプション組み合わせ`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// analyzeコマンド専用フラグ
	analyzeCmd.Flags().BoolVarP(&includePackageDeps, "include-package-deps", "p", false, "同リポジトリ内のパッケージ間依存関係を解析")
	analyzeCmd.Flags().BoolVarP(&highlightSDPViolations, "highlight-sdp-violations", "s", false, "SDP（Stable Dependencies Principle）違反のエッジを赤色でハイライト")
	analyzeCmd.Flags().StringVarP(&targetPackages, "target-packages", "t", "", "解析対象とするパッケージ名をカンマ区切りで指定（例: main,cmd）。指定しない場合は全パッケージが対象")
	analyzeCmd.Flags().StringVarP(&excludePackages, "exclude-packages", "e", "", "解析対象から除外するパッケージ名をカンマ区切りで指定（例: test,mock,vendor）")
	analyzeCmd.Flags().StringVarP(&excludeDirs, "exclude-dirs", "d", "", "解析対象から除外するディレクトリパスをカンマ区切りで指定（例: testdata,vendor,third_party）")
}

// runAnalyze はanalyzeコマンドの実行ロジック
func runAnalyze(cmd *cobra.Command, args []string) error {
	// 設定を構築
	config := depsee.Config{
		TargetDir:              args[0],
		IncludePackageDeps:     includePackageDeps,
		HighlightSDPViolations: highlightSDPViolations,
		TargetPackages:         targetPackages,
		ExcludePackages:        excludePackages,
		ExcludeDirs:            excludeDirs,
		LogLevel:               GetLogLevel(),
		LogFormat:              GetLogFormat(),
	}

	// Depseeインスタンスを作成して実行
	app := depsee.New()
	return app.Analyze(config)
}
