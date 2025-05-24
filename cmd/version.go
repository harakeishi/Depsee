package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ビルド時に設定される変数
var (
	buildCommit = "none"
	buildDate   = "unknown"
)

// GetBuildInfo はビルド情報を返します
func GetBuildInfo() (string, string, string) {
	return GetVersion(), buildCommit, buildDate
}

// versionCmd はversionサブコマンドを表します
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン情報を表示",
	Long: `depseeのバージョン情報を表示します。

例:
  depsee version`,
	Run: func(cmd *cobra.Command, args []string) {
		v, c, d := GetBuildInfo()
		fmt.Printf("depsee %s\n", v)
		fmt.Printf("commit: %s\n", c)
		fmt.Printf("built at: %s\n", d)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// versionコマンド専用フラグがあれば、ここで定義
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
