package gogroup

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewMini(t *testing.T) {
	//mini := NewMini(context.Background())
	var mini MiniGroup
	mini.Go(func(ctx context.Context) {
		time.Sleep(time.Second * 2)
		return
	})
	mini.Go(func(ctx context.Context) {
		time.Sleep(time.Second)
		return
	})
	mini.Go(func(ctx context.Context) {
		panic(1)
		return
	})

	<-mini.Watch().Done()
	fmt.Println(mini.ExitInfo())
}

func TestNewMini2(t *testing.T) {
	var a atomic.Int32
	mini := NewMini(context.Background())
	//var mini MiniGroup
	mini.Go(func(ctx context.Context) {
		time.Sleep(time.Second * 2)
		return
	})
	mini.GoTk(func() {
		a.Add(1)
		fmt.Println(time.Now().UnixMilli(), "GoTk:add 1")
	}, 400*time.Millisecond)
	mini.GoTkWithFuncInfo(func() {
		a.Add(1)
		fmt.Println(time.Now().UnixMilli(), "GoTkWithFuncInfo:add 1")
	}, 400*time.Millisecond, FuncInfo{
		FuncName:    "aaa",
		File:        "bbb",
		Line:        10,
		Description: "xxx",
	})
	time.Sleep(time.Second)
	fmt.Println(time.Now().UnixMilli(), "cancel")
	mini.Cancel(fmt.Errorf("xxx"))
	mini.Wait()
	if mini.Err().Error() != "xxx" {
		t.Fatal("err not xxx", mini.Err())
	}

	if a.Load() != 4 {
		t.Fatal("a not 4")
	}

}
