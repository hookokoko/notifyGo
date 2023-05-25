package service

import "notifyGo/internal"

type TargetService struct{}

func NewTargetService() *TargetService {
	return &TargetService{}
}

func (ts *TargetService) GetTarget(targetId int64) []internal.ITarget {
	targets := []internal.ITarget{
		internal.IdTarget{Id: "111"},
		internal.EmailTarget{Email: "ch_hakun@163.com"},
		internal.PhoneTarget{Phone: "+8618800187099"},
	}
	return targets
}
