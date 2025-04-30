package config

import (
	"os"
)

const (
	defaultPort     = 6254
	defaultInterval = 10
)

type Config struct {
	Port               int    `yaml:"port"`
	Interval           int    `yaml:"interval"`
	NodeName           string `yaml:"nodeName"`
	KubeResourcesLabel bool   `yaml:"kubeResourcesLabel"`
}

func (c *Config) SetPort(port int) {
	c.Port = port
}

func (c *Config) SetInterval(interval int) {
	c.Interval = interval
}

func (c *Config) SetNodeName(nodeName string) {
	c.NodeName = nodeName
}

func (c *Config) SetKubeResourcesLabel(kubeResourcesLabel bool) {
	c.KubeResourcesLabel = kubeResourcesLabel
}

func NewDefaultConfig() *Config {
	return &Config{
		Port:     defaultPort,
		Interval: defaultInterval,

		// Set NodeName from `NODE_NAME` env. If not set, leave it empty.
		NodeName:           os.Getenv("NODE_NAME"),
		KubeResourcesLabel: false,
	}
}
