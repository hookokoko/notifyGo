package mq

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Topic struct {
	Name   string `toml:"name"`
	Weight int    `toml:"weight"`
}

type TopicMapping struct {
	Strategy string  `toml:"strategy"`
	Group    string  `toml:"group"`
	Topics   []Topic `toml:"topic"`
}

type Config struct {
	Host          []string                `toml:"host"`
	TopicMappings map[string]TopicMapping `toml:"topicMappings"`
}

func NewConfig(path string) *Config {
	conf := new(Config)
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		log.Fatal("初始化Mq失败", err)
	}

	return conf
}

func (c *Config) GetTopicsByChannel(channel string) []string {
	topicCfg, ok := c.TopicMappings[channel]
	if !ok {
		log.Printf("找不到该channel的topic：%s", channel)
		return nil
	}
	topics := make([]string, 0, len(topicCfg.Topics))

	for _, item := range topicCfg.Topics {
		topics = append(topics, item.Name)
	}
	return topics
}

func (c *Config) GetHosts() []string {
	return c.Host
}

func (c *Config) GetGroupIdByChannel(channel string) string {
	topicCfg, ok := c.TopicMappings[channel]
	if !ok {
		log.Printf("找不到该channel的topic：%s", channel)
		return ""
	}
	return topicCfg.Group
}

func reload() error {
	return nil
}
