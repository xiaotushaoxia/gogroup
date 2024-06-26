package gogroup

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func NewMini(ctx context.Context) *MiniGroup {
	g := &MiniGroup{groupBase: groupBase{root: ctx}}
	g.init()
	return g
}

type MiniGroup struct {
	groupBase
	watcher
}

func (g *MiniGroup) Go(f func(ctx context.Context)) {
	g.init()
	g.goWithFuncInfo(f, ParserFuncInfo(f))
}

func (g *MiniGroup) GoTk(f func(), d time.Duration) {
	g.init()
	g.goWithFuncInfo(toTkFunc(f, d), ParserFuncInfo(f))
}

func (g *MiniGroup) GoTkWithFuncInfo(f func(), d time.Duration, fi FuncInfo) {
	g.init()
	g.goWithFuncInfo(toTkFunc(f, d), fi)
}

func (g *MiniGroup) GoWithFuncInfo(f func(context.Context), fi FuncInfo) {
	g.init()
	g.goWithFuncInfo(f, fi)
}

func (g *MiniGroup) CancelAndWait(err error) {
	g.init()
	g.cancelCause(err)
	<-g.Watch().Done()
}

func (g *MiniGroup) Cancel(err error) {
	g.init()
	g.cancelCause(err)
}

func (g *MiniGroup) Err() error {
	g.init()
	return context.Cause(g.ctx)
}

func (g *MiniGroup) Wait() {
	g.init()
	<-g.Watch().Done()
}

func (g *MiniGroup) ExitInfo() *ExitInfo {
	g.init()
	select {
	case <-g.Watch().Done():
		return &ExitInfo{Cause: g.Err()}
	default:
		return nil
	}
}

func (g *MiniGroup) init() {
	g.initOnce.Do(func() {
		g.initBase()
	})
}

func (g *MiniGroup) goWithFuncInfo(f func(context.Context), fi FuncInfo) {
	g.wg.Add(1)
	go func() {
		defer g.handleExit(fi)
		f(g.ctx)
	}()
}

func (g *MiniGroup) handleExit(fi FuncInfo) {
	if p := recover(); p != nil {
		st := stack(4)
		gid := parserGoroutineInStack(st)
		tail := gid + fmt.Sprintf(": panic(%v) exit. \nstack: %s", p, string(st))
		g.cancelCause(errors.New(fi.String() + ": " + tail))
	} else {
		g.cancelCause(errors.New(fi.String() + ": exit"))
	}
	g.wg.Done()
}

// NewMiniAndGo
// args should be func(ctx) or func(),time.Duration. like f1(ctx),f2(),1s,f3(ctx),f4(),500ms
func NewMiniAndGo(ctx context.Context, args ...any) GoGroup {
	group := NewMini(ctx)
	groupRunArgs(group, args...)
	return group
}
