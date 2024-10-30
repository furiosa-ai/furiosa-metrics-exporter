package config

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultPort     = 6254
	defaultInterval = 10
)

type Config struct {
	Port     int    `yaml:"port"`
	Interval int    `yaml:"interval"`
	NodeName string `yaml:"nodeName"`
}

func (c *Config) SetFromFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&c.Port, "port", c.Port, "Port to listen on")
	cmd.Flags().IntVar(&c.Interval, "interval", c.Interval, "Interval in seconds")
	cmd.Flags().StringVar(&c.NodeName, "node-name", c.NodeName, "Node name")
}

func NewDefaultConfig() *Config {
	return &Config{
		Port:     defaultPort,
		Interval: defaultInterval,

		// Set NodeName from `NODE_NAME` env. If not set, leave it empty.
		NodeName: os.Getenv("NODE_NAME"),
	}
}
