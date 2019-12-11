package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Query struct {
	Path string `json:"path"`
}

type WorkItemInfoList struct {
	List []WorkItemInfo `json:"workItems"`
}

type WorkItemInfo struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

func (c *AzureDevopsClient) QueryWorkItems(queryPath, projectId string) (list WorkItemInfoList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/wit/wiql/%v?api-version=%v",
		projectId,
		queryPath,
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
	}

	return
}
