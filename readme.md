# gogroup

`GoGroup` manages some goroutines. when one of them exits, the other will be canceled(ensure that you're listening to `ctx.Done()`). 

## Installation

```bash
go get -u github.com/xiaotushaoxia/gogroup
```

## Whit is GoGroup?

```go
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

```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xiaotushaoxia/gogroup"
)

func main() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()
	g := gogroup.New(ctx) // 3s later, g2 cancel.
	g.Go(runTcpServer)
	g.GoTk(syncData, time.Second)
	g.Wait()
	fmt.Println(g.ExitInfo())
	// output
	//====Group ExitInfo====
	//FirstUse: 2024-06-28 21:23:29.170927 at: main.go:14
	//FirstGoTime: 2024-06-28 21:23:29.170927
	//Cause: context deadline exceeded
	//CancelTime: 2024-06-28 21:23:39.171416, ExitTime: 2024-06-28 21:23:39.179450, CancelByRootContext
	//2 GoExitInfos
	//==1==
	//FuncInfo: func main.syncData in file main.go:47
	//StartTime: 2024-06-28 21:23:29.170927, ExitTime: 2024-06-28 21:23:39.171416
	//==2==
	//FuncInfo: func main.runTcpServer in file main.go:34
	//StartTime: 2024-06-28 21:23:29.170927, ExitTime: 2024-06-28 21:23:39.179450
}

func runTcpServer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// do job
			fmt.Println("new client connected")
			time.Sleep(time.Second)
		}
	}
}

func syncData() {
	fmt.Println("sync data from server")
}
```
