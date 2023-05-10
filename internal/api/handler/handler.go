package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type PushHandler struct{}

func NewPushHandler() *PushHandler {
	return &PushHandler{}
}

func (p *PushHandler) Send(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "send")
}

func (p *PushHandler) SendBatch(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "send batch")
}
