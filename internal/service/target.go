package service

import "notifyGo/internal"

type TargetService struct{}

func NewTargetService() *TargetService {
	return &TargetService{}
}

func (ts *TargetService) GetTarget(targetId int64) []internal.ITarget {
	targets := []internal.ITarget{
		internal.IdTarget{"111"},
		internal.EmailTarget{"ch_hakun@163.com"},
		internal.PhoneTarget{"+8618800187099"},
	}
	return targets
}
