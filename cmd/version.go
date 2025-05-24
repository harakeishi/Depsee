package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd はversionサブコマンドを表します
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン情報を表示",
	Long: `depseeのバージョン情報を表示します。

例:
  depsee version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("depsee %s\n", GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// versionコマンド専用フラグがあれば、ここで定義
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
