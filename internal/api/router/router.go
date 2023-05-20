package router

import (
	"notifyGo/internal/api/handler"

	"github.com/gin-gonic/gin"
)

type MsgPusher struct {
	PushHandler *handler.PushHandler
}

func NewMsgPusher() *MsgPusher {
	return &MsgPusher{PushHandler: handler.NewPushHandler()}
}

func (m *MsgPusher) GetRouter() *gin.Engine {
	router := gin.New()
	//router.Use()

	g := router.Group("/message")
	//g.Use()

	// 发送消息
	g.POST("send", m.PushHandler.SendNew)
	g.POST("sendBatch", m.PushHandler.SendBatch)

	// 查看消息记录
	//g.GET("Send")
	//g.GET("SendBatch")
	return router
}
