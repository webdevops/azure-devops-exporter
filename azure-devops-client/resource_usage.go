package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type ResourceUsageBuild struct {
	DistributedTaskAgents *int `json:"distributedTaskAgents"`
	PaidPrivateAgentSlots *int `json:"paidPrivateAgentSlots"`
	TotalUsage            *int `json:"totalUsage"`
	XamlControllers       *int `json:"xamlControllers"`
}

func (c *AzureDevopsClient) GetResourceUsageBuild() (ret ResourceUsageBuild, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"/_apis/build/resourceusage?api-version=%v",
		// FIXME: hardcoded api version
		url.QueryEscape("5.1-preview.2"),
	)
	response, err := c.rest().R().Get(url)
	if err := c.checkResponse(response, err); err != nil {
		error = err
		return
	}

	err = json.Unmarshal(response.Body(), &ret)
	if err != nil {
		error = err
		return
	}

	return
}
