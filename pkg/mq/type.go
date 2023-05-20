package mq

type Config struct {
	host []string
}

func NewConfig(host []string) Config {
	return Config{host: host}
}
