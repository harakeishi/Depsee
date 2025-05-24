package cmd

import (
	"os"

	"github.com/harakeishi/depsee/internal/logger"
	"github.com/spf13/cobra"
)

var version = "v0.0.5"

var (
	// グローバルフラグ
	logLevel  string
	logFormat string
)

// rootCmd はベースコマンドを表します
var rootCmd = &cobra.Command{
	Use:   "depsee",
	Short: "Goコード依存可視化ツール",
	Long: `depseeは、Goコードの依存関係を解析し、可視化するツールです。

構造体、インターフェース、関数間の依存関係を分析し、
Mermaid記法での相関図を生成します。`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute はルートコマンドを実行します
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// グローバルフラグの定義
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "ログレベル (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "ログフォーマット (text, json)")

	// ローカルフラグの例（必要に応じて）
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig はコンフィグファイルとENV変数を読み込みます
func initConfig() {
	// ログ設定の初期化
	logger.Init(logger.Config{
		Level:  logger.LogLevel(logLevel),
		Format: logFormat,
		Output: os.Stderr,
	})
}

// GetVersion はバージョン文字列を返します
func GetVersion() string {
	return version
}

// GetLogLevel はログレベルを返します
func GetLogLevel() string {
	return logLevel
}

// GetLogFormat はログフォーマットを返します
func GetLogFormat() string {
	return logFormat
}
