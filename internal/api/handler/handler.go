package handler

import (
	"net/http"
	"notifyGo/internal"
	"notifyGo/internal/engine/process"

	"github.com/gin-gonic/gin"
)

type PushHandler struct{}

func NewPushHandler() *PushHandler {
	return &PushHandler{}
}

func (p *PushHandler) Send(ctx *gin.Context) {
	task := internal.TaskInfo{
		SendChannel:     "email",
		MessageContent:  "happyhappyhappy",
		MessageReceiver: "chenhaokun",
	}

	proc := process.NewMsgSendProcess()
	err := proc.Process(ctx, task, 0)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "消息处理处理失败")
	}

	ctx.JSON(http.StatusOK, "send")
}

func (p *PushHandler) SendBatch(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "send batch")
}
