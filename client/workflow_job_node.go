package awx

import (
	"fmt"
)

// WorkflowJobNodeService implements awx workflow job node apis.
type WorkflowJobNodeService struct {
	client *Client
}

// ListWorkflowJobNodesResponse represents `ListWorkflowJobNodes` endpoint response.
type ListWorkflowJobNodesResponse struct {
	Pagination
	Results []*WorkflowJobNode `json:"results"`
}

const workflowJobNodeAPIEndpoint = "/api/v2/workflow_job_nodes/"

// GetWorkflowJobNodeByID shows the details of a workflow job node.
func (jt *WorkflowJobNodeService) GetWorkflowJobNodeByID(id int, params map[string]string) (*WorkflowJobNode, error) {
	result := new(WorkflowJobNode)
	endpoint := fmt.Sprintf("%s%d/", workflowJobNodeAPIEndpoint, id)
	resp, err := jt.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// ListWorkflowJobNodes shows a list of workflow job nodes.
func (jt *WorkflowJobNodeService) ListWorkflowJobNodes(params map[string]string) ([]*WorkflowJobNode, *ListWorkflowJobNodesResponse, error) {
	result := new(ListWorkflowJobNodesResponse)

	resp, err := jt.client.Requester.GetJSON(workflowJobNodeAPIEndpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}
