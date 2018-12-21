package AzureDevopsClient

import (
	"fmt"
	"net/url"
	"encoding/json"
	"time"
)

type RepositoryList struct {
	Count int `json:"count"`
	List []Repository `json:"value"`
}

type Repository struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Url string `json:"url"`
	State string `json:"state"`
	WellFormed string `json:"wellFormed"`
	Revision int64 `json:"revision"`
	Visibility string `json:"visibility"`

	Links Links `json:"_links"`
}

type RepositoryCommitList struct {
	Count int `json:"count"`
	List []RepositoryCommit `json:"value"`
}

type RepositoryCommit struct {
	CommitId string
	Author Author
	Committer Author
	Comment string
	CommentTruncated bool
	ChangeCounts struct {
		Add int64
		Edit int64
		Delete int64
	}

	Url string
	RemoteUrl string
}

func (c *AzureDevopsClient) ListRepositories(project string) (list RepositoryList, error error) {
	url := fmt.Sprintf(
		"%v/_apis/git/repositories?api-version=4.1",
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

func (c *AzureDevopsClient) ListCommits(project string, repository string, fromDate time.Time) (list RepositoryCommitList, error error) {
	url := fmt.Sprintf(
		"_apis/git/repositories/%s/commits?searchCriteria.fromDate=%s&api-version=5.0-preview.1",
		url.QueryEscape(repository),
		url.QueryEscape(fromDate.Format(time.RFC3339)),
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
