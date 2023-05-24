package service

import (
	"notifyGo/internal"
	"notifyGo/internal/model"
)

type ContentService struct {
	// 是否需要自己管理连接池
	NotifyGoDAO *model.NotifyGoDAO
}

func NewContentService() *ContentService {
	return &ContentService{
		NotifyGoDAO: model.NewNotifyGoDAO(),
	}
}

func (cs *ContentService) GetContent(target internal.Target, templateId int64, variable map[string]interface{}) string {
	// get 语言 by target
	// mock下查看对应target所在的国家
	tpl := model.Template{}
	has, err := cs.NotifyGoDAO.Where("id = ?", templateId).Get(&tpl)
	if err != nil || !has {
		return ""
	}
	return tpl.ChsContent
}
