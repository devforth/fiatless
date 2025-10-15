package cache

import (
	"context"
	"sync"
)

type LazyCache[T any] struct {
	mu      sync.Mutex
	value   T
	err     error
	fetched bool
	fetch   func(context.Context) (T, error)
}

func NewLazyCache[T any](fetch func(context.Context) (T, error)) *LazyCache[T] {
	return &LazyCache[T]{
		fetch: fetch,
	}
}

func (l *LazyCache[T]) Get(ctx context.Context) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.fetched {
		l.value, l.err = l.fetch(ctx)
		l.fetched = true
	}

	return l.value, l.err
}

func (l *LazyCache[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fetched = false
	var zero T
	l.value = zero
	l.err = nil
}
