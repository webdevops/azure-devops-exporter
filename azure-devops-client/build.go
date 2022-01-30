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

type TimelineRecordList struct {
	List []TimelineRecord `json:"records"`
}

type TimelineRecord struct {
	RecordType   string  `json:"type"`
	Name         string  `json:"name"`
	Id           string  `json:"id"`
	ParentId     string  `json:"parentId"`
	ErrorCount   float64 `json:"errorCount"`
	WarningCount float64 `json:"warningCount"`
	Result       string  `json:"result"`
	WorkerName   string  `json:"workerName"`
	Identifier   string  `json:"identifier"`
	StartTime    time.Time
	FinishTime   time.Time
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
		"%v/_apis/build/definitions?api-version=%v&$top=9999",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
	)
	response, err := c.rest().R().Get(url)
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
		"%v/_apis/build/builds?api-version=%v&maxBuildsPerDefinition=%s&deletedFilter=excludeDeleted",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
		url.QueryEscape(int64ToString(c.LimitBuildsPerDefinition)),
	)
	response, err := c.rest().R().Get(url)
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
		"%v/_apis/build/builds?api-version=%v&maxBuildsPerDefinition=%s&deletedFilter=excludeDeleted",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
		url.QueryEscape("1"),
	)
	response, err := c.rest().R().Get(url)
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
		"%v/_apis/build/builds?api-version=%v&minTime=%s&$top=%v&queryOrder=finishTimeDescending",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
		url.QueryEscape(minTime.Format(time.RFC3339)),
		url.QueryEscape(int64ToString(c.LimitBuildsPerProject)),
	)
	response, err := c.rest().R().Get(url)
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

func (c *AzureDevopsClient) ListBuildHistoryWithStatus(project string, minTime time.Time, statusFilter string) (list BuildList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/builds?api-version=%v&minTime=%s&statusFilter=%v",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
		url.QueryEscape(minTime.Format(time.RFC3339)),
		url.QueryEscape(statusFilter),
	)
	response, err := c.rest().R().Get(url)
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

func (c *AzureDevopsClient) ListBuildTimeline(project string, buildID string) (list TimelineRecordList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/builds/%v/Timeline",
		url.QueryEscape(project),
		url.QueryEscape(buildID),
	)
	response, err := c.rest().R().Get(url)
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
