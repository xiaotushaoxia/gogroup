package gogroup

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var (
	doneCtx, _ = context.WithDeadline(context.Background(), time.Time{})
	ctxErr     = context.DeadlineExceeded
)

func TestFirstSuccess1(t *testing.T) {
	err := FirstSuccess(context.Background(),
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
	)
	if err != nil {
		t.Fatal("err not nil")
	}
}
func TestFirstSuccess2(t *testing.T) {
	err := FirstSuccess(context.Background(),
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
	)
	if err == nil {
		t.Fatal("err is nil")
	}
}

func TestFirstSuccess3(t *testing.T) {
	err := FirstSuccess(doneCtx,
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
	)
	if err == nil {
		t.Fatal("err is nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
}
func TestFirstSuccess4(t *testing.T) {
	err := FirstSuccess(doneCtx,
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
	)
	if err == nil {
		t.Fatal("err is nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
}

func ff(ctx context.Context) (int, error) {
	return 0, fmt.Errorf("xx")
}

func ff2(ctx context.Context) (int, error) {
	time.Sleep(time.Second)
	return 2, nil
}

func ff3(ctx context.Context) (int, error) {
	return 3, nil
}

func TestFirstSuccessWithResult1(t *testing.T) {
	result, err := FirstSuccessWithResult(context.Background(), ff, ff, ff, ff2,
		func(ctx context.Context) (int, error) {
			return 3, nil
		},
	)
	if err != nil {
		t.Fatal("err != nil ")
	}
	if result != 3 {
		t.Fatal("result != 3")
	}
}

func TestFirstSuccessWithResult2(t *testing.T) {
	result, err := FirstSuccessWithResult(context.Background(), ff, ff, ff)
	if err == nil {
		t.Fatal("err == nil ")
	}
	errors := UnwrapMultiError(err)
	for _, er := range errors {
		if er.Error() != "xx" {
			t.Fatal("err not " + "xx" + ", " + er.Error())
		}
	}
	if result != 0 {
		t.Fatal("result != 0")
	}
}

func TestFirstSuccessWithResult3(t *testing.T) {
	result, err := FirstSuccessWithResult(doneCtx, ff, ff, ff, ff2, func(ctx context.Context) (int, error) {
		return 3, nil
	})
	if err == nil {
		t.Fatal("err nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
	if result != 0 {
		t.Fatal("result not empty")
	}
}

func TestFirstSuccessWithResult4(t *testing.T) {
	result, err := FirstSuccessWithResult(doneCtx, ff, ff, ff)
	if err == nil {
		t.Fatal("err nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
	if result != 0 {
		t.Fatal("result not empty")
	}
}

func TestAllSuccessWithResult1(t *testing.T) {
	result, err := AllSuccessWithResult(context.Background(), ff2, ff3, ff)
	if len(result) != 0 {
		t.Fatal("result not 0")
	}
	if err == nil {
		t.Fatal("err == nil ")
	}
	if err.Error() != "xx" {
		t.Fatal("err not xx")
	}
}
func TestAllSuccessWithResult2(t *testing.T) {
	result, err := AllSuccessWithResult(context.Background(), ff2, ff3)
	if err != nil {
		t.Fatal("err not nil")
	}
	if len(result) != 2 || result[0]+result[1] != 5 {
		t.Fatal("result not right")
	}
}
func TestAllSuccessWithResult3(t *testing.T) {
	result, err := AllSuccessWithResult(doneCtx, ff2, ff3, ff)
	if err == nil {
		t.Fatal("err nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
	if len(result) != 0 {
		t.Fatal("result not empty")
	}
}

func TestAllSuccessWithResult4(t *testing.T) {
	result, err := AllSuccessWithResult(doneCtx, ff2, ff3)
	if err == nil {
		t.Fatal("err nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
	if len(result) != 0 {
		t.Fatal("result not empty")
	}
}

func TestAllSuccess(t *testing.T) {
	err := AllSuccess(context.Background(),
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
	)
	if err == nil {
		t.Fatal("err nil")
	}
}

func TestAllSuccess2(t *testing.T) {
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
	)
	if err != nil {
		t.Fatal("err not nil")
	}
}

func TestAllSuccess3(t *testing.T) {
	err := AllSuccess(doneCtx,
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("x")
		},
	)
	if err == nil {
		t.Fatal("err nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
}

func TestAllSuccess4(t *testing.T) {
	err := AllSuccess(doneCtx,
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
	)
	if err == nil {
		t.Fatal("err nil")
	}
	if err.Error() != ctxErr.Error() {
		t.Fatal("err not " + ctxErr.Error())
	}
}

func Test_nil(t *testing.T) {
	var a atomic.Value
	t2, ok := a.Load().(time.Time)
	fmt.Println(t2, ok)
}
