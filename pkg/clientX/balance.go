package clientX

import (
	"fmt"
	"sync/atomic"
)

type Balancer[T any] interface {
	Name() string
	Next() (T, error)
}

type RRBalance[T any] struct {
	cnt     int64
	objects []T
}

func (rr *RRBalance[T]) Name() string {
	return "round-robin"
}

func (rr *RRBalance[T]) Next() (T, error) {
	var object T
	if len(rr.objects) == 0 {
		return object, fmt.Errorf("没有可选择的对象")
	}
	cnt := atomic.AddInt64(&rr.cnt, 1)
	index := cnt % int64(len(rr.objects))
	object = rr.objects[index]
	return object, nil
}

type BalanceBuilder[T any] struct {
	name    string
	objects []T
}

func NewBalanceBuilder[T any](name string, objects []T) *BalanceBuilder[T] {
	return &BalanceBuilder[T]{
		name:    name,
		objects: objects,
	}
}

func (bb *BalanceBuilder[T]) Build(name string) Balancer[T] {
	switch name {
	case "round-robin":
		return &RRBalance[T]{cnt: -1, objects: bb.objects}
	case "random":
		//return &RandomBalance[T]{}
	}
	return &RRBalance[T]{cnt: -1, objects: bb.objects}
}
