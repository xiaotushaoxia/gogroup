package gogroup

import (
	"context"
	"fmt"
	"time"
)

func ExampleAllSuccess() {
	err := AllSuccess(context.Background(),
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("timeout")
		},
	)

	if err != nil && err.Error() == "timeout" {
		return
	}
	panic("unexpected")
}

func ExampleGroup() {
	f1 := func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// do job
				time.Sleep(time.Second)
			}
		}
	}
	f2 := func() {
		// sync data
	}

	f3 := func() {
		// update data to server
	}

	var g Group
	g.Go(f1)
	g.GoTk(f2, time.Second) // use default FuncInfo
	g.GoTkWithFuncInfo(f3, time.Second, FuncInfo{
		FuncName:    "updateData",
		File:        "example_test.go",
		Line:        50,
		Description: "update data to server",
	}) // use custom FuncInfo
	g.Wait() // panic in f1 or f2 or f3, g will be canceled

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	g2 := New(ctx) // 10s later, g2 cancel.
	g2.Go(f1)
	g2.GoTk(f2, time.Second)
	g2.Wait()
}
