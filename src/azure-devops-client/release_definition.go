package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type ReleaseDefinitionList struct {
	Count int `json:"count"`
	List []ReleaseDefinition `json:"value"`
}

type ReleaseDefinition struct {
	Id int64 `json:"id"`
	Name string
	ReleaseNameFormat string
	Links Links `json:"_links"`
}

func (c *AzureDevopsClient) ListReleaseDefinitions(project string) (list ReleaseDefinitionList, error error) {
	url := fmt.Sprintf(
		"%v/_apis/release/definitions?api-version=5.0-preview.3&isDeleted=false&$top=100&isDeleted=false",
		url.QueryEscape(project),
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
