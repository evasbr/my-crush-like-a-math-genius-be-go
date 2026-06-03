package configuration

import (
	"context"
	"encoding/json"
	"evasbr/mclamg/exception"
	"log"
	"strconv"

	"github.com/go-redis/redis/v9"
)

func NewRedis(config Config) *redis.Client {
	redisURL := config.Get("REDIS_URL")
	if redisURL == "" {
		redisURL = config.Get("KV_URL") // Vercel KV integration
	}

	var options *redis.Options
	if redisURL != "" {
		var err error
		options, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Failed to parse Redis connection URL: %v", err)
		}
	} else {
		host := config.Get("REDIS_HOST")
		port := config.Get("REDIS_PORT")
		if host == "" || port == "" {
			log.Fatalf("Redis configuration error: both REDIS_HOST and REDIS_PORT must be specified if REDIS_URL/KV_URL is not set.")
		}

		maxPoolSizeStr := config.Get("REDIS_POOL_MAX_SIZE")
		if maxPoolSizeStr == "" {
			log.Fatalf("Redis configuration error: REDIS_POOL_MAX_SIZE is missing.")
		}
		maxPoolSize, err := strconv.Atoi(maxPoolSizeStr)
		if err != nil {
			log.Fatalf("Redis configuration error: REDIS_POOL_MAX_SIZE ('%s') must be a valid integer: %v", maxPoolSizeStr, err)
		}

		minIdlePoolSizeStr := config.Get("REDIS_POOL_MIN_IDLE_SIZE")
		if minIdlePoolSizeStr == "" {
			log.Fatalf("Redis configuration error: REDIS_POOL_MIN_IDLE_SIZE is missing.")
		}
		minIdlePoolSize, err := strconv.Atoi(minIdlePoolSizeStr)
		if err != nil {
			log.Fatalf("Redis configuration error: REDIS_POOL_MIN_IDLE_SIZE ('%s') must be a valid integer: %v", minIdlePoolSizeStr, err)
		}

		options = &redis.Options{
			Addr:         host + ":" + port,
			PoolSize:     maxPoolSize,
			MinIdleConns: minIdlePoolSize,
		}
	}

	return redis.NewClient(options)
}

func SetCache[T any](cacheManager *redis.Client, ctx context.Context, prefix string, key string, executeData func(context.Context, string) (T, error)) *T {
	var data []byte
	var object T
	if err := cacheManager.Get(ctx, prefix+"_"+key).Scan(&data); err == nil {
		err := json.Unmarshal(data, &object)
		exception.PanicLogging(err)

		return &object
	}
	value, err := executeData(ctx, key)
	if err != nil {
		panic(exception.NotFoundError{
			Message: err.Error(),
		})
	}
	cacheValue, err := json.Marshal(value)
	exception.PanicLogging(err)

	if err := cacheManager.Set(ctx, prefix+"_"+key, cacheValue, -1).Err(); err != nil {
		exception.PanicLogging(err)
	}
	return &value
}
