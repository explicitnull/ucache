package ucache

import "time"

type Config struct {
	ItemsNum             int
	AverageItemCost      int
	MinCacheableItemCost int
	MaxCacheableItemCost int
	TTL                  time.Duration
}
