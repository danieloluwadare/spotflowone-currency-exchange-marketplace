package rates

import (
	"context"
	"fmt"
	"sync"

	"spotflowone/internal/domain"
)

type CachedProvider struct {
	upstream Provider
	mu       sync.RWMutex
	cache    map[string]domain.MarketRate
}

// NewCachedProvider creates a new cached rate provider that wraps an upstream provider.
// The cache stores the last successfully fetched rate for each currency pair.
// Returns a pointer to the initialized cached provider.
func NewCachedProvider(upstream Provider) *CachedProvider {
	return &CachedProvider{
		upstream: upstream,
		cache:    make(map[string]domain.MarketRate),
	}
}

// GetRate attempts to fetch a rate from the upstream provider, falling back to cache on failure.
// If the upstream request succeeds, the result is cached and returned with Source="LIVE".
// If the upstream request fails but a cached rate exists, it returns the cached rate with Source="CACHED".
// If no upstream provider is configured, it attempts to return a cached rate only.
// Returns an error if both upstream and cache lookups fail.
func (c *CachedProvider) GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.MarketRate, error) {
	key := fmt.Sprintf("%s:%s", domain.NormalizeCurrency(baseCurrency), domain.NormalizeCurrency(quoteCurrency))
	if c.upstream == nil {
		return c.fromCache(key)
	}

	rate, err := c.upstream.GetRate(ctx, baseCurrency, quoteCurrency)
	if err == nil && rate.Available {
		c.mu.Lock()
		c.cache[key] = rate
		c.mu.Unlock()
		return rate, nil
	}

	cached, cacheErr := c.fromCache(key)
	if cacheErr == nil {
		cached.Source = "CACHED"
		return cached, nil
	}

	if err != nil {
		return domain.MarketRate{}, err
	}
	return domain.MarketRate{}, cacheErr
}

// fromCache retrieves a rate from the internal cache by currency pair key.
// Returns the cached rate with Available=true if found, or an error if not in cache.
func (c *CachedProvider) fromCache(key string) (domain.MarketRate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rate, ok := c.cache[key]
	if !ok {
		return domain.MarketRate{}, fmt.Errorf("no cached rate for %s", key)
	}
	rate.Available = true
	return rate, nil
}
