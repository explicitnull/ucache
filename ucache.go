package ucache

import "time"

type Config struct {
	ItemsNum        int
	AverageItemCost int
	MinItemCost     int
	MaxItemCost     int
	TTL             time.Duration
}
