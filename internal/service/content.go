package service

import "notifyGo/internal"

type ContentService struct{}

func NewContentService() *ContentService {
	return &ContentService{}
}

func (cs *ContentService) GetContent(target internal.Target, templateId uint64, variable map[string]interface{}) string {
	return "hello world"
}
