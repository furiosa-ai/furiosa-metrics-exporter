package main

import (
	"fmt"
	"os"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/server"
)

func main() {
	cmd := server.NewCommand()
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
