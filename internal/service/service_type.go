package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/fastly/go-fastly/v15/fastly"
	"golang.org/x/sync/singleflight"
)

type ServiceTypeCheckKey struct {
	ServiceID string
}

type ServiceTypeResult struct {
	Type string
}

type ServiceTypeChecker struct {
	client *fastly.Client

	mu    sync.Mutex
	cache map[ServiceTypeCheckKey]ServiceTypeResult
	group singleflight.Group
}

func NewServiceTypeChecker(client *fastly.Client) *ServiceTypeChecker {
	return &ServiceTypeChecker{
		client: client,
		cache:  make(map[ServiceTypeCheckKey]ServiceTypeResult),
	}
}

func (c *ServiceTypeChecker) GetType(
	ctx context.Context,
	serviceID string,
) (string, error) {
	result, err := c.Get(ctx, serviceID)
	if err != nil {
		return "", err
	}

	return result.Type, nil
}

func (c *ServiceTypeChecker) Get(
	ctx context.Context,
	serviceID string,
) (ServiceTypeResult, error) {
	key := ServiceTypeCheckKey{
		ServiceID: serviceID,
	}

	c.mu.Lock()
	if cached, ok := c.cache[key]; ok {
		c.mu.Unlock()
		return cached, nil
	}
	c.mu.Unlock()

	value, err, _ := c.group.Do(serviceID, func() (interface{}, error) {
		c.mu.Lock()
		if cached, ok := c.cache[key]; ok {
			c.mu.Unlock()
			return cached, nil
		}
		c.mu.Unlock()

		serviceDetails, err := c.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
			ServiceID: serviceID,
		})
		if err != nil {
			return ServiceTypeResult{}, err
		}

		result := ServiceTypeResult{}
		if serviceDetails != nil {
			result.Type = fastly.ToValue(serviceDetails.Type)
		}

		c.mu.Lock()
		c.cache[key] = result
		c.mu.Unlock()

		return result, nil
	})
	if err != nil {
		return ServiceTypeResult{}, err
	}

	result, ok := value.(ServiceTypeResult)
	if !ok {
		return ServiceTypeResult{}, fmt.Errorf("unexpected service type result type %T", value)
	}

	return result, nil
}
