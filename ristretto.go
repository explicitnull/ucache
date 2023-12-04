// ucache provides universal cache that should be able to wrap any functions and return objects of any type;
package ucache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/explicitnull/promcommon"
	"github.com/pkg/errors"
)

const (
	ristrettoCountersMultiplier = 10
	ristrettoBufferItems        = 64
	doneRecently                = "doneRecently"
)

type Ristretto struct {
	cache                *ristretto.Cache
	cacheMetrics         promcommon.CacheIncrementer
	averageItemCost      int
	maxCacheableItemCost int
	minCacheableItemCost int
	ttl                  time.Duration
}

func NewRistretto(cacheMetrics promcommon.CacheIncrementer, config Config) (*Ristretto, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(config.ItemsNum * ristrettoCountersMultiplier),
		MaxCost:     int64(config.ItemsNum * config.AverageItemCost),
		BufferItems: ristrettoBufferItems,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can't create ristretto cache")
	}

	return &Ristretto{
		cache:                cache,
		cacheMetrics:         cacheMetrics,
		averageItemCost:      config.AverageItemCost,
		minCacheableItemCost: config.MinItemCost,
		maxCacheableItemCost: config.MaxItemCost,
		ttl:                  config.TTL,
	}, nil
}

func (r *Ristretto) DoWithError(
	ctx context.Context,
	cachedFn func() error,
	key string,
	operation string,
) error {
	if _, done := r.cache.Get(key); done {
		r.cacheMetrics.IncHits(operation)

		return nil
	}

	r.cacheMetrics.IncMisses(operation)

	if err := cachedFn(); err != nil {
		return err
	}

	r.cache.SetWithTTL(key, doneRecently, int64(r.averageItemCost), r.ttl)

	return nil
}

func (r *Ristretto) GetObjectWithError(
	ctx context.Context,
	cachedFn func() (any, error),
	key string,
	operation string,
) (any, error) {
	object, found := r.cache.Get(key)
	if found {
		r.cacheMetrics.IncHits(operation)

		return object, nil
	}

	r.cacheMetrics.IncMisses(operation)

	object, err := cachedFn()
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(object)
	if err != nil {
		return nil, errors.Wrap(err, "can't marshal received object")
	}

	if len(bytes) >= r.minCacheableItemCost || len(bytes) <= r.maxCacheableItemCost {
		r.cache.SetWithTTL(key, object, int64(r.averageItemCost), r.ttl)
	}

	return object, nil
}
