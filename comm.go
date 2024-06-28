package gogroup

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const microsecondDate = time.DateTime + ".000000"

// GoGroup manages some goroutines.
// When one of them exits, the other will be canceled (ensure that you're listening to `ctx.Done()`)
type GoGroup interface {
	// Go start a goroutine in GoGroup
	// usually a continuously running service, such as an HTTP server
	Go(func(context.Context))

	// GoTk start a goroutine in GoGroup to exec tk every d
	// tkf don't need ctx, I want the execution of tkf cannot be interrupted.
	// it is syntactic sugar of Go(func(context.Context)).
	// same as:
	// 	tkf, d := func() {}, time.Second
	//	g.Go(func(ctx context.Context) {
	//		ticker := time.NewTicker(d)
	//		defer ticker.Stop()
	//		done := ctx.Done()
	//		for {
	//			select {
	//			case <-done:
	//				return
	//			case <-ticker.C:
	//				select {
	//				case <-done:
	//					return
	//				default:
	//					tkf()
	//				}
	//			}
	//		}
	//	})
	GoTk(tkf func(), d time.Duration)

	// GoWithFuncInfo
	// same as Go, but with custom FuncInfo
	// default FuncInfo has File, FuncName and Line
	// Description filed of FuncInfo can be filled in by the user.
	// if f is wrapped, it will lose the actual File, FuncName and Line, so we need custom FuncInfo
	GoWithFuncInfo(f func(context.Context), fi FuncInfo)

	// GoTkWithFuncInfo
	// same as GoTk, but with custom FuncInfo
	GoTkWithFuncInfo(func(), time.Duration, FuncInfo)

	Cancel(error) // cancel ctx of GoGroup. Note: GoGroup won't exit immediately

	Watch() context.Context // return a ctx who will be canceled when GoGroup exit
	Wait()                  // same as <-Watch().Done(), but uninterrupted
	CancelAndWait(error)    // syntactic sugar, same as Cancel(err);Wait()
	Err() error             // the error who cause GoGroup exited
	ExitInfo() *ExitInfo

	// I refer to the 5 functions Watch, Wait, CancelAndWait, Err, and ExitInfo collectively as f5.
	// call Go will panic("group is exited") after f5 called
	// f5 will block until GoGroup exited
}

type GoInfo struct {
	FuncInfo   FuncInfo
	Panic      any
	PanicStack []byte
	ExitTime   time.Time
	StartTime  time.Time
}

type FuncInfo struct {
	FuncName    string
	File        string
	Line        int
	Description string // for display
}

type ExitInfo struct {
	GoInfos              []GoInfo
	CancelByUser         bool
	CancelByRootContext  bool
	CancelBySubGoroutine bool
	Cause                error
	FirstUseLine         string
	FirstUseTime         time.Time
	CancelTime           *time.Time // maybe nil
	FirstGoTime          *time.Time // maybe nil
	ExitTime             *time.Time // maybe nil
}

