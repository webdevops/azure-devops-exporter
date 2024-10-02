package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type ReleaseList struct {
	Count int       `json:"count"`
	List  []Release `json:"value"`
}

type Release struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`

	Definition struct {
		Id    int64  `json:"id"`
		Name  string `json:"name"`
		Links Links  `json:"_links"`
	} `json:"releaseDefinition"`

	Project Project `json:"projectReference"`

	Queue AgentPoolQueue `json:"queue"`

	Reason        string    `json:"reason"`
	Result        bool      `json:"result"`
	Status        string    `json:"status"`
	CreatedOn     time.Time `json:"createdOn"`
	QueueTime     time.Time `json:"queueTime"`
	QueuePosition string    `json:"queuePosition"`
	StartTime     time.Time `json:"startTime"`
	FinishTime    time.Time `json:"finishTime"`
	Uri           string    `json:"uri"`
	Url           string    `json:"url"`

	Artifacts    []ReleaseArtifact    `json:"artifacts"`
	Environments []ReleaseEnvironment `json:"environments"`

	RequestedBy  IdentifyRef `json:"requestedBy"`
	RequestedFor IdentifyRef `json:"requestedFor"`

	Links Links `json:"_links"`
}

type ReleaseArtifact struct {
	SourceId string `json:"sourceId"`
	Type     string `json:"type"`
	Alias    string `json:"alias"`

	DefinitionReference struct {
		Definition struct {
			Id   string
			Name string
		}

		Project struct {
			Id   string
			Name string
		}

		Repository struct {
			Id   string
			Name string
		}

		Version struct {
			Id   string
			Name string
		}

		Branch struct {
			Id   string
			Name string
		}
	} `json:"definitionReference"`
}

type ReleaseEnvironment struct {
	Id                      int64  `json:"id"`
	ReleaseId               int64  `json:"releaseId"`
	DefinitionEnvironmentId int64  `json:"definitionEnvironmentId"`
	Name                    string `json:"name"`
	Status                  string `json:"status"`
	Rank                    int64  `json:"rank"`

	TriggerReason string `json:"triggerReason"`

	DeploySteps []ReleaseEnvironmentDeployStep `json:"deploySteps"`

	PreDeployApprovals  []ReleaseEnvironmentApproval `json:"preDeployApprovals"`
	PostDeployApprovals []ReleaseEnvironmentApproval `json:"postDeployApprovals"`

	CreatedOn      time.Time `json:"createdOn"`
	QueuedOn       time.Time `json:"queuedOn"`
	LastModifiedOn time.Time `json:"lastModifiedOn"`

	TimeToDeploy float64 `json:"timeToDeploy"`
}

type ReleaseEnvironmentDeployStep struct {
	Id              int64
	DeploymentId    int64
	Attemt          int64
	Reason          string
	Status          string
	OperationStatus string

	ReleaseDeployPhases []ReleaseEnvironmentDeployStepPhase

	QueuedOn       time.Time
	LastModifiedOn time.Time
}

type ReleaseEnvironmentDeployStepPhase struct {
	Id        int64
	PhaseId   string
	Name      string
	Rank      int64
	PhaseType string
	Status    string
	StartedOn time.Time `json:"startedOn"`
}

type ReleaseEnvironmentApproval struct {
	Id               int64
	Revision         int64
	ApprovalType     string
	Status           string
	Comments         string
	IsAutomated      bool
	IsNotificationOn bool
	TrialNumber      int64 `json:"trialNumber"`
	Attempt          int64 `json:"attempt"`
	Rank             int64 `json:"rank"`

	Approver   IdentifyRef `json:"approver"`
	ApprovedBy IdentifyRef `json:"approvedBy"`

	CreatedOn  time.Time `json:"createdOn"`
	ModifiedOn time.Time `json:"modifiedOn"`
}

func (r *Release) QueueDuration() time.Duration {
	return r.StartTime.Sub(r.QueueTime)
}

func (c *AzureDevopsClient) ListReleases(project string, releaseDefinitionId int64) (list ReleaseList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/release/releases?api-version=%v&isDeleted=false&$expand=94&definitionId=%s&$top=%v",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
		url.QueryEscape(int64ToString(releaseDefinitionId)),
		url.QueryEscape(int64ToString(c.LimitReleasesPerDefinition)),
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

func (c *AzureDevopsClient) ListReleaseHistory(project string, minTime time.Time) (list ReleaseList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/release/releases?api-version=%v&isDeleted=false&$expand=94&minCreatedTime=%s&$top=%v&queryOrder=descending",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
		url.QueryEscape(minTime.Format(time.RFC3339)),
		url.QueryEscape(int64ToString(c.LimitReleasesPerProject)),
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

	continuationToken := response.Header().Get("x-ms-continuationtoken")

	for continuationToken != "" {
		continuationUrl := fmt.Sprintf(
			"%v&continuationToken=%v",
			url,
			continuationToken,
		)

		response, err = c.restVsrm().R().Get(continuationUrl)
		if err := c.checkResponse(response, err); err != nil {
			error = err
			return
		}

		var tmpList ReleaseList
		err = json.Unmarshal(response.Body(), &tmpList)
		if err != nil {
			error = err
			return
		}

		list.Count += tmpList.Count
		list.List = append(list.List, tmpList.List...)

		continuationToken = response.Header().Get("x-ms-continuationtoken")
	}

	return
}
