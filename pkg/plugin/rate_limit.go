package plugin

import (
	"github.com/hellofresh/janus/pkg/api"
	"github.com/hellofresh/janus/pkg/errors"
	"github.com/hellofresh/janus/pkg/middleware"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/store"
	"github.com/hellofresh/stats-go"
	"github.com/mitchellh/mapstructure"
	"github.com/ulule/limiter"
)

const (
	// DefaultPrefix is the default prefix to use for the key in the store.
	DefaultPrefix = "limiter"
)

type rateLimitConfig struct {
	Limit  string `json:"limit"`
	Policy string `json:"policy"`
}

// RateLimit represents the rate limit plugin
type RateLimit struct {
	storage     store.Store
	statsClient stats.Client
}

// NewRateLimit creates a new instance of HostMatcher
func NewRateLimit(storage store.Store, statsClient stats.Client) *RateLimit {
	return &RateLimit{storage, statsClient}
}

// GetName retrieves the plugin's name
func (h *RateLimit) GetName() string {
	return "rate_limit"
}

// GetMiddlewares retrieves the plugin's middlewares
func (h *RateLimit) GetMiddlewares(rawConfig map[string]interface{}, referenceSpec *api.Spec) ([]router.Constructor, error) {
	var rateLimitConfig rateLimitConfig
	err := mapstructure.Decode(rawConfig, &rateLimitConfig)
	if err != nil {
		return nil, err
	}

	rate, err := limiter.NewRateFromFormatted(rateLimitConfig.Limit)
	if err != nil {
		return nil, err
	}

	limiterStore, err := h.getLimiterStore(rateLimitConfig.Policy, referenceSpec.Name)
	if err != nil {
		return nil, err
	}

	limiterInstance := limiter.NewLimiter(limiterStore, rate)
	return []router.Constructor{
		middleware.NewRateLimitLogger(limiterInstance, h.statsClient).Handler,
		limiter.NewHTTPMiddleware(limiterInstance).Handler,
	}, nil
}

func (h *RateLimit) getLimiterStore(policy string, prefix string) (limiter.Store, error) {
	if prefix == "" {
		prefix = DefaultPrefix
	}

	switch policy {
	case "redis":
		redisStorage, ok := h.storage.(*store.RedisStore)
		if !ok {
			return nil, errors.ErrInvalidStorage
		}

		return limiter.NewRedisStoreWithOptions(redisStorage.Pool, limiter.StoreOptions{
			Prefix:   prefix,
			MaxRetry: limiter.DefaultMaxRetry,
		})
	case "local":
		return limiter.NewMemoryStore(), nil
	default:
		return nil, errors.ErrInvalidPolicy
	}
}
