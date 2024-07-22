package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

type TagList struct {
	Count int      `json:"count"`
	List  []string `json:"value"`
}

type Tag struct {
	Name  string
	Value string
	Type  string
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
	State        string  `json:"state"`
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

	requestUrl := ""

	if statusFilter == "all" {
		requestUrl = fmt.Sprintf(
			"%v/_apis/build/builds?api-version=%v&statusFilter=%v",
			url.QueryEscape(project),
			url.QueryEscape(c.ApiVersion),
			url.QueryEscape(statusFilter),
		)
	} else {
		requestUrl = fmt.Sprintf(
			"%v/_apis/build/builds?api-version=%v&minTime=%s&statusFilter=%v",
			url.QueryEscape(project),
			url.QueryEscape(c.ApiVersion),
			url.QueryEscape(minTime.Format(time.RFC3339)),
			url.QueryEscape(statusFilter),
		)
	}

	response, err := c.rest().R().Get(requestUrl)
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

func (c *AzureDevopsClient) ListBuildTags(project string, buildID string) (list TagList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/build/builds/%v/tags",
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

func extractTagKeyValue(tag string) (k string, v string, error error) {
	parts := strings.Split(tag, "=")
	if len(parts) != 2 {
		error = fmt.Errorf("could not extract key value pair from tag '%s'", tag)
		return
	}
	k = parts[0]
	v = parts[1]
	return
}

func extractTagSchema(tagSchema string) (n string, t string, error error) {
	parts := strings.Split(tagSchema, ":")
	if len(parts) != 2 {
		error = fmt.Errorf("could not extract type from tag schema '%s'", tagSchema)
		return
	}
	n = parts[0]
	t = parts[1]
	return
}

func (t *TagList) Extract() (tags map[string]string, error error) {
	tags = make(map[string]string)
	for _, t := range t.List {
		k, v, err := extractTagKeyValue(t)
		if err != nil {
			error = err
			return
		}
		tags[k] = v
	}
	return
}

func (t *TagList) Parse(tagSchema []string) (pTags []Tag, error error) {
	tags, err := t.Extract()
	if err != nil {
		error = err
		return
	}
	for _, ts := range tagSchema {
		name, _type, err := extractTagSchema(ts)
		if err != nil {
			error = err
			return
		}

		value, isPresent := tags[name]
		if isPresent {
			pTags = append(pTags, Tag{
				Name:  name,
				Value: value,
				Type:  _type,
			})
		}
	}
	return
}
