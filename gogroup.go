package gogroup

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	groupStateInit    int32 = 0
	groupStateRunning int32 = 1
	groupStateExited  int32 = 2

	cancelFlagInit                 int32 = 0
	cancelFlagCancelByUser         int32 = 1
	cancelFlagCancelByRootContext  int32 = 2
	cancelFlagCancelBySubGoroutine int32 = 3

	microsecondDate = time.DateTime + ".000000"
)

func New(ctx context.Context) *Group {
	g := &Group{groupBase: groupBase{root: ctx}}
	g.init()
	return g
}

// Group can be use only once
type Group struct {
	groupBase
	watcher

	state atomic.Int32

	exitsM sync.Mutex
	exits  []GoExitInfo

	cancelAt   atomic.Value // time.Time
	cancelFlag atomic.Int32 // init:0,cancel by user 1;cancel by sub goroutine:2; cancel by root ctx: 3

	watchRootOnce sync.Once

	firstUseLine string
	firstUseTime time.Time
	exitTime     atomic.Value
	firstGoTime  atomic.Value
}

func (g *Group) Go(f func(context.Context)) {
	g.init()
	g.goWithFuncInfo(f, ParserFuncInfo(f))
}

func (g *Group) GoTk(f func(), d time.Duration) {
	g.init()
	g.goWithFuncInfo(toTkFunc(f, d), ParserFuncInfo(f))
}

// GoWithFuncInfo
// default FuncInfo has `FuncName File Line`
// use `fi` to support add Description for FuncInfo. and after wrapped, `FuncName File Line` is not original anymore
func (g *Group) GoWithFuncInfo(f func(context.Context), fi FuncInfo) {
	g.init()
	g.goWithFuncInfo(f, fi)
}

func (g *Group) GoTkWithFuncInfo(f func(), d time.Duration, fi FuncInfo) {
	g.goWithFuncInfo(toTkFunc(f, d), fi)
}

func (g *Group) CancelAndWait(err error) {
	g.init()
	g.cancelByCanceler(err, cancelFlagCancelByUser)
	<-g.Watch().Done()
}

func (g *Group) Cancel(err error) {
	g.init()
	g.cancelByCanceler(err, cancelFlagCancelByUser)
}

func (g *Group) Wait() {
	g.init()
	<-g.Watch().Done()
}

// Err return the error who cause ctx cancel
func (g *Group) Err() error {
	g.init()
	return context.Cause(g.ctx)
}

func (g *Group) ExitInfo() *ExitInfo {
	g.init()
	if g.state.Load() != groupStateExited {
		return nil
	}
	a, ok := g.cancelAt.Load().(time.Time)
	if !ok { // should not happen
		return nil
	}
	ft, ok := g.firstGoTime.Load().(time.Time)
	if !ok { // should not happen
		return nil
	}
	et, ok := g.exitTime.Load().(time.Time)
	if !ok { // should not happen
		return nil
	}
	v := &ExitInfo{
		CancelTime:   a,
		FirstUseTime: g.firstUseTime,
		ExitTime:     et,
		FirstGoTime:  ft,
		Cause:        g.Err(),
		FirstUseLine: g.firstUseLine,
	}
	switch g.cancelFlag.Load() {
	case cancelFlagCancelByUser:
		v.CancelByUser = true
	case cancelFlagCancelByRootContext:
		v.CancelByRootContext = true
	case cancelFlagCancelBySubGoroutine:
		v.CancelBySubGoroutine = true
	}
	g.exitsM.Lock()
	v.GoExitInfos = append(make([]GoExitInfo, 0, len(g.exits)), g.exits...)
	g.exitsM.Unlock()
	return v
}

func (g *Group) init() {
	g.initOnce.Do(func() {
		g.initBase()
		g.waitFunc = g.waitAndSetExit
		g.firstUseTime = time.Now()
		g.firstUseLine = getCallerLine(7)
	})
}

func (g *Group) panicIfExited() {
	if v := g.state.Load(); v == groupStateExited {
		panic("group is exited")
	}
}

func (g *Group) goWithFuncInfo(f func(context.Context), fi FuncInfo) {
	g.panicIfExited()
	g.watchRootContext()
	g.state.CompareAndSwap(groupStateInit, groupStateRunning)
	if g.firstGoTime.Load() == nil {
		g.firstGoTime.CompareAndSwap(nil, time.Now())
	}
	g.wg.Add(1)
	go func() {
		defer g.handleExit(fi)
		f(g.ctx)
	}()
}

func (g *Group) waitAndSetExit() {
	g.wg.Wait()
	g.state.Store(groupStateExited)
	g.exitTime.Store(time.Now())
}

func (g *Group) watchRootContext() {
	g.watchRootOnce.Do(func() {
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			select {
			case <-g.root.Done():
			case <-g.ctx.Done():
				if g.root.Err() == nil {
					return
				}
			}
			if g.cancelFlag.CompareAndSwap(cancelFlagInit, cancelFlagCancelByRootContext) {
				g.cancelAt.CompareAndSwap(nil, time.Now())
			}
		}()
	})
}

func (g *Group) handleExit(fi FuncInfo) {
	p := recover()
	ei, err := getGoExitInfo(fi, p)
	g.exitsM.Lock()
	g.exits = append(g.exits, ei)
	// cancel must be protected by exitsM. otherwise g may be canceled by other ei
	g.cancelByCanceler(err, cancelFlagCancelBySubGoroutine)
	g.exitsM.Unlock()
	g.wg.Done()
}

func (g *Group) cancelByCanceler(err error, canceler int32) *Group {
	if g.cancelFlag.CompareAndSwap(cancelFlagInit, canceler) {
		g.cancelAt.CompareAndSwap(nil, time.Now())
		g.cancelCause(err)
	}
	return g
}

// NewAndGo
// args should be func(ctx) or func(),time.Duration. like f1(ctx),f2(),1s,f3(ctx),f4(),500ms
func NewAndGo(ctx context.Context, args ...any) GoGroup {
	group := New(ctx)
	groupRunArgs(group, args...)
	return group
}
