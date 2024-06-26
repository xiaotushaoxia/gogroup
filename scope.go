package gogroup

import (
	"context"
	"errors"
)

type ResultErrorFunc[T any] func(context.Context) (T, error)
type ErrorFunc func(context.Context) error

func AllSuccessWithResult[T any](ctx context.Context, fs ...ResultErrorFunc[T]) (ts []T, err error) {
	return allSuccessWithResult(ctx, fs, fs2fis(fs))
}

func FirstSuccessWithResult[T any](ctx context.Context, fs ...ResultErrorFunc[T]) (t T, err error) {
	return firstSuccessWithResult(ctx, fs, fs2fis(fs))
}

func AllSuccess(ctx context.Context, fs ...ErrorFunc) error {
	fs2, fis := convertErrorFunc2ResultErrorFunc(fs...)
	_, err := allSuccessWithResult(ctx, fs2, fis)
	return err
}

func FirstSuccess(ctx context.Context, fs ...ErrorFunc) error {
	fs2, fis := convertErrorFunc2ResultErrorFunc(fs...)
	_, err := firstSuccessWithResult(ctx, fs2, fis)
	return err
}

func allSuccessWithResult[T any](ctx context.Context, fs []ResultErrorFunc[T], fis []FuncInfo) (ts []T, err error) {
	ct, ce, g, cleanup := groupCall(ctx, fs, fis)
	defer cleanup()
	done := g.Watch().Done()
	done2 := ctx.Done()
	for i := 0; i < len(fs); i++ {
		select {
		case <-done:
			return nil, g.Err()
		case <-done2:
			return nil, ctx.Err()
		default:
			select {
			case <-done:
				return nil, g.Err()
			case <-done2:
				return nil, ctx.Err()
			case e := <-ce:
				return nil, e
			case t := <-ct:
				ts = append(ts, t)
			}
		}
	}
	return
}

func firstSuccessWithResult[T any](ctx context.Context, fs []ResultErrorFunc[T], fis []FuncInfo) (t T, err error) {
	ct, ce, g, cleanup := groupCall(ctx, fs, fis)
	defer cleanup()
	var errs []error
	done := g.Watch().Done()
	done2 := ctx.Done()
	for i := 0; i < len(fs); i++ {
		select {
		case <-done:
			return t, g.Err()
		case <-done2:
			return t, ctx.Err()
		default:
			select {
			case <-done:
				return t, g.Err()
			case <-done2:
				return t, ctx.Err()
			case t = <-ct:
				return t, nil
			case e := <-ce:
				errs = append(errs, e)
			}
		}
	}
	err = errors.Join(errs...)
	return
}

func groupCall[T any](ctx context.Context, fs []ResultErrorFunc[T], fis []FuncInfo) (chan T, chan error, GoGroup, func()) {
	ct := make(chan T, len(fs))     // can't be 1, if group exit when exec `ct <- t2`, maybe deadlock
	ce := make(chan error, len(fs)) // can't be 1, if group exit when exec `ce <- er`, maybe deadlock

	g := NewMini(ctx)
	for i, _f := range fs {
		f := _f
		g.GoWithFuncInfo(
			func(_ctx context.Context) {
				if t2, er := f(_ctx); er == nil {
					ct <- t2
				} else {
					ce <- er
				}
				<-_ctx.Done()
			},
			fis[i],
		)
	}
	return ct, ce, g, func() {
		g.CancelAndWait(context.Canceled)
	}
}

func convertErrorFunc2ResultErrorFunc(fs ...ErrorFunc) ([]ResultErrorFunc[struct{}], []FuncInfo) {
	rfs := make([]ResultErrorFunc[struct{}], 0, len(fs))
	fis := make([]FuncInfo, 0, len(fs))
	for _, f := range fs {
		_f := f
		rfs = append(rfs, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, _f(ctx)
		})
		fis = append(fis, ParserFuncInfo(_f))
	}
	return rfs, fis
}

func fs2fis[T any](fs []T) []FuncInfo {
	var fis = make([]FuncInfo, 0, len(fs))
	for _, f := range fs {
		fis = append(fis, ParserFuncInfo(f))
	}
	return fis
}
