package config

type Config struct {
	Port     int `yaml:"port"`
	Interval int `yaml:"interval"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Port:     6254,
		Interval: 10,
	}
}
