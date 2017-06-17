package plugin

import (
	"github.com/hellofresh/janus/pkg/api"
	"github.com/hellofresh/janus/pkg/middleware"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/mitchellh/mapstructure"
)

// RequestTransformer will apply a template to a request body to transform it's contents ready for an upstream API
type RequestTransformer struct{}

// NewRequestTransformer creates a new instance of RequestTransformer
func NewRequestTransformer() *RequestTransformer {
	return &RequestTransformer{}
}

// GetName retrieves the plugin's name
func (h *RequestTransformer) GetName() string {
	return "request_transformer"
}

// GetMiddlewares retrieves the plugin's middlewares
func (h *RequestTransformer) GetMiddlewares(rawConfig map[string]interface{}, referenceSpec *api.Spec) ([]router.Constructor, error) {
	var config middleware.RequestTransformerConfig
	err := mapstructure.Decode(rawConfig, &config)
	if err != nil {
		return nil, err
	}

	return []router.Constructor{
		middleware.NewRequestTransformer(config).Handler,
	}, nil
}
