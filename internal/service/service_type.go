package service

import (
	"context"
	"sync"

	"github.com/fastly/go-fastly/v15/fastly"
)

type ServiceTypeChecker struct {
	client *fastly.Client

	mu    sync.Mutex
	cache map[string]string
}

func NewServiceTypeChecker(client *fastly.Client) *ServiceTypeChecker {
	return &ServiceTypeChecker{
		client: client,
		cache:  make(map[string]string),
	}
}

func (c *ServiceTypeChecker) GetType(
	ctx context.Context,
	serviceID string,
) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if serviceType, ok := c.cache[serviceID]; ok {
		return serviceType, nil
	}

	// GetServiceDetails returns the service type, but without a version query
	// parameter the API can also include every service version. Some services
	// have hundreds of versions, so constrain the response to a single version
	// while still retrieving the top-level service metadata needed for validation.
	version := 1
	serviceDetails, err := c.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
		Version:   &version,
	})
	if err != nil {
		return "", err
	}

	serviceType := ""
	if serviceDetails != nil {
		serviceType = fastly.ToValue(serviceDetails.Type)
	}

	c.cache[serviceID] = serviceType

	return serviceType, nil
}
