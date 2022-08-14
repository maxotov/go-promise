package promise

import (
	"context"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Promise[T any] chan T

func (p Promise[T]) NewPromise() Promise[T] {
	return make(chan T, 1)
}

func (p Promise[T]) Resolve(x T) {
	p <- x
	close(p)
}

func (p Promise[T]) WaitForResolve(ctx context.Context) (T, error) {
	var t T
	select {
	case result := <-p:
		return result, nil
	case <-ctx.Done():
		return t, ctx.Err()
	}
}

type Promises[T any] struct {
	sync.Mutex
	promises []Promise[T]
	resolved bool
	value    T
}

func (pp *Promises[T]) Add(promise Promise[T]) {
	pp.Lock()
	defer pp.Unlock()

	pp.promises = append(pp.promises, promise)
}

func (pp *Promises[T]) Resolve(x T) {
	pp.Lock()
	defer pp.Unlock()

	for _, p := range pp.promises {
		p.Resolve(x)
	}

	pp.resolved = true
	pp.value = x

}

func (pp *Promises[T]) Value() (T, bool) {
	pp.Lock()
	defer pp.Unlock()

	if pp.resolved {
		return pp.value, true
	}

	var t T
	return t, false
}

const (
	cacheDefaultExpiration = 1 * time.Hour
	cacheCleanupInterval   = 10 * time.Minute
)

type MultiPromises[T any] struct {
	v *cache.Cache
}

func (s MultiPromises[T]) NewMultiPromises() *MultiPromises[T] {
	return &MultiPromises[T]{
		v: cache.New(cacheDefaultExpiration, cacheCleanupInterval),
	}
}

func (s *MultiPromises[T]) WaitForValue(ctx context.Context, key string) (T, error) {
	p := s.AddAndGet(key)

	return p.WaitForResolve(ctx)
}

func (s *MultiPromises[T]) AddAndGet(key string) Promise[T] {
	promise := make(Promise[T]).NewPromise()

	// We use Add to avoids race conditions when set initial slice of promises by key
	err := s.v.Add(key, &Promises[T]{promises: []Promise[T]{promise}}, cache.DefaultExpiration)
	if err == nil { // If item doesn't exists
		return promise
	}

	pp := s.Get(key)
	if x, isResolved := pp.Value(); isResolved {
		promise.Resolve(x)
		return promise
	}

	pp.Add(promise)

	return promise
}

func (s *MultiPromises[T]) Get(key string) *Promises[T] {
	var pp *Promises[T]
	v, ok := s.v.Get(key)
	if ok {
		pp = v.(*Promises[T])
	}

	return pp
}

func (s *MultiPromises[T]) Resolve(key string, value T) {
	// We use Add to avoids race conditions when set resolved promise by key
	err := s.v.Add(key, &Promises[T]{resolved: true, value: value}, cache.DefaultExpiration)
	if err == nil { // If item doesn't exists
		return
	}

	pp := s.Get(key)
	pp.Resolve(value)
}
