package awx

import (
	"fmt"
)

// WorkflowApprovalsService implements awx workflow approvals apis.
type WorkflowApprovalsService struct {
	client *Client
}

const workflowApprovalsAPIEndpoint = "/api/v2/workflow_approvals/"

// GetWorkflowApproval shows the details of a workflow approval.
func (p *WorkflowApprovalsService) GetWorkflowApproval(id int) (*Job, error) {
	result := new(Job)
	endpoint := fmt.Sprintf("%s%d/", workflowApprovalsAPIEndpoint, id)
	resp, err := p.client.Requester.GetJSON(endpoint, result, nil)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}
	return result, nil
}
