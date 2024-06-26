package gogroup

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewGroup2(t *testing.T) {
	a := New(context.Background())

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

func TestNewGroup(t *testing.T) {
	a := New(context.Background())

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

func TestNewGroup3(t *testing.T) {
	var b Group
	b.Go(func(ctx context.Context) {
		panic(1)
	})
	b.Go(func(ctx context.Context) {
		time.Sleep(time.Second)
		panic(1)
	})
	b.Wait()
	fmt.Println(b.ExitInfo())
}

func Test_Group_2(t *testing.T) {
	a := New(context.Background())
	a.GoWithFuncInfo(funcCanPanic3, ParserFuncInfo(funcCanPanic3))
	a.Wait()
	fmt.Println(a.ExitInfo())
}

func TestNewAndGo(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Microsecond)

	andGo := NewAndGo(ctx, process, processOnce, time.Second, process, processOnce, time.Second, process, processPanic)
	fmt.Println(andGo)
}

func Test_block_until_cancel(t *testing.T) {
	var g Group
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
