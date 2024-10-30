package config

import "os"

type Config struct {
	Port     int    `yaml:"port"`
	Interval int    `yaml:"interval"`
	NodeName string `yaml:"nodeName"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Port:     6254,
		Interval: 10,

		// Set NodeName from `NODE_NAME` env. If not set, try it with OS HostName.
		NodeName: func() string {
			var nodeName string
			var err error

			nodeName = os.Getenv("NODE_NAME")
			if nodeName == "" {
				nodeName, err = os.Hostname()
				if err != nil {
					panic(err)
				}
			}

			return nodeName
		}(),
	}
}
