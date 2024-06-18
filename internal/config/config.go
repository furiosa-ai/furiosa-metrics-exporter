package config

type Config struct {
	Port     int
	Interval int
}

func NewDefaultConfig() *Config {
	return &Config{
		Port:     6254,
		Interval: 1,
	}
}
