package gogroup

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_stack(t *testing.T) {
	a := bufio.NewReader(bytes.NewReader([]byte("12345")))

	fmt.Println(a.ReadLine())

}

func TestParserFuncInfo(t *testing.T) {
	_ff := ff
	fmt.Println(ParserFuncInfo(_ff).String())
}

func Test_loop(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
		i++
	}
}

func Test_parserGoroutineInStack(t *testing.T) {
	a := []byte(" goroutine 9 [running]:")
	fmt.Println(parserGoroutineInStack(a))

	a = []byte("   goroutine 9 [running]:")
	fmt.Println(parserGoroutineInStack(a))

	a = []byte("   goroutine 9111 [running]:")
	fmt.Println(parserGoroutineInStack(a))
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

func Test_recover(t *testing.T) {
	p := recover()
	fmt.Println(p)
}
