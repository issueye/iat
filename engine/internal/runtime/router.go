package runtime

import (
	"iat/common/model"
	"iat/engine/internal/service"
)

type Router struct {
	registry *service.RegistryService
}

func NewRouter(registry *service.RegistryService) *Router {
	return &Router{registry: registry}
}

func (r *Router) Route(capabilityName string) (*model.Agent, error) {
	agents, err := r.registry.Discover(capabilityName)
	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return nil, nil
	}

	// Simple round-robin or first available for now
	// In the future, this could be based on load, performance, or cost
	return &agents[0], nil
}
