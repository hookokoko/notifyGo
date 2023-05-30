package email

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCtx(t *testing.T) {
	ctx := context.TODO()
	timeOutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	wait(timeOutCtx)
}

func wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		fmt.Println("before")
		return ctx.Err()
		//default:
	}
	fmt.Println("after")
	time.Sleep(time.Minute)
	return nil
}
