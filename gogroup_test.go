package gogroup

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewGroup(t *testing.T) {
	a := New(context.Background())
	testGroup(a)
}

func TestNewGroup2(t *testing.T) {
	a := New(context.Background())
	testGroup2(a)
}

func TestNewGroup3(t *testing.T) {
	var b Group
	testGroup3(&b)
}

func TestGroupGoWithFuncInfo(t *testing.T) {
	a := New(context.Background())
	testGroupGoWithFuncInfo(a)
}

func TestGroupBlockUntilCancel(t *testing.T) {
	var g Group
	testBlockUntilCancel(t, &g)
}

func TestGroupFully(t *testing.T) {
	var g Group
	testFully(t, &g)
}

func TestGroupGoAfterWait(t *testing.T) {
	var g Group
	testGoAfterWait(t, &g)
}

func TestNewAndGo(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Microsecond)

	andGo := NewAndGo(ctx, process, processOnce, processOnce, time.Second, process, processOnce, time.Second, process, processPanic)
	andGo.Wait()
	fmt.Println(andGo.Err())
}
