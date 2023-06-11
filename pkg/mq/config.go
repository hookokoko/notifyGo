package mq

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var changeSignal *sync.Cond
var mu *sync.Mutex

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
	mu = &sync.Mutex{}
	changeSignal = sync.NewCond(mu)
	// 监控文件变化，热加载
	conf.watch(path)
	return conf
}

func (c *Config) watch(path string) {
	dir := filepath.Dir(path)
	file := filepath.Base(path)
	ext := filepath.Ext(path)

	viper.SetConfigName(file[:len(file)-len(ext)])
	viper.SetConfigType(ext[1:])
	viper.AddConfigPath(dir)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		defer func() { mu.Unlock() }()
		mu.Lock()
		changeSignal.Broadcast()
		c.reload(path)
	})
}

func (c *Config) reload(path string) {
	_, err := toml.DecodeFile(path, c)
	if err != nil {
		log.Fatal("重新加载Mq配置失败", err)
	}
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
