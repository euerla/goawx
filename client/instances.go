package awx

import (
	"fmt"
)

// InstancesService implements awx instances apis.
type InstancesService struct {
	client *Client
}

// ListInstancesResponse represents `ListInstances` endpoint response.
type ListInstancesResponse struct {
	Pagination
	Results []*Instance `json:"results"`
}

const instancesAPIEndpoint = "/api/v2/instances/"

// ListInstances shows a list of awx instances (first page only).
func (i *InstancesService) ListInstances(params map[string]string) ([]*Instance, *ListInstancesResponse, error) {
	result := new(ListInstancesResponse)
	resp, err := i.client.Requester.GetJSON(instancesAPIEndpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// ListAllInstances shows a list of awx instances and follows the pagination.
func (i *InstancesService) ListAllInstances(params map[string]string) ([]*Instance, error) {
	return fetchAllInstancePages(i.client, instancesAPIEndpoint, params)
}

func fetchAllInstancePages(client *Client, firstPath string, params map[string]string) ([]*Instance, error) {
	results := make([]*Instance, 0)

	path, query := firstPath, params
	for {
		page := new(ListInstancesResponse)
		resp, err := client.Requester.GetJSON(path, page, query)
		if err != nil {
			return nil, err
		}

		if err := CheckResponse(resp); err != nil {
			return nil, err
		}

		results = append(results, page.Results...)

		nextPath, nextQuery, ok := nextPageFrom(page.Next)
		if !ok {
			return results, nil
		}
		path, query = nextPath, nextQuery
	}
}

// GetInstanceByID shows the details of an instance.
func (i *InstancesService) GetInstanceByID(id int, params map[string]string) (*Instance, error) {
	result := new(Instance)
	endpoint := fmt.Sprintf("%s%d/", instancesAPIEndpoint, id)
	resp, err := i.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// ReadyControlInstancesParams returns the query parameters of
// /api/v2/instances/?node_type=control&enabled=true&node_state=ready&errors__regex=^$
func ReadyControlInstancesParams() map[string]string {
	return map[string]string{
		"node_type":     "control",
		"enabled":       "true",
		"node_state":    "ready",
		"errors__regex": "^$",
	}
}

// ListReadyControlInstances returns all enabled, ready control nodes which
// have no errors, following the pagination.
func (i *InstancesService) ListReadyControlInstances() ([]*Instance, error) {
	return i.ListAllInstances(ReadyControlInstancesParams())
}

// InstanceCapacity holds the capacity values of a single node.
type InstanceCapacity struct {
	ID               int    `json:"id"`
	Hostname         string `json:"hostname"`
	Capacity         int    `json:"capacity"`
	ConsumedCapacity int    `json:"consumed_capacity"`
}

// CapacityValues returns the capacity and consumed_capacity of the node.
func (i *Instance) CapacityValues() InstanceCapacity {
	return InstanceCapacity{
		ID:               i.ID,
		Hostname:         i.Hostname,
		Capacity:         i.Capacity,
		ConsumedCapacity: i.ConsumedCapacity,
	}
}

// CapacityValuesOf returns the capacity values of the given nodes.
func CapacityValuesOf(instances []*Instance) []InstanceCapacity {
	values := make([]InstanceCapacity, 0, len(instances))
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		values = append(values, instance.CapacityValues())
	}

	return values
}

// TotalCapacity sums up capacity and consumed_capacity of the given nodes.
func TotalCapacity(instances []*Instance) (capacity int, consumedCapacity int) {
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		capacity += instance.Capacity
		consumedCapacity += instance.ConsumedCapacity
	}

	return capacity, consumedCapacity
}
