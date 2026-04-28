package awx

import (
	"fmt"
)

// InventoryUpdatesService implements awx inventory updates apis.
type InventoryUpdatesService struct {
	client *Client
}

const inventoryUpdatesAPIEndpoint = "/api/v2/inventory_updates/"

// GetInventoryUpdate shows the details of an inventory update.
func (p *InventoryUpdatesService) GetInventoryUpdate(id int) (*Job, error) {
	result := new(Job)
	endpoint := fmt.Sprintf("%s%d/", inventoryUpdatesAPIEndpoint, id)
	resp, err := p.client.Requester.GetJSON(endpoint, result, nil)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}
	return result, nil
}
