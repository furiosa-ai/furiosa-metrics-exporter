package main

import (
	"os"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/cmd"
)

func main() {
	cli := cmd.NewCommand()
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
