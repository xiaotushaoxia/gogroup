package gogroup

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GoGroup
// manages some goroutines.
// when one of them exits, the other goroutines will be canceled (ensure that you're listening to `context.Done()`)
type GoGroup interface {
	Go(func(context.Context))
	GoTk(func(), time.Duration)
	GoWithFuncInfo(func(context.Context), FuncInfo)
	GoTkWithFuncInfo(func(), time.Duration, FuncInfo)
	CancelAndWait(error)
	Cancel(err error)
	Watch() context.Context
	Wait()
	Err() error
	ExitInfo() *ExitInfo
}

type GoExitInfo struct {
	FuncInfo   FuncInfo
	Panic      any
	PanicStack []byte
	Time       time.Time
}

type FuncInfo struct {
	FuncName    string
	File        string
	Line        int
	Description string // for display
}

type ExitInfo struct {
	CancelTime           time.Time
	GoExitInfos          []GoExitInfo
	CancelByUser         bool
	CancelByRootContext  bool
	CancelBySubGoroutine bool
	Cause                error
	FirstUseLine         string
	FirstUseTime         time.Time
	FirstGoTime          time.Time
	ExitTime             time.Time
}

func (ei *ExitInfo) String() string {
	if ei == nil {
		return "ExitInfo<nil>"
	}
	var builder strings.Builder
	builder.WriteString("====Group ExitInfo====\n")
	if ei.FirstUseLine != "" {
		builder.WriteString("FirstUse: ")
		builder.WriteString(ei.FirstUseTime.Format(microsecondDate))
		builder.WriteString(" at: ")
		builder.WriteString(ei.FirstUseLine)
		builder.WriteByte('\n')
	}
	if !ei.FirstGoTime.IsZero() {
		builder.WriteString(fmt.Sprintf("FirstGoTime: %s\n", ei.FirstGoTime.Format(microsecondDate)))
	}
	if !ei.CancelTime.IsZero() {
		builder.WriteString(fmt.Sprintf("Cause: %v\nCancelTime: %s, ExitTime: %s, ", ei.Cause, ei.CancelTime.Format(microsecondDate), ei.ExitTime.Format(microsecondDate)))
		if ei.CancelByUser {
			builder.WriteString("CancelByUser\n")
		} else if ei.CancelBySubGoroutine {
			builder.WriteString("CancelBySubGoroutine\n")
		} else if ei.CancelByRootContext {
			builder.WriteString("CancelByRootContext\n")
		} else {
			builder.WriteString("UnknownCanceler\n")
		}
	} else {
		builder.WriteString(fmt.Sprintf("Cause: %v", ei.Cause))
	}
	if len(ei.GoExitInfos) == 0 {
		builder.WriteString("No GoExitInfo\n")
		return builder.String()
	}
	if len(ei.GoExitInfos) == 1 {
		builder.WriteString("1 GoExitInfo\n")
	} else {
		builder.WriteString(fmt.Sprintf("%d GoExitInfos\n", len(ei.GoExitInfos)))
	}
	for i, gei := range ei.GoExitInfos {
		builder.WriteString(fmt.Sprintf("==%d==\n", i+1))
		builder.WriteString("FuncInfo: " + gei.FuncInfo.String())
		builder.WriteByte('\n')
		builder.WriteString("ExitTime: " + gei.Time.Format(microsecondDate))
		builder.WriteByte('\n')
		if gei.Panic == nil {
			continue
		}
		builder.WriteString(fmt.Sprintf("Panic: %v\n", gei.Panic))
		builder.WriteString("PanicStack: ")
		builder.WriteString(string(gei.PanicStack))
	}
	return builder.String()
}

func (fi FuncInfo) String() string {
	var sb strings.Builder
	sb.Grow(len(fi.FuncName) + len(fi.File) + len(fi.Description) + 16)
	sb.WriteString("func ")
	sb.WriteString(fi.FuncName)

	sb.WriteString(" in file ")
	sb.WriteString(fi.File)
	sb.WriteString(": ")
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
func (w *watcher) Watch() context.Context {
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

func groupRunArgs(group GoGroup, args ...any) {
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
	group.Wait()
}
