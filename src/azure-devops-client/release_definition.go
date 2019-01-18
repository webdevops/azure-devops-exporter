package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type ReleaseDefinitionList struct {
	Count int                 `json:"count"`
	List  []ReleaseDefinition `json:"value"`
}

type ReleaseDefinition struct {
	Id                int64 `json:"id"`
	Name              string
	Path              string
	ReleaseNameFormat string

	Environments []ReleaseDefinitionEnvironment

	Links Links `json:"_links"`
}

type ReleaseDefinitionEnvironment struct {
	Id   int64
	Name string
	Rank int64

	Owner          IdentifyRef
	CurrentRelease struct {
		Id  int64
		Url string
	}

	BadgeUrl string
}

func (c *AzureDevopsClient) ListReleaseDefinitions(project string) (list ReleaseDefinitionList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/release/definitions?api-version=5.0-preview.3&isDeleted=false&$top=%v&$expand=2",
		url.QueryEscape(project),
		url.QueryEscape(int64ToString(c.LimitReleaseDefinitionsPerProject)),
	)
	response, err := c.restVsrm().R().Get(url)
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
