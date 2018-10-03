package AzureDevopsClient

import (
	"fmt"
	"time"
	"net/url"
	"encoding/json"
)

type ReleaseList struct {
	Count int `json:"count"`
	List []Release `json:"value"`
}

type Release struct {
	Id int64 `json:"id"`
	Name string

	Definition struct {
		Id int64
		Name string
		Links Links `json:"_links"`
	} `json:"releaseDefinition"`

	Project Project `json:"projectReference"`

	Queue AgentPoolQueue

	Reason string
	Result string
	Status string
	QueueTime time.Time
	QueuePosition string
	StartTime time.Time
	FinishTime time.Time
	Uri string
	Url string

	RequestedBy IdentifyRef
	RequestedFor IdentifyRef

	Links Links `json:"_links"`
}

func (r *Release) QueueDuration() time.Duration {
	return r.StartTime.Sub(r.QueueTime)
}


func (c *AzureDevopsClient) ListReleases(project string) (list ReleaseList, error error) {
	url := fmt.Sprintf(
		"%v/_apis/release/releases?api-version=4.1-preview.6&isDeleted=false&%24expand=94",
		url.QueryEscape(project),
	)
	response, err := c.restDev().R().Get(url)

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
