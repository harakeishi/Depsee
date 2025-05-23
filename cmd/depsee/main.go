package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/harakeishi/depsee/internal/analyzer"
	"github.com/harakeishi/depsee/internal/graph"
)

const version = "v0.1.0"

func main() {
	var (
		showVersion = flag.Bool("version", false, "バージョン情報を表示")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `depsee: Goコード依存可視化ツール\n\n`)
		fmt.Fprintf(os.Stderr, `Usage: depsee analyze <target_dir>\n`)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("depsee", version)
		return
	}

	args := flag.Args()
	if len(args) < 2 || args[0] != "analyze" {
		flag.Usage()
		os.Exit(1)
	}
	targetDir := args[1]
	if _, err := os.Stat(targetDir); err != nil {
		fmt.Fprintf(os.Stderr, "ディレクトリが存在しません: %s\n", targetDir)
		os.Exit(1)
	}

	fmt.Printf("[info] 解析対象ディレクトリ: %s\n", targetDir)

	result, err := analyzer.AnalyzeDir(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] 解析失敗: %v\n", err)
		os.Exit(1)
	}

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

	// 依存グラフ構築・出力
	g := graph.BuildDependencyGraph(result)
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
