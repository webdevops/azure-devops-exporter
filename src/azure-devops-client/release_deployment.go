package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type ReleaseDeploymentList struct {
	Count int       `json:"count"`
	List  []ReleaseDeployment `json:"value"`
}

type ReleaseDeployment struct {
	Id   int64 `json:"id"`
	Name string

	Release struct {
		Id    int64
		Name  string
		Links Links `json:"_links"`
	} `json:"release"`

	ReleaseDefinition struct {
		Id    int64
		Name  string
		Path  string
	} `json:"releaseDefinition"`

	Artifacts []ReleaseArtifact

	ReleaseEnvironment ReleaseDeploymentEnvironment

	PreDeployApprovals []ReleaseEnvironmentApproval
	PostDeployApprovals []ReleaseEnvironmentApproval

	Reason string
	DeploymentStatus string
	OperationStatus  string

	Attempt int64

	// sometimes dates are not valid here
	QueuedOn string `json:"queuedOn,omitempty"`
	StartedOn string `json:"startedOn,omitempty"`
	CompletedOn string `json:"completedOn,omitempty"`

	RequestedBy  IdentifyRef
	RequestedFor IdentifyRef

	Links Links `json:"_links"`
}

type ReleaseDeploymentEnvironment struct {
	Id    int64
	Name  string
}


func (d *ReleaseDeployment) ApprovedBy() (string) {
	var approverList []string
	for _, approval := range d.PreDeployApprovals {
		if !approval.IsAutomated {
			if approval.ApprovedBy.DisplayName != "" {
				approverList = append(approverList, approval.ApprovedBy.DisplayName)
			}
		}
	}

	return strings.Join(approverList[:],",")
}

func (d *ReleaseDeployment) QueuedOnTime() (*time.Time) {
	return parseTime(d.QueuedOn)
}

func (d *ReleaseDeployment) StartedOnTime() (*time.Time) {
	return parseTime(d.StartedOn)
}

func (d *ReleaseDeployment) CompletedOnTime() (*time.Time) {
	return parseTime(d.CompletedOn)
}

func (c *AzureDevopsClient) ListReleaseDeployments(project string, releaseDefinitionId int64) (list ReleaseDeploymentList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/release/deployments?api-version=5.0-preview.2&isDeleted=false&$expand=94&definitionId=%s&$top=%v",
		url.QueryEscape(project),
		url.QueryEscape(int64ToString(releaseDefinitionId)),
		url.QueryEscape(int64ToString(c.LimitDeploymentPerDefinition)),
	)
	response, err := c.restVsrm().R().Get(url)
	if err := c.checkResponse(response, err); err != nil {
		error = err
		return
	}

	err = json.Unmarshal(response.Body(), &list)
	if err != nil {
		error = err
		return
	}

	return
}
