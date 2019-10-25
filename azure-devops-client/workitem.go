package AzureDevopsClient

import (
	"encoding/json"
)

type WorkItem struct {
	// We need only these fields for bugs. Add more as needed.
	Id     int64          `json:"id"`
	Fields WorkItemFields `json:"fields"`
}

type WorkItemFields struct {
	Title        string `json:"System.Title"`
	Path         string `json:"System.AreaPath"`
	CreatedDate  string `json:"System.CreatedDate"`
	AcceptedDate string `json:"Microsoft.VSTS.CodeReview.AcceptedDate"`
	ResolvedDate string `json:"Microsoft.VSTS.Common.ResolvedDate"`
	ClosedDate   string `json:"Microsoft.VSTS.Common.ClosedDate"`
}

func (c *AzureDevopsClient) GetWorkItem(workItemUrl string) (workItem WorkItem, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	response, err := c.rest().R().Get(workItemUrl)
	if err := c.checkResponse(response, err); err != nil {
		error = err
		return
	}

	err = json.Unmarshal(response.Body(), &workItem)
	if err != nil {
		error = err
	}

	return
}
