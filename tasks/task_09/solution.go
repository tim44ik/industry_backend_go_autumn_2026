package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error)) (
	[]Result[R],
	error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(in) == 0 {
		return []Result[R]{}, nil
	}

	if ctx == nil || fn == nil {
		return nil, nil
	}

	workersCount := min(workers, len(in))

	sem := make(chan struct{}, workersCount)
	res := make([]Result[R], len(in))

	var wg sync.WaitGroup
	wg.Add(len(in))

	for i, val := range in {
		go func(index int, v T) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
			}

			defer func() {
				<-sem
			}()

			r, err := fn(ctx, v)
			if err != nil {
				var zero R
				r = zero
			}

			res[index] = Result[R]{Value: r, Err: err}
		}(i, val)
	}

	wg.Wait()
	close(sem)

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return res, nil
}
