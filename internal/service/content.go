package service

type ContentService struct{}

func NewContentService() *ContentService {
	return &ContentService{}
}

func (cs *ContentService) GetContent() string {
	return "hello world"
}
