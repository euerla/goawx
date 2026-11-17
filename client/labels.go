package awx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// LabelService implements awx labels apis.
type LabelService struct {
	client *Client
}

// ListLabelsResponse represents `ListLabels` endpoint response.
type ListLabelsResponse struct {
	Pagination
	Results []*Label `json:"results"`
}

const labelsAPIEndpoint = "/api/v2/labels/"

// ErrLabelsNotPromptedOnLaunch is returned if labels are passed to a launch
// endpoint of a template which has `ask_labels_on_launch = false`.
var ErrLabelsNotPromptedOnLaunch = errors.New("awx: labels cannot be set on launch, ask_labels_on_launch is false")

// ListLabels shows a list of awx labels.
func (l *LabelService) ListLabels(params map[string]string) ([]*Label, *ListLabelsResponse, error) {
	result := new(ListLabelsResponse)
	resp, err := l.client.Requester.GetJSON(labelsAPIEndpoint, result, params)
	if err != nil {
		return nil, result, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, result, err
	}

	return result.Results, result, nil
}

// ListAllLabels returns all labels of the awx instance, following the pagination.
func (l *LabelService) ListAllLabels(params map[string]string) ([]*Label, error) {
	results := make([]*Label, 0)

	nextURL := labelsAPIEndpoint
	nextParams := params

	for {
		nextURLParsed, err := url.Parse(nextURL)
		if err != nil {
			return nil, err
		}

		nextURLQueryParams := make(map[string]string)
		for paramName, paramValues := range nextURLParsed.Query() {
			if len(paramValues) > 0 {
				nextURLQueryParams[paramName] = paramValues[0]
			}
		}

		for paramName, paramValue := range nextParams {
			nextURLQueryParams[paramName] = paramValue
		}

		result := new(ListLabelsResponse)
		resp, err := l.client.Requester.GetJSON(nextURLParsed.Path, result, nextURLQueryParams)
		if err != nil {
			return nil, err
		}

		if err := CheckResponse(resp); err != nil {
			return nil, err
		}

		results = append(results, result.Results...)

		next, ok := result.Next.(string)
		if !ok || next == "" {
			break
		}

		nextURL = next
		nextParams = nil
	}

	return results, nil
}

