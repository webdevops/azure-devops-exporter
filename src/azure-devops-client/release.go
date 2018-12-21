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
	Id   int64 `json:"id"`
	Name string

	Definition struct {
		Id    int64
		Name  string
		Links Links `json:"_links"`
	} `json:"releaseDefinition"`

	Project Project `json:"projectReference"`

	Queue AgentPoolQueue

	Reason        string
	Result        bool
	Status        string
	QueueTime     time.Time
	QueuePosition string
	StartTime     time.Time
	FinishTime    time.Time
	Uri           string
	Url           string

	RequestedBy  IdentifyRef
	RequestedFor IdentifyRef

	Links Links `json:"_links"`
}

func (r *Release) QueueDuration() time.Duration {
	return r.StartTime.Sub(r.QueueTime)
}

func (c *AzureDevopsClient) ListReleases(project string, releaseDefinitionId int64) (list ReleaseList, error error) {
	url := fmt.Sprintf(
		"%v/_apis/release/releases?api-version=4.1-preview.6&isDeleted=false&expand=94&$top=100&latestAttemptsOnly=true&definitionIdFilter=%s",
		url.QueryEscape(project),
		url.QueryEscape(string(releaseDefinitionId)),
	)
	response, err := c.restVsrm().R().Get(url)

	if err != nil {
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
