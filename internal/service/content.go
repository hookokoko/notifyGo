package service

import (
	"notifyGo/internal"
	"notifyGo/internal/model"
)

type ContentService struct {
	// 是否需要自己管理连接池
	TemplateDAO model.ITemplateDAO
}

func NewContentService() *ContentService {
	return &ContentService{
		TemplateDAO: model.NewITemplateDAO(),
	}
}

func (cs *ContentService) GetContent(target internal.ITarget, templateId int64, variable map[string]interface{}) string {
	content, err := cs.TemplateDAO.GetContent(templateId, "")
	if err != nil {
		return ""
	}
	return content
}