func (ei ExitInfo) String() string {
	var builder strings.Builder
	builder.WriteString("====Group ExitInfo====\n")
	if ei.FirstUseLine != "" {
		builder.WriteString("FirstUse: ")
		builder.WriteString(ei.FirstUseTime.Format(microsecondDate))
		builder.WriteString(" at: ")
		builder.WriteString(ei.FirstUseLine)
		builder.WriteByte('\n')
	}
	if ei.FirstGoTime != nil {
		builder.WriteString("FirstGoTime: ")
		builder.WriteString(ei.FirstGoTime.Format(microsecondDate))
		builder.WriteByte('\n')
	}
	if ei.Cause == nil {
		builder.WriteString("Cause: <nil> (unexpected: call ExitInfo() before Go?)")
	} else {
		builder.WriteString("Cause: " + ei.Cause.Error())
	}
	builder.WriteByte('\n')
	if ei.CancelTime != nil {
		builder.WriteString("CancelTime: ")
		builder.WriteString(ei.CancelTime.Format(microsecondDate))
		if ei.ExitTime != nil {
			builder.WriteString(", ExitTime: ")
			builder.WriteString(ei.ExitTime.Format(microsecondDate))
		}
		if ei.CancelByUser {
			builder.WriteString(", CancelByUser\n")
		} else if ei.CancelBySubGoroutine {
			builder.WriteString(", CancelBySubGoroutine\n")
		} else if ei.CancelByRootContext {
			builder.WriteString(", CancelByRootContext\n")
		} else {
			builder.WriteString(", UnknownCanceler\n")
		}
	}
	if len(ei.GoInfos) == 0 {
		builder.WriteString("No GoInfo\n")
		return builder.String()
	}
	if len(ei.GoInfos) == 1 {
		builder.WriteString("1 GoInfo\n")
	} else {
		builder.WriteString(strconv.Itoa(len(ei.GoInfos)) + " GoInfos\n")
	}
	for i, gei := range ei.GoInfos {
		builder.WriteString("==")
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString("==\n")
		builder.WriteString("FuncInfo: ")
		builder.WriteString(gei.FuncInfo.String())
		builder.WriteByte('\n')
		builder.WriteString("StartTime: ")
		builder.WriteString(gei.StartTime.Format(microsecondDate))
		builder.WriteString(", ExitTime: ")
		builder.WriteString(gei.ExitTime.Format(microsecondDate))
		builder.WriteByte('\n')
		if gei.Panic == nil {
			continue
		}
		builder.WriteString(fmt.Sprintf("Panic: %v\n", gei.Panic))
		builder.WriteString("PanicStack: ")
		if gei.PanicStack[len(gei.PanicStack)-1] == '\n' {
			builder.Write(gei.PanicStack[:len(gei.PanicStack)-1])
		} else {
			builder.Write(gei.PanicStack)
		}
		if i < len(ei.GoInfos)-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

func (fi FuncInfo) String() string {
	var sb strings.Builder
	sb.Grow(len(fi.FuncName) + len(fi.File) + len(fi.Description) + 17)
	sb.WriteString("func ")
	sb.WriteString(fi.FuncName)

	sb.WriteString(" in file ")
	sb.WriteString(fi.File)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(fi.Line))
	if fi.Description != "" {
		sb.WriteByte('(')
		sb.WriteString(fi.Description)
		sb.WriteByte(')')
	}
	return sb.String()
}

type watcher struct {
	watchOnce sync.Once
	watchCtx  context.Context
	wg        sync.WaitGroup
	waitFunc  func()
}

// Watch return a context who be canceled when group exit
func (w *watcher) watch() context.Context {
	w.watchOnce.Do(func() {
		watchCtx, cancel := context.WithCancel(context.Background())
		go func() {
			if w.waitFunc == nil {
				w.wg.Wait()
			} else {
				w.waitFunc()
			}
			cancel()
		}()
		w.watchCtx = watchCtx
	})
	return w.watchCtx
}

type groupBase struct {
	root        context.Context
	ctx         context.Context
	cancelCause context.CancelCauseFunc
	initOnce    sync.Once
}

func (g *groupBase) initBase() {
	if g.root == nil {
		g.root = context.Background()
	}
	g.ctx, g.cancelCause = context.WithCancelCause(g.root)
}

func groupGoArgs(group GoGroup, args ...any) {
	for i := 0; i < len(args)-1; i++ {
		v0 := args[i]
		switch ff := v0.(type) {
		case func(context.Context):
			group.Go(ff)
		case func():
			i++
			d, ok := args[i].(time.Duration)
			if !ok {
				continue
			}
			group.GoTk(ff, d)
		}
	}
	if last, ok := args[len(args)-1].(func(context.Context)); ok {
		group.Go(last)
	}
	//group.Wait()  // not block default, up to the caller
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
