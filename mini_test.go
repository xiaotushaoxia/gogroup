package gogroup

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewMiniGroup(t *testing.T) {
	a := NewMini(context.Background())
	testGroup(a)
}

func TestNewMiniGroup2(t *testing.T) {
	a := NewMini(context.Background())
	testGroup2(a)
}

func TestNewMiniGroup3(t *testing.T) {
	var b MiniGroup
	testGroup3(&b)
}

func TestMiniGroupGoWithFuncInfo(t *testing.T) {
	a := NewMini(context.Background())
	testGroupGoWithFuncInfo(a)
}

func TestMiniGroupBlockUntilCancel(t *testing.T) {
	var g MiniGroup
	testBlockUntilCancel(t, &g)
}

func TestMiniGroupFully(t *testing.T) {
	var g MiniGroup
	testFully(t, &g)
}

func TestMiniGroupGoAfterWait(t *testing.T) {
	var g MiniGroup
	testGoAfterWait(t, &g)
}

func TestNewMiniAndGo(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Microsecond)

	andGo := NewMiniAndGo(ctx, process, processOnce, processOnce, time.Second, process, processOnce, time.Second, process, processPanic)
	andGo.Wait()
	fmt.Println(andGo.Err())
}
