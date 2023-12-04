package ucache

import "context"

type Cache interface {
	DoWithError(
		ctx context.Context,
		cachedFn func() error,
		key string,
		operation string,
	) error

	GetObjectWithError(
		ctx context.Context,
		cachedFn func() (any, error),
		key string,
		operation string,
	) (any, error)
}
