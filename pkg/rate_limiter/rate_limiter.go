package rate_limiter

import (
	"time"

	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/memorystore"
)

func NewRateLimiter(tokens uint64, interval time.Duration) (limiter.Store, error) {
	return memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: interval,
	})
}
