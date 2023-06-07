package mq

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func Test_Config(t *testing.T) {
	cfg := NewConfig("../../config/kafka_topic.toml")
	log.Printf("%+v\n", cfg)
	var tps map[string]map[string]string
	tB, err := json.Marshal(cfg.Topics)
	assert.Nil(t, err)
	err = json.Unmarshal(tB, &tps)
	assert.Nil(t, err)
	log.Printf("%+v\n", tps)

	match, ok := tps["email1"]["verification"]
	if !ok {
		log.Fatal("找不到对应topic")
	}
	log.Println(match)
}
