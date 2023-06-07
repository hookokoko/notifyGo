package mq

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Host   []string
	Topics TopicMappings `toml:"topicMappings"`
}

type TopicMappings struct {
	EmailMappings  MsgTypeMappings `toml:"email" json:"email"`
	SmsMappings    MsgTypeMappings `toml:"sms" json:"sms"`
	WechatMappings MsgTypeMappings `toml:"wechat" json:"wechat"`
	PushMappings   MsgTypeMappings `toml:"push" json:"push"`
}

type MsgTypeMappings struct {
	Verification string `toml:"verification" json:"verification"`
	Notification string `toml:"notification" json:"notification"`
	Marketing    string `toml:"marketing" json:"marketing"`
}

func NewConfig(path string) Config {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		panic(err)
	}
	return config
}

func (c Config) AllTopics() []string {
	return []string{
		"HighTopic",
		"MediumTopic",
		"LowTopic",
	}
}
