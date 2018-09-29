package AzureDevopsClient

import (
	"fmt"
	"net/url"
	"encoding/json"
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

