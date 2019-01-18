package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type BuildDefinitionList struct {
	Count int               `json:"count"`
	List  []BuildDefinition `json:"value"`
}

type BuildDefinition struct {
	Id              int64
	Name            string
	Path            string
	Revision        int64
	QueueStatus     string
	BuildNameFormat string
	Links           Links `json:"_links"`
}

type BuildList struct {
	Count int     `json:"count"`
	List  []Build `json:"value"`
}

type Build struct {
	Id                  int64  `json:"id"`
	BuildNumber         string `json:"buildNumber"`
	BuildNumberRevision int64  `json:"buildNumberRevision"`
	Quality             string `json:"quality"`

	Definition BuildDefinition

	Project Project

	Queue AgentPoolQueue

	Reason        string
	Result        string
	Status        string
	QueueTime     time.Time
	QueuePosition string
	StartTime     time.Time
	FinishTime    time.Time
	Uri           string
	Url           string
	SourceBranch  string
	SourceVersion string

	RequestedBy  IdentifyRef
	RequestedFor IdentifyRef

	Links Links `json:"_links"`
}

func (b *Build) QueueDuration() time.Duration {
	return b.StartTime.Sub(b.QueueTime)
}

func (c *AzureDevopsClient) ListBuildDefinitions(project string) (list BuildDefinitionList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/definitions?api-version=4.1&$top=9999",
		url.QueryEscape(project),
	)
	response, err := c.restDev().R().Get(url)
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

func (c *AzureDevopsClient) ListBuilds(project string) (list BuildList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/builds?api-version=4.1&maxBuildsPerDefinition=%s&deletedFilter=excludeDeleted",
		url.QueryEscape(project),
		url.QueryEscape(int64ToString(c.LimitBuildsPerDefinition)),
	)
	response, err := c.restDev().R().Get(url)
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

func (c *AzureDevopsClient) ListLatestBuilds(project string) (list BuildList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/builds?api-version=4.1&maxBuildsPerDefinition=%s&deletedFilter=excludeDeleted",
		url.QueryEscape(project),
		url.QueryEscape("1"),
	)
	response, err := c.restDev().R().Get(url)
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

func (c *AzureDevopsClient) ListBuildHistory(project string, minTime time.Time) (list BuildList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/builds?api-version=4.1&minTime=%s",
		url.QueryEscape(project),
		url.QueryEscape(minTime.Format(time.RFC3339)),
	)
	response, err := c.restDev().R().Get(url)
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
