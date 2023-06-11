package router

import (
	"notifyGo/internal/api/handler"
	"notifyGo/pkg/mq"

	"github.com/gin-gonic/gin"
)

type MsgPusher struct {
	PushHandler *handler.PushHandler
}

func NewMsgPusher(mqCfg *mq.Config) *MsgPusher {
	return &MsgPusher{PushHandler: handler.NewPushHandler(mqCfg)}
}

func (m *MsgPusher) GetRouter() *gin.Engine {
	router := gin.New()
	//router.Use()

	g := router.Group("/message")
	//g.Use()

	// 发送消息
	g.POST("send", m.PushHandler.Send)
	g.POST("sendBatch", m.PushHandler.SendBatch)

	// 查看消息记录
	//g.GET("Send")
	//g.GET("SendBatch")
	return router
}
