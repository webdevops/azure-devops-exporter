package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Query struct {
	Path string `json:"path"`
}

type WorkItemList struct {
	List []WorkItemInfo `json:"workItems"`
}

type WorkItemInfo struct {
	id  int    `json:"id"`
	url string `json:"url"`
}

// type WorkItem struct {
// 	// We need only these fields for bugs. Add more as needed.
// 	Id           string `json:"id"`
// 	Title        string `json:"title"`
// 	Team         string `json:"team"`
// 	Rank         string `json:"rank"`
// 	DateCreated  string `json:"dateCreated"`
// 	DateAccepted string `json:"dateAccepted"`
// 	DateResolved string `json:"dateResolved"`
// 	DateClosed   string `json:"dateClosed"`
// }

func (c *AzureDevopsClient) QueryWorkItems(queryPath, projectId string) (list WorkItemList, error error) {
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
