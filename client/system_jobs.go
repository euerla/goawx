package awx

import (
	"fmt"
)

// SystemJobsService implements awx system jobs apis.
type SystemJobsService struct {
	client *Client
}

const systemJobsAPIEndpoint = "/api/v2/system_jobs/"

// GetSystemJob shows the details of a system job.
func (p *SystemJobsService) GetSystemJob(id int) (*Job, error) {
	result := new(Job)
	endpoint := fmt.Sprintf("%s%d/", systemJobsAPIEndpoint, id)
	resp, err := p.client.Requester.GetJSON(endpoint, result, nil)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}
	return result, nil
}
