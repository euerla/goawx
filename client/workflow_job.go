package awx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Enum of job statuses.
const (
	WorkflowJobStatusNew        = "new"
	WorkflowJobStatusPending    = "pending"
	WorkflowJobStatusWaiting    = "waiting"
	WorkflowJobStatusRunning    = "running"
	WorkflowJobStatusSuccessful = "successful"
	WorkflowJobStatusFailed     = "failed"
	WorkflowJobStatusError      = "error"
	WorkflowJobStatusCanceled   = "canceled"
)

// WorkflowJobService implements awx workflow job apis.
type WorkflowJobService struct {
	client *Client
}

// CancelWorkflowJobResponse represents `CancelWorkflowJob` endpoint response.
type CancelWorkflowJobResponse struct {
	Detail string `json:"detail"`
}

const WorkflowJobAPIEndpoint = "/api/v2/workflow_jobs/"

// GetWorkflowJob shows the details of a workflow job.
func (j *WorkflowJobService) GetWorkflowJob(id int, params map[string]string) (*WorkflowJob, error) {
	result := new(WorkflowJob)
	endpoint := fmt.Sprintf("%s%d/", WorkflowJobAPIEndpoint, id)
	resp, err := j.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// CancelWorkflowJob cancels a workflow job.
func (j *WorkflowJobService) CancelWorkflowJob(id int, data map[string]interface{}, params map[string]string) (*CancelWorkflowJobResponse, error) {
	result := new(CancelWorkflowJobResponse)
	endpoint := fmt.Sprintf("%s%d/cancel/", WorkflowJobAPIEndpoint, id)
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := j.client.Requester.PostJSON(endpoint, bytes.NewReader(payload), result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// RelaunchWorkflowJob relaunch a workflow job.
func (j *WorkflowJobService) RelaunchWorkflowJob(id int, data map[string]interface{}, params map[string]string) (*WorkflowJobLaunch, error) {
	result := new(WorkflowJobLaunch)
	endpoint := fmt.Sprintf("%s%d/relaunch/", WorkflowJobAPIEndpoint, id)
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := j.client.Requester.PostJSON(endpoint, bytes.NewReader(payload), result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// ListWorkflowJobsResponse represents `ListWorkflowJobs` endpoint response.
type ListWorkflowJobsResponse struct {
	Pagination
	Results []*WorkflowJob `json:"results"`
}

const workflowJobLabelsAPIEndpoint = "/api/v2/workflow_jobs/%d/labels/"

// ListWorkflowJobs shows a list of workflow jobs.
func (j *WorkflowJobService) ListWorkflowJobs(params map[string]string) ([]*WorkflowJob, *ListWorkflowJobsResponse, error) {
	result := new(ListWorkflowJobsResponse)
	resp, err := j.client.Requester.GetJSON(WorkflowJobAPIEndpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// ListAllWorkflowJobs shows a list of workflow jobs and follows the pagination.
func (j *WorkflowJobService) ListAllWorkflowJobs(params map[string]string) ([]*WorkflowJob, error) {
	return fetchAllWorkflowJobPages(j.client, WorkflowJobAPIEndpoint, params)
}

// fetchAllWorkflowJobPages retrieves all workflow jobs by fetching firstPath and following next-page links.
// It uses params for the initial query and returns the accumulated jobs or the first request or response error.
func fetchAllWorkflowJobPages(client *Client, firstPath string, params map[string]string) ([]*WorkflowJob, error) {
	results := make([]*WorkflowJob, 0)

	path, query := firstPath, params
	for {
		page := new(ListWorkflowJobsResponse)
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

// ListWorkflowJobsByLabel lists workflow jobs tagged with the given label id.
func (j *WorkflowJobService) ListWorkflowJobsByLabel(labelID int, params map[string]string) ([]*WorkflowJob, *ListWorkflowJobsResponse, error) {
	return j.ListWorkflowJobs(withQuery(params, "labels", strconv.Itoa(labelID)))
}

// ListWorkflowJobsByLabelName lists workflow jobs tagged with a label of the given name.
func (j *WorkflowJobService) ListWorkflowJobsByLabelName(name string, params map[string]string) ([]*WorkflowJob, *ListWorkflowJobsResponse, error) {
	return j.ListWorkflowJobs(withQuery(params, "labels__name", name))
}

// FindWorkflowJobsByLabelName is the pagination-aware variant of ListWorkflowJobsByLabelName
func (j *WorkflowJobService) FindWorkflowJobsByLabelName(name string, params map[string]string) ([]*WorkflowJob, error) {
	jobs, err := j.ListAllWorkflowJobs(withQuery(params, "labels__name", name))
	if err == nil {
		return jobs, nil
	}

	if !isUnsupportedLabelFilter(err) {
		return nil, err
	}

	jobs, err = j.ListAllWorkflowJobs(params)
	if err != nil {
		return nil, err
	}

	return FilterWorkflowJobsByLabelName(jobs, name), nil
}

// ListLabelsForWorkflowJob lists the labels attached to a workflow job.
func (j *WorkflowJobService) ListLabelsForWorkflowJob(id int, params map[string]string) ([]*Label, *ListLabelsResponse, error) {
	result := new(ListLabelsResponse)
	endpoint := fmt.Sprintf(workflowJobLabelsAPIEndpoint, id)

	resp, err := j.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}
