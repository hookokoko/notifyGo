package service

import "notifyGo/internal"

type TargetService struct{}

func NewTargetService() *TargetService {
	return &TargetService{}
}

func (ts *TargetService) GetTarget(targetId int64) []internal.Target {
	targets := []internal.Target{
		{UserId: 1111, Email: "ch_haokun@163.com", Phone: "+8613132281931"},
		{UserId: 1112, Email: "hookokoko@126.com", Phone: "+8618800187095"},
	}
	return targets
}
