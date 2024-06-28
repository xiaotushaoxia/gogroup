package gogroup

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParserFuncInfo(t *testing.T) {
	_ff := ff
	fi := ParserFuncInfo(_ff)
	if !strings.HasSuffix(fi.FuncName, ".ff") {
		t.Fatal("FuncName not ff")
	}

	if fi.Line != 88 {
		t.Fatal("Line not 88")
	}
}

func Test_parserGoroutineInStack(t *testing.T) {
	a := []byte(" goroutine 9 [running]:")
	if parserGoroutineInStack(a) != "goroutine 9" {
		t.Fatal("not goroutine 9")
	}

	a = []byte("   goroutine 9 [running]:")
	if parserGoroutineInStack(a) != "goroutine 9" {
		t.Fatal("not goroutine 9")
	}

	a = []byte("   goroutine 9111 [running]:")
	if parserGoroutineInStack(a) != "goroutine 9111" {
		t.Fatal("not goroutine 9111")
	}
}

func Test_wait(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		return
	}()
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(time.Microsecond)
			wg.Add(1)
			go func(_i int) {
				defer wg.Done()
				time.Sleep(time.Millisecond * 100)
				fmt.Println(_i)
				return
			}(i)
		}
	}()

	wg.Wait()
}

func Test_stack(t *testing.T) {
	bs := stack(4)
	lines := bytes2lines(bs)
	if len(lines) != 3 {
		t.Fatal("stack size error. 3")
	}

	bs = stack(0)
	lines = bytes2lines(bs)
	if len(lines) != 11 {
		t.Fatal("stack size error. 11")
	}

	bs = stack(5)
	lines = bytes2lines(bs)
	if len(lines) != 11 {
		t.Fatal("stack size error. 11")
	}

	bs = stack(100)
	lines = bytes2lines(bs)
	if len(lines) != 11 {
		t.Fatal("stack size error. 11")
	}

	bs = stack(2)
	lines = bytes2lines(bs)
	if len(lines) != 7 {
		t.Fatal("stack size error. 7")
	}
}

func Test_parserGoroutineInStackFail(t *testing.T) {
	if parserGoroutineInStack([]byte("abc")) != "" {
		t.Fatal("abc not empty")
	}

	if parserGoroutineInStack([]byte("goroutine xxxx")) != "" {
		t.Fatal("abc not empty")
	}
}
