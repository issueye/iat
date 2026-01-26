package orchestrator

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

	return &agents[0], nil
}
