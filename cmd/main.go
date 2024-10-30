package main

import (
	"os"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/server"
)

func main() {
	cmd := server.NewCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
