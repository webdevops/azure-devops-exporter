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
	ReleaseNameFormat string `json:"releaseNameFormat"`

	Environments []ReleaseDefinitionEnvironment

	LastRelease Release `json:"lastRelease"`

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
	}  `json:"currentRelease"`

	BadgeUrl string `json:"badgeUrl"`
}

func (c *AzureDevopsClient) ListReleaseDefinitions(project string) (list ReleaseDefinitionList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/release/definitions?api-version=%v&isDeleted=false&$top=%v&$expand=environments,lastRelease",
		url.QueryEscape(project),
		url.QueryEscape(c.ApiVersion),
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
