package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type RepositoryList struct {
	Count int          `json:"count"`
	List  []Repository `json:"value"`
}

type Repository struct {
	Id         string
	Name       string
	Url        string
	State      string
	WellFormed string
	Revision   int64
	Visibility string
	Size       int64

	Links Links `json:"_links"`
}

type RepositoryCommitList struct {
	Count int                `json:"count"`
	List  []RepositoryCommit `json:"value"`
}

type RepositoryCommit struct {
	CommitId         string
	Author           Author
	Committer        Author
	Comment          string
	CommentTruncated bool
	ChangeCounts     struct {
		Add    int64
		Edit   int64
		Delete int64
	}

	Url       string
	RemoteUrl string
}

type RepositoryPushList struct {
	Count int              `json:"count"`
	List  []RepositoryPush `json:"value"`
}

type RepositoryPush struct {
	PushId int64
}

func (c *AzureDevopsClient) ListRepositories(project string) (list RepositoryList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/git/repositories",
		url.QueryEscape(project),
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

func (c *AzureDevopsClient) ListCommits(project string, repository string, fromDate time.Time) (list RepositoryCommitList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"_apis/git/repositories/%s/commits?searchCriteria.fromDate=%s&api-version=%v",
		url.QueryEscape(repository),
		url.QueryEscape(fromDate.Format(time.RFC3339)),
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

func (c *AzureDevopsClient) ListPushes(project string, repository string, fromDate time.Time) (list RepositoryPushList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"_apis/git/repositories/%s/pushes?searchCriteria.fromDate=%s&api-version=%v",
		url.QueryEscape(repository),
		url.QueryEscape(fromDate.Format(time.RFC3339)),
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
