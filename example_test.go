package gogroup

import (
	"context"
	"fmt"
	"time"
)

func ExampleAllSuccess() {
	err := AllSuccess(context.Background(),
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("timeout")
		},
	)

	if err != nil && err.Error() == "timeout" {
		return
	}
	panic("unexpected")
}

func ExampleGroup() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()
	g := New(ctx) // 3s later, g2 cancel.
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
	//2 GoInfos
	//==1==
	//FuncInfo: func main.syncData in file main.go: 38
	//ExitTime: 2024-06-28 21:23:39.171416
	//==2==
	//FuncInfo: func main.runTcpServer in file main.go: 21
	//ExitTime: 2024-06-28 21:23:39.179450

	var g2 Group // default use context.Background() as root ctx.
	g2.Go(runTcpServer)
	g2.GoTk(syncData, time.Second) // use default FuncInfo
	g2.GoTkWithFuncInfo(uploadData, time.Second, FuncInfo{
		FuncName:    "uploadData",
		File:        "example_test.go",
		Line:        107,
		Description: "update data to server",
	}) // use custom FuncInfo
	g2.Wait() // if panic in syncData, g2 will be canceled; if runTcpServer exit or panic,  g2 will be canceled
}

func ExampleNewAndGo() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()

	group := NewAndGo(ctx,
		runTcpServer,
		syncData, time.Second,
		uploadData, time.Second,
	)
	<-group.Watch().Done() // can be interrupted
	group.Wait()           // can't be interrupted

	fmt.Println(group.ExitInfo()) //
	// output
	//do job
	//do job
	//upload data to server
	//sync data
	//upload data to server
	//sync data
	//do job
	//sync data
	//upload data to server
	// ====Group ExitInfo====
	//FirstUse: 2024-06-27 00:00:00.214867 at: /gogroup/gogroup.go:208
	//FirstGoTime: 2024-06-27 00:00:00.214867
	//Cause: context deadline exceeded
	//CancelTime: 2024-06-27 00:00:00.215397, ExitTime: 2024-06-27 00:00:00.216698, CancelByRootContext
	//3 GoInfos
	//==1==
	//FuncInfo: func github.com/xiaotushaoxia/gogroup.Test_example.func2 in file /src/gogroup/example_test.go: 116
	//ExitTime: 2024-06-27 00:00:00.215397
	//==2==
	//FuncInfo: func github.com/xiaotushaoxia/gogroup.Test_example.func3 in file /src/gogroup/example_test.go: 121
	//ExitTime: 2024-06-27 00:00:00.215397
	//==3==
	//FuncInfo: func github.com/xiaotushaoxia/gogroup.Test_example.func1 in file /src/gogroup/example_test.go: 104
	//ExitTime: 2024-06-27 00:00:00.216698
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

func uploadData() {
	fmt.Println("upload data to server")
}

func syncData() {
	fmt.Println("sync data from server")
}
