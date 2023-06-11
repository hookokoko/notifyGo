package server

import (
	"net/http"
	"notifyGo/internal/api/router"
	"notifyGo/pkg/mq"
)

func NewServer(addr string) *http.Server {
	mqCfg := mq.NewConfig("/Users/hooko/GolandProjects/notifyGo/config/kafka_topic.toml")
	pusher := router.NewMsgPusher(mqCfg)
	return &http.Server{
		Addr:    addr,
		Handler: pusher.GetRouter(),
	}
}
