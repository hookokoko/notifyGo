package handler

import (
	"context"
	"net/http"
	"notifyGo/internal"
	"notifyGo/internal/service"

	"github.com/gin-gonic/gin"
)

type PushHandler struct {
}

func NewPushHandler() *PushHandler {
	return &PushHandler{}
}

func (p *PushHandler) SendBatch(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "send batch")
}

func (p *PushHandler) Send(ctx *gin.Context) {
	c := service.NewCore()
	err := c.Send(context.TODO(),
		"email",
		"notification",
		internal.EmailTarget{Email: "ch_haokun@163.com"},
		123,
		map[string]interface{}{},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "消息处理处理失败")
	}
	ctx.JSON(http.StatusOK, "send ok")
}
