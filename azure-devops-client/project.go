package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type ProjectList struct {
	Count int       `json:"count"`
	List  []Project `json:"value"`
}

type Project struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
	State       string `json:"state"`
	WellFormed  string `json:"wellFormed"`
	Revision    int64  `json:"revision"`
	Visibility  string `json:"visibility"`

	RepositoryList RepositoryList
}

func (c *AzureDevopsClient) ListProjects() (list ProjectList, err error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	// Retrieve the first 1000 projects
	firstBatchURL := fmt.Sprintf(
		"_apis/projects?$top=1000&api-version=%v",
		url.QueryEscape(c.ApiVersion),
	)
	firstBatchResponse, err := c.rest().R().Get(firstBatchURL)
	if err := c.checkResponse(firstBatchResponse, err); err != nil {
		return list, err
	}

	var firstBatch ProjectList
	err = json.Unmarshal(firstBatchResponse.Body(), &firstBatch)
	if err != nil {
		return list, err
	}

	list.Count = firstBatch.Count
	list.List = append(list.List, firstBatch.List...)

	// Fetch the remaining projects, if any
	fetchedProjects := firstBatch.Count
	for int64(fetchedProjects) < c.LimitProject {
		remainingProjectsURL := fmt.Sprintf(
			"_apis/projects?$top=%v&$skip=%v&api-version=%v",
			1000,
			fetchedProjects,
			url.QueryEscape(c.ApiVersion),
		)

		remainingProjectsResponse, err := c.rest().R().Get(remainingProjectsURL)
		if err := c.checkResponse(remainingProjectsResponse, err); err != nil {
			return list, err
		}

		var remainingProjects ProjectList
		err = json.Unmarshal(remainingProjectsResponse.Body(), &remainingProjects)
		if err != nil {
			return list, err
		}

		remainingProjectsCount := remainingProjects.Count
		if remainingProjectsCount == 0 {
			break // No more projects to fetch
		}

		list.Count += remainingProjectsCount
		list.List = append(list.List, remainingProjects.List...)

		fetchedProjects += remainingProjectsCount
	}

	for key, project := range list.List {
		list.List[key].RepositoryList, _ = c.ListRepositories(project.Id)
	}

	return list, nil
}
