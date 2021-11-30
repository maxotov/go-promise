package promise

import (
	"context"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Int64Promise chan int64

func NewInt64Promise() Int64Promise {
	return Int64Promise(make(chan int64, 1))
}

func (p Int64Promise) Resolve(x int64) {
	p <- x
	close(p)
}

func (p Int64Promise) WaitForResolve(ctx context.Context) (int64, error) {
	select {
	case result := <-p:
		return result, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

type Int64Promises struct {
	sync.Mutex
	promises []Int64Promise
	resolved bool
	value    int64
}

func (pp *Int64Promises) Add(promise Int64Promise) {
	pp.Lock()
	defer pp.Unlock()

	pp.promises = append(pp.promises, promise)
}

func (pp *Int64Promises) Resolve(x int64) {
	pp.Lock()
	defer pp.Unlock()

	for _, p := range pp.promises {
		p.Resolve(x)
	}

	pp.resolved = true
	pp.value = x

}

func (pp *Int64Promises) Value() (int64, bool) {
	pp.Lock()
	defer pp.Unlock()

	if pp.resolved {
		return pp.value, true
	}

	return 0, false
}

const (
	cacheDefaultExpiration = 1 * time.Hour
	cacheCleanupInterval   = 10 * time.Minute
)

type Int64MultiPromises struct {
	v *cache.Cache
}

func NewInt64MultiPromises() *Int64MultiPromises {
	return &Int64MultiPromises{
		v: cache.New(cacheDefaultExpiration, cacheCleanupInterval),
	}
}

func (s *Int64MultiPromises) WaitForValue(ctx context.Context, key string) (int64, error) {
	p := s.AddAndGet(key)

	return p.WaitForResolve(ctx)
}

func (s *Int64MultiPromises) AddAndGet(key string) Int64Promise {
	promise := NewInt64Promise()

	// We use Add to avoids race conditions when set initial slice of promises by key
	err := s.v.Add(key, &Int64Promises{promises: []Int64Promise{promise}}, cache.DefaultExpiration)
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

func (s *Int64MultiPromises) Get(key string) *Int64Promises {
	var pp *Int64Promises
	v, ok := s.v.Get(key)
	if ok {
		pp = v.(*Int64Promises)
	}

	return pp
}

func (s *Int64MultiPromises) Resolve(key string, value int64) {
	// We use Add to avoids race conditions when set resolved promise by key
	err := s.v.Add(key, &Int64Promises{resolved: true, value: value}, cache.DefaultExpiration)
	if err == nil { // If item doesn't exists
		return
	}

	pp := s.Get(key)
	pp.Resolve(value)
}