// GetLabelByID shows the details of an awx label.
func (l *LabelService) GetLabelByID(id int, params map[string]string) (*Label, error) {
	result := new(Label)
	endpoint := fmt.Sprintf("%s%d/", labelsAPIEndpoint, id)
	resp, err := l.client.Requester.GetJSON(endpoint, result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateLabel creates an awx label. `name` is mandatory, `organization` is optional.
// Bewusst ohne das globale `mandatoryFields` - das ist nicht thread-safe.
func (l *LabelService) CreateLabel(data map[string]interface{}, params map[string]string) (*Label, error) {
	validate, status := ValidateParams(data, []string{"name"})
	if !status {
		return nil, fmt.Errorf("mandatory input arguments are absent: %s", validate)
	}

	result := new(Label)
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := l.client.Requester.PostJSON(labelsAPIEndpoint, bytes.NewReader(payload), result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateLabel updates an awx label.
func (l *LabelService) UpdateLabel(id int, data map[string]interface{}, params map[string]string) (*Label, error) {
	result := new(Label)
	endpoint := fmt.Sprintf("%s%d/", labelsAPIEndpoint, id)
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := l.client.Requester.PatchJSON(endpoint, bytes.NewReader(payload), result, params)
	if err != nil {
		return nil, err
	}

	if err := CheckResponse(resp); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteLabel deletes an awx label.
func (l *LabelService) DeleteLabel(id int) error {
	endpoint := fmt.Sprintf("%s%d/", labelsAPIEndpoint, id)

	resp, err := l.client.Requester.Delete(endpoint, nil, nil)
	if err != nil {
		return err
	}

	return CheckResponse(resp)
}

// FindLabelByName finds a label by name, optionally scoped to an organization if organizationID is positive.
func (l *LabelService) FindLabelByName(name string, organizationID int) (*Label, error) {
	params := map[string]string{"name": name}
	if organizationID > 0 {
		params["organization"] = strconv.Itoa(organizationID)
	}

	labels, _, err := l.ListLabels(params)
	if err != nil {
		return nil, err
	}

	for _, label := range labels {
		if label == nil || label.Name != name {
			continue
		}
		if organizationID > 0 && label.Organization != organizationID {
			continue
		}
		return label, nil
	}

	return nil, nil
}

// GetOrCreateLabel returns the label with the given name and organization, creating it if it does not exist.
// It returns an error for an empty name or on lookup/creation failure; concurrent creation falls back to lookup.
func (l *LabelService) GetOrCreateLabel(name string, organizationID int) (*Label, error) {
	if name == "" {
		return nil, fmt.Errorf("label name must not be empty")
	}

	label, err := l.FindLabelByName(name, organizationID)
	if err != nil {
		return nil, err
	}
	if label != nil {
		return label, nil
	}

	data := map[string]interface{}{"name": name}
	if organizationID > 0 {
		data["organization"] = organizationID
	}

	label, err = l.CreateLabel(data, nil)
	if err == nil {
		return label, nil
	}

	// If someone else has created the label in the meantime (or a race condition has occurred):
	// perform the search again and return the existing value.
	var awxErr *AWXError
	if errors.As(err, &awxErr) && awxErr.StatusCode == 400 {
		if existing, findErr := l.FindLabelByName(name, organizationID); findErr == nil && existing != nil {
			return existing, nil
		}
	}

	return nil, err
}

// LabelIDs returns the ids of the labels contained in a summary_fields.labels.
func (l *Labels) LabelIDs() []int {
	if l == nil {
		return nil
	}

	ids := make([]int, 0, len(l.Results))
	for _, label := range l.Results {
		if label == nil {
			continue
		}
		ids = append(ids, label.ID)
	}

	return ids
}

// LabelNames returns the names of the labels contained in a summary_fields.labels.
func (l *Labels) LabelNames() []string {
	if l == nil {
		return nil
	}

	names := make([]string, 0, len(l.Results))
	for _, label := range l.Results {
		if label == nil {
			continue
		}
		names = append(names, label.Name)
	}

	return names
}

// HasLabel reports whether the label with the given name is present.
func (l *Labels) HasLabel(name string) bool {
	if l == nil {
		return false
	}

	for _, label := range l.Results {
		if label != nil && label.Name == name {
			return true
		}
	}

	return false
}

// HasLabelID reports whether the label with the given id is attached to the object.
func (l *Labels) HasLabelID(id int) bool {
	if l == nil {
		return false
	}

	for _, label := range l.Results {
		if label != nil && label.ID == id {
			return true
		}
	}

	return false
}

// GetLabels returns summary_fields.labels of the job, never nil.
func (j *Job) GetLabels() *Labels {
	if j == nil || j.SummaryFields == nil || j.SummaryFields.Labels == nil {
		return &Labels{}
	}

	return j.SummaryFields.Labels
}

// HasLabel reports whether the job carries a label with the given name.
func (j *Job) HasLabel(name string) bool { return j.GetLabels().HasLabel(name) }

// HasLabelID reports whether the job carries the label with the given id.
func (j *Job) HasLabelID(id int) bool { return j.GetLabels().HasLabelID(id) }

// GetLabels returns summary_fields.labels of the workflow job, never nil.
func (j *WorkflowJob) GetLabels() *Labels {
	if j == nil || j.SummaryFields == nil || j.SummaryFields.Labels == nil {
		return &Labels{}
	}

	return j.SummaryFields.Labels
}

// HasLabel reports whether the workflow job carries a label with the given name.
func (j *WorkflowJob) HasLabel(name string) bool { return j.GetLabels().HasLabel(name) }

// HasLabelID reports whether the workflow job carries the label with the given id.
func (j *WorkflowJob) HasLabelID(id int) bool { return j.GetLabels().HasLabelID(id) }

// FilterJobsByLabelName keeps only jobs which carry the label with the given name.
func FilterJobsByLabelName(jobs []*Job, name string) []*Job {
	filtered := make([]*Job, 0, len(jobs))
	for _, job := range jobs {
		if job.HasLabel(name) {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

// FilterJobsByLabelID keeps only jobs which carry the label with the given id.
func FilterJobsByLabelID(jobs []*Job, labelID int) []*Job {
	filtered := make([]*Job, 0, len(jobs))
	for _, job := range jobs {
		if job.HasLabelID(labelID) {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

// FilterWorkflowJobsByLabelName keeps only workflow jobs carrying the given label.
func FilterWorkflowJobsByLabelName(jobs []*WorkflowJob, name string) []*WorkflowJob {
	filtered := make([]*WorkflowJob, 0, len(jobs))
	for _, job := range jobs {
		if job.HasLabel(name) {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

// FilterWorkflowJobsByLabelID keeps only workflow jobs carrying the given label id.
func FilterWorkflowJobsByLabelID(jobs []*WorkflowJob, labelID int) []*WorkflowJob {
	filtered := make([]*WorkflowJob, 0, len(jobs))
	for _, job := range jobs {
		if job.HasLabelID(labelID) {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

// nextPageFrom splits an AWX pagination `next` link into path and query parameters.
func nextPageFrom(next interface{}) (path string, params map[string]string, ok bool) {
	link, isString := next.(string)
	if !isString || link == "" {
		return "", nil, false
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return "", nil, false
	}

	params = make(map[string]string, len(parsed.Query()))
	for key, values := range parsed.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return parsed.Path, params, true
}

// isUnsupportedLabelFilter reports whether err is a 400 AWX error indicating an unsupported label filter.
func isUnsupportedLabelFilter(err error) bool {
	var awxErr *AWXError
	if !errors.As(err, &awxErr) || awxErr.StatusCode != 400 {
		return false
	}

	body := strings.ToLower(awxErr.Body + " " + awxErr.Message)
	for _, needle := range []string{
		"labels", "invalid filter", "cannot resolve keyword", "unknown field", "not a valid field",
	} {
		if strings.Contains(body, needle) {
			return true
		}
	}

	return false
}

// LabelsPromptable reports whether labels may be set on launch of this job template.
func (jt *JobTemplate) LabelsPromptable() bool {
	return jt != nil && jt.AskLabelsOnLaunch
}

// LabelsPromptable reports whether labels may be set on launch of this workflow job template.
func (jt *WorkflowJobTemplate) LabelsPromptable() bool {
	return jt != nil && jt.AskLabelsOnLaunch
}

// LabelsPromptable reports whether labels were prompted on launch of this job.
func (j *Job) LabelsPromptable() bool {
	return j != nil && j.AskLabelsOnLaunch
}

// EnsureLabelsAllowed returns ErrLabelsNotPromptedOnLaunch if labelIDs are given
// but the template does not allow labels on launch.
func EnsureLabelsAllowed(askLabelsOnLaunch bool, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}
	if !askLabelsOnLaunch {
		return ErrLabelsNotPromptedOnLaunch
	}

	return nil
}

// LaunchDataWithLabels returns a copy of data with the given label ids set in the
// `labels` field, ready to be passed to the Launch methods.
func LaunchDataWithLabels(data map[string]interface{}, labelIDs []int) map[string]interface{} {
	if len(labelIDs) == 0 {
		return data
	}

	launchData := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		launchData[k] = v
	}

	ids := make([]int, len(labelIDs))
	copy(ids, labelIDs)
	launchData["labels"] = ids

	return launchData
}

// withQuery returns a copy of params with the given key/value added.
func withQuery(params map[string]string, key, value string) map[string]string {
	merged := make(map[string]string, len(params)+1)
	for k, v := range params {
		merged[k] = v
	}
	merged[key] = value

	return merged
}
