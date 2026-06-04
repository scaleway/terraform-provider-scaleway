package list

import (
	"context"
	"runtime"
	"slices"
	"sync"

	"golang.org/x/sync/errgroup"
)

var defaultFetchLimit = runtime.NumCPU()

// FetchFunc defines a function that fetches items for a given target.
type FetchFunc[T any, Target any] func(ctx context.Context, target Target) ([]T, error)

// Comparator defines a function that compares two items for sorting.
// Returns negative if a < b, zero if a == b, positive if a > b.
type Comparator[T any] func(a, b T) int

// FetchConcurrently fetches items for the provided targets concurrently,
// limiting the number of active goroutines to the default fetch limit.
// It returns all fetched items combined into a single slice.
// Items are sorted using the provided comparator if not nil.
func FetchConcurrently[T any, Target any](ctx context.Context, targets []Target, fetch FetchFunc[T, Target], compare Comparator[T]) ([]T, error) {
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

	if compare != nil {
		slices.SortFunc(allItems, compare)
	}

	return allItems, nil
}
