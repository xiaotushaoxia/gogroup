package gogroup

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
)

// stack returns a formatted stack trace, skipping the most recent 'skip' calls
//
//	 example:
//		stack: goroutine 50 [running]:
//	1	runtime/debug.Stack()
//	1	        xxxxxxxxxxx/src/runtime/debug/stack.go:24 +0x5e
//	2	main.Stack(0xc000cfdd60?)
//	2	        xxxxxxxxxxx/utils.go:499 +0x25
//	3	main.(*MiniGroup).go1.func1.1()
//	3	        xxxxxxxxxxx/utils.go:454 +0x65
//	4	panic({0xa6f100?, 0xf8d898?})
//	4	       xxxxxxxxxxx//src/runtime/panic.go:914 +0x21f
//	5	main.main.func3()
//	5	       xxxxxxxxxxx/main.go:71 +0x45
//	6	main.(*MiniGroup).GoTk.toTkFunc.func1()
//	6	        xxxxxxxxxxx/utils.go:491 +0x87
//	7	main.(*MiniGroup).go1.func1()
//	7	        xxxxxxxxxxx/utils.go:459 +0x6c
//	8	created by main.(*MiniGroup).go1 in goroutine 1
//	8	        xxxxxxxxxxx/utils.go:451 +0xc5
func stack(n int) []byte {
	lines := bytes2lines(debug.Stack())
	if n == 0 || len(lines) <= n*2+1 { // 如果skip太多反而一条都不跳过
		return bytes.Join(lines, []byte{})
	}
	lines = append([][]byte{lines[0]}, lines[2*n+1:]...)
	return bytes.Join(lines, []byte{})
}

func bytes2lines(bs []byte) [][]byte {
	reader := bufio.NewReader(bytes.NewReader(bs))
	var lines = make([][]byte, 0)
	for {
		readBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// todo not sure what to do... maybe just ignore it
			break
		}
		lines = append(lines, readBytes)
	}
	return lines
}

func getCallerLine(skip int) string {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(skip, pcs)
	pcs = pcs[:n]

	frames := runtime.CallersFrames(pcs)
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}

func toTkFunc(f func(), d time.Duration) func(context.Context) {
	return func(ctx context.Context) {
		tk := time.NewTicker(d)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			default: // if f cost longer than d, may not exit forever
				select {
				case <-ctx.Done():
					return
				case <-tk.C:
					f()
				}
			}
		}
	}
}

func ParserFuncInfo(f any) FuncInfo {
	pf := reflect.ValueOf(f).Pointer()
	funcForPC := runtime.FuncForPC(pf)
	file, line := funcForPC.FileLine(pf)
	return FuncInfo{
		FuncName: funcForPC.Name(),
		File:     file,
		Line:     line,
	}
}

func parserGoroutineInStack(bs []byte) string {
	flag := "goroutine "
	indexAny := bytes.Index(bs, []byte(flag))
	if indexAny == -1 {
		return ""
	}
	bs2 := bs[indexAny+len(flag):]
	indexByte := bytes.IndexByte(bs2, ' ')
	if indexAny == -1 {
		return ""
	}
	return flag + string(bs2[:indexByte])
}

func getGoExitInfo(fi FuncInfo, panicValue any) (GoExitInfo, error) {
	var tail string
	ei := GoExitInfo{FuncInfo: fi, Time: time.Now()}
	if panicValue != nil {
		st := stack(5)
		gid := parserGoroutineInStack(st)
		tail = gid + fmt.Sprintf(": panic(%v) exit", panicValue)
		ei.Panic, ei.PanicStack = panicValue, st
	} else {
		tail = "exit"
	}
	return ei, errors.New(fi.String() + ": " + tail)
}

func UnwrapMultiError(err error) (errors []error) {
	if err == nil {
		return nil
	}
	if i, ok := err.(interface{ Unwrap() []error }); ok {
		return i.Unwrap()
	}
	if i, ok := err.(interface{ WrappedErrors() []error }); ok {
		return i.WrappedErrors()
	}
	if i, ok := err.(interface{ Errors() []error }); ok {
		return i.Errors()
	}
	return []error{err}
}
