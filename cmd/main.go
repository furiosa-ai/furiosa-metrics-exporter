package main

import (
	"os"

	"github.com/furiosa-ai/furiosa-metric-exporter/internal/cmd"
)

func main() {
	cli := cmd.NewCommand()
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
