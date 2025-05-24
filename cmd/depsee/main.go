package main

import (
	"os"

	"github.com/harakeishi/depsee/internal/cli"
)

func main() {
	app := cli.NewCLI()
	if err := app.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
