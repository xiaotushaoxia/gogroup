package gogroup

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func testGroup2(a GoGroup) {

	a.Go(func(ctx context.Context) {
		for i := 0; i < 10; i++ {
			if i == 3 {
				panic(i)
			}
			fmt.Println(i)
		}
	})
	time.Sleep(time.Millisecond)
	a.Go(func(ctx context.Context) {
		for i := 0; i < 10; i++ {
			if i == 8 {
				panic(i)
			}
			fmt.Println(i * 100)
		}
	})
	a.GoTk(func() {
		fmt.Println("tk")
	}, time.Second)

	a.Wait()
	fmt.Println(a.ExitInfo())
}

func testGroup(a GoGroup) {
	a.Go(func(ctx context.Context) {
		for i := 0; i < 10; i++ {
			if i == 3 {
				panic(i)
			}
			fmt.Println(i)
			time.Sleep(time.Second / 2)
		}
	})
	a.Go(func(ctx context.Context) {
		for i := 0; i < 10; i++ {
			if i == 8 {
				panic(i)
			}
			fmt.Println(i * 100)
			time.Sleep(time.Second / 2)
		}
	})
	a.GoTk(func() {
		fmt.Println("tk")
	}, time.Second)

	time.Sleep(time.Second * 5)
	a.Cancel(fmt.Errorf("xxsff"))

	a.Wait()
	fmt.Println(a.ExitInfo())
}

func testGroup3(a GoGroup) {
	a.Go(func(ctx context.Context) {
		panic(1)
	})
	a.Go(func(ctx context.Context) {
		time.Sleep(time.Second)
		panic(1)
	})
	a.Wait()
	fmt.Println(a.ExitInfo())
}

func process(ctx context.Context) {
	time.Sleep(time.Second * 3)
	return
}

func processOnce() {
	fmt.Println("processOnce")
	time.Sleep(time.Second * 1 / 20)
	return
}

func processPanic(ctx context.Context) {
	time.Sleep(time.Second * 1 / 20)
	panic(111)
	return
}
func funcCanPanic(ctx context.Context) {
	panic("xsdd")
}

func funcCanPanic2(ctx context.Context) {
	funcCanPanic(ctx)
}

func funcCanPanic3(ctx context.Context) {
	funcCanPanic2(ctx)
}

func testGroupGoWithFuncInfo(a GoGroup) {
	a.GoWithFuncInfo(funcCanPanic3, ParserFuncInfo(funcCanPanic3))
	a.GoWithFuncInfo(funcCanPanic3, ParserFuncInfo(funcCanPanic3))
	a.GoWithFuncInfo(funcCanPanic3, ParserFuncInfo(funcCanPanic3))
	a.Wait()
	<-a.Watch().Done()
	fmt.Println(a.ExitInfo())
}

func testBlockUntilCancel(t *testing.T, g GoGroup) {
	g.Go(func(ctx context.Context) {
		<-ctx.Done()
		return
	})

	g.Go(func(ctx context.Context) {
		<-ctx.Done()
		return
	})
	go func() {
		time.Sleep(time.Second)
		g.Cancel(fmt.Errorf("xxx"))
	}()
	g.Wait()
	if g.Err().Error() != "xxx" {
		t.Fatal("err not eq")
	}
}

func testFully(t *testing.T, g GoGroup) {
	var a atomic.Int32
	g.Go(func(ctx context.Context) {
		time.Sleep(time.Second * 2)
		return
	})
	g.GoTk(func() {
		a.Add(1)
		fmt.Println(time.Now().UnixMilli(), "GoTk:add 1")
	}, 400*time.Millisecond)
	g.GoTkWithFuncInfo(func() {
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
	g.CancelAndWait(fmt.Errorf("xxx"))
	if g.Err().Error() != "xxx" {
		t.Fatal("err not xxx", g.Err())
	}

	if a.Load() != 4 {
		t.Fatal("a not 4")
	}
}

func TestUnwrapMultiError(t *testing.T) {
	var err error
	if UnwrapMultiError(err) != nil {
		t.Fatal("Unwrap nil not nil")
	}
	err = errors.Join(fmt.Errorf("1"), fmt.Errorf("2"))
	if errs := UnwrapMultiError(err); len(errs) != 2 {
		t.Fatalf("Unwrap 2 return %d", len(errs))
	}
	err = fmt.Errorf("1")
	if errs := UnwrapMultiError(err); len(errs) != 1 {
		t.Fatalf("Unwrap 1 return %d", len(errs))
	}
}

func testGoAfterWait(t *testing.T, a GoGroup) {
	a.Wait()

	var p any

	func() {
		defer func() {
			p = recover()
		}()
		a.Go(func(ctx context.Context) {
			<-ctx.Done()
		})
	}()

	if p == nil {
		t.Fatal("should panic not panic")
	}

}
