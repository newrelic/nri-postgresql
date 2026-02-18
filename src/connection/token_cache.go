package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
)

type tokenCacheEntry struct {
	token     string
	createdAt time.Time
}

type tokenCache struct {
	cache map[string]*tokenCacheEntry
	mutex sync.RWMutex
}

var (
	globalTokenCache *tokenCache
	cacheOnce        sync.Once
)

func getTokenCache() *tokenCache {
	cacheOnce.Do(func() {
		globalTokenCache = &tokenCache{
			cache: make(map[string]*tokenCacheEntry),
		}
	})
	return globalTokenCache
}

func (tc *tokenCache) getCachedToken(ctx context.Context, endpoint, region, username string) (string, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s", endpoint, region, username)
	
	tc.mutex.RLock()
	entry, exists := tc.cache[cacheKey]
	if exists {
		tc.mutex.RUnlock()
		log.Debug("CACHE HIT: Reusing cached AWS IAM token for %s, created: %v", cacheKey, entry.createdAt.Format("15:04:05"))
		return entry.token, nil
	}
	tc.mutex.RUnlock()

	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	if entry, exists := tc.cache[cacheKey]; exists {
		log.Debug("CACHE HIT (double-check): Reusing cached AWS IAM token for %s, created: %v", cacheKey, entry.createdAt.Format("15:04:05"))
		return entry.token, nil
	}

	log.Info("CACHE MISS: Generating new AWS IAM token for %s", cacheKey)
	
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	token, err := auth.BuildAuthToken(ctx, endpoint, region, username, cfg.Credentials)
	if err != nil {
		return "", fmt.Errorf("failed to generate IAM auth token: %w", err)
	}
	
	createdAt := time.Now()
	tc.cache[cacheKey] = &tokenCacheEntry{
		token:     token,
		createdAt: createdAt,
	}
	
	log.Info("TOKEN CACHED: New AWS IAM token cached for %s, created at %v", cacheKey, createdAt.Format("15:04:05"))
	
	return token, nil
}
