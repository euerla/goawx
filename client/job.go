package awx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Enum of job statuses.
const (
	JobStatusNew        = "new"
	JobStatusPending    = "pending"
	JobStatusWaiting    = "waiting"
	JobStatusRunning    = "running"
	JobStatusSuccessful = "successful"
	JobStatusFailed     = "failed"
	JobStatusError      = "error"
	JobStatusCanceled   = "canceled"
)

// JobService implements awx job apis.
type JobService struct {
	client *Client
}

// HostSummariesResponse represents `JobHostSummaries` endpoint response.
type HostSummariesResponse struct {
	Pagination
	Results []HostSummary `json:"results"`
}

// JobEventsResponse represents `JobEvents` endpoint response.
type JobEventsResponse struct {
	Pagination
	Results []JobEvent `json:"results"`
}

// CancelJobResponse represents `CancelJob` endpoint response.
type CancelJobResponse struct {
	Detail string `json:"detail"`
}

const jobAPIEndpoint = "/api/v2/jobs/"

// GetJob shows the details of a job.
func (j *JobService) GetJob(id int, params map[string]string) (*Job, error) {
	result := new(Job)
	endpoint := fmt.Sprintf("%s%d/", jobAPIEndpoint, id)
	resp, err := j.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// CancelJob cancels a job.
func (j *JobService) CancelJob(id int, data map[string]interface{}, params map[string]string) (*CancelJobResponse, error) {
	result := new(CancelJobResponse)
	endpoint := fmt.Sprintf("%s%d/cancel/", jobAPIEndpoint, id)
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

// RelaunchJob relaunch a job.
func (j *JobService) RelaunchJob(id int, data map[string]interface{}, params map[string]string) (*JobLaunch, error) {
	result := new(JobLaunch)
	endpoint := fmt.Sprintf("%s%d/relaunch/", jobAPIEndpoint, id)
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

// GetHostSummaries get a job hosts summaries.
func (j *JobService) GetHostSummaries(id int, params map[string]string) ([]HostSummary, *HostSummariesResponse, error) {
	result := new(HostSummariesResponse)
	endpoint := fmt.Sprintf("%s%d/job_host_summaries/", jobAPIEndpoint, id)
	resp, err := j.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// GetJobEvents get a list of job events.
func (j *JobService) GetJobEvents(id int, params map[string]string) ([]JobEvent, *JobEventsResponse, error) {
	result := new(JobEventsResponse)
	endpoint := fmt.Sprintf("%s%d/job_events/", jobAPIEndpoint, id)
	resp, err := j.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// ListJobsResponse represents `ListJobs` endpoint response.
type ListJobsResponse struct {
	Pagination
	Results []*Job `json:"results"`
}

const jobLabelsAPIEndpoint = "/api/v2/jobs/%d/labels/"

// ListJobs retrieves a list of jobs matching the provided query parameters.
// It returns the jobs, the full list response, and any error encountered.
func (j *JobService) ListJobs(params map[string]string) ([]*Job, *ListJobsResponse, error) {
	result := new(ListJobsResponse)
	resp, err := j.client.Requester.GetJSON(jobAPIEndpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// ListAllJobs retrieves all jobs across every result page matching the provided query parameters.
// It returns the combined job list, or an error if fetching any page fails.
func (j *JobService) ListAllJobs(params map[string]string) ([]*Job, error) {
	return fetchAllJobPages(j.client, jobAPIEndpoint, params)
}

// fetchAllJobPages retrieves all job pages from firstPath with params, following pagination until no next page.
// It returns the accumulated jobs or the first request or response error encountered.
func fetchAllJobPages(client *Client, firstPath string, params map[string]string) ([]*Job, error) {
	results := make([]*Job, 0)

	path, query := firstPath, params
	for {
		page := new(ListJobsResponse)
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

// ListJobsByLabel returns jobs filtered by the specified label ID using the provided query parameters.
func (j *JobService) ListJobsByLabel(labelID int, params map[string]string) ([]*Job, *ListJobsResponse, error) {
	return j.ListJobs(withQuery(params, "labels", strconv.Itoa(labelID)))
}

// ListJobsByLabelName lists jobs tagged with a label matching the given name.
func (j *JobService) ListJobsByLabelName(name string, params map[string]string) ([]*Job, *ListJobsResponse, error) {
	return j.ListJobs(withQuery(params, "labels__name", name))
}

// FindJobsByLabelName returns all jobs tagged with a label of the given name, following pagination.
// If the server does not support filtering by label name, it lists all jobs and filters them locally.
// An error is returned when the underlying job listing fails.
func (j *JobService) FindJobsByLabelName(name string, params map[string]string) ([]*Job, error) {
	jobs, err := j.ListAllJobs(withQuery(params, "labels__name", name))
	if err == nil {
		return jobs, nil
	}

	if !isUnsupportedLabelFilter(err) {
		return nil, err
	}

	jobs, err = j.ListAllJobs(params)
	if err != nil {
		return nil, err
	}

	return FilterJobsByLabelName(jobs, name), nil
}

// FindJobsByLabel returns all jobs tagged with the given label ID, following pagination.
func (j *JobService) FindJobsByLabel(labelID int, params map[string]string) ([]*Job, error) {
	jobs, err := j.ListAllJobs(withQuery(params, "labels", strconv.Itoa(labelID)))
	if err == nil {
		return jobs, nil
	}

	if !isUnsupportedLabelFilter(err) {
		return nil, err
	}

	jobs, err = j.ListAllJobs(params)
	if err != nil {
		return nil, err
	}

	return FilterJobsByLabelID(jobs, labelID), nil
}

// ListLabelsForJob returns the labels associated with the job identified by id.
// The provided params are used as query parameters for the request.
func (j *JobService) ListLabelsForJob(id int, params map[string]string) ([]*Label, *ListLabelsResponse, error) {
	result := new(ListLabelsResponse)
	endpoint := fmt.Sprintf(jobLabelsAPIEndpoint, id)

	resp, err := j.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// JobHasLabel reports whether the job identified by jobID has a label with the given name.
// It returns an error if the job's labels cannot be listed.
func (j *JobService) JobHasLabel(jobID int, name string) (bool, error) {
	labels, _, err := j.ListLabelsForJob(jobID, nil)
	if err != nil {
		return false, err
	}

	for _, label := range labels {
		if label != nil && label.Name == name {
			return true, nil
		}
	}

	return false, nil
}
