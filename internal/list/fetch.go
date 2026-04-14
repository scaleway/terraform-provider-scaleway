package list

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

var defaultFetchLimit = runtime.NumCPU()

// FetchFunc defines a function that fetches items for a given target.
type FetchFunc[T any, Target any] func(ctx context.Context, target Target) ([]T, error)

// FetchConcurrently fetches items for the provided targets concurrently,
// limiting the number of active goroutines to the default fetch limit.
// It returns all fetched items combined into a single slice.
func FetchConcurrently[T any, Target any](ctx context.Context, targets []Target, fetch FetchFunc[T, Target]) ([]T, error) {
	var mu sync.Mutex

	var allItems []T

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(defaultFetchLimit)

	for _, target := range targets {
		g.Go(func() error {
			items, err := fetch(ctx, target)
			if err != nil {
				return err
			}

			mu.Lock()

			allItems = append(allItems, items...)

			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return allItems, nil
}
