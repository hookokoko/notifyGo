package handler

import (
	"context"
	"net/http"
	"notifyGo/internal"
	"notifyGo/internal/service"
	"strconv"

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
	params := make(map[string]interface{})
	err := ctx.BindJSON(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "绑定参数出错")
	}
	c := service.NewCore()

	channel := params["channel"].(string)

	var (
		target internal.ITarget
	)

	switch params["channel"].(string) {
	case "email":
		email, ok := params["email"].(string)
		if !ok {
			ctx.JSON(http.StatusBadRequest, "获取邮件地址失败")
			return
		}
		target = internal.EmailTarget{Email: email}
	case "sms":
		phone, ok := params["phone"].(string)
		if !ok {
			ctx.JSON(http.StatusBadRequest, "获取手机号失败")
			return
		}
		target = internal.PhoneTarget{Phone: phone}
	case "push":
		userId, ok := params["userId"].(string)
		if !ok {
			ctx.JSON(http.StatusBadRequest, "获取用户id失败")
			return
		}
		target = internal.IdTarget{Id: userId}
	default:
		ctx.JSON(http.StatusBadRequest, "不支持的发送渠道")
		return
	}

	templateIdStr := params["templateId"].(string)
	templateId, _ := strconv.ParseInt(templateIdStr, 10, 64)
	msgType := params["msgType"].(string)
	err = c.Send(context.TODO(), channel, msgType, target, templateId, map[string]interface{}{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "消息处理处理失败")
		return
	}
	ctx.JSON(http.StatusOK, "send ok")
	return
}
