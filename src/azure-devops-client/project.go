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

func (c *AzureDevopsClient) ListProjects() (list ProjectList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/projects",
		url.QueryEscape(*c.collection),
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

	for key, project := range list.List {
		list.List[key].RepositoryList, _ = c.ListRepositories(project.Name)
	}

	return
}
