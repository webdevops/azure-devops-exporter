package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type AgentQueueList struct {
	Count int          `json:"count"`
	List  []AgentQueue `json:"value"`
}

type AgentQueue struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Pool struct {
		Id       int64
		Scope    string
		Name     string
		IsHosted bool
		PoolType string
		Size     int64
	}
}

func (c *AzureDevopsClient) ListAgentQueues(project string) (list AgentQueueList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/distributedtask/queues",
		url.QueryEscape(project),
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

	return
}

type AgentPoolList struct {
	Count int              `json:"count"`
	Value []AgentPoolEntry `json:"value"`
}

type AgentPoolEntry struct {
	CreatedOn     time.Time `json:"createdOn"`
	AutoProvision bool      `json:"autoProvision"`
	AutoUpdate    bool      `json:"autoUpdate"`
	AutoSize      bool      `json:"autoSize"`
	CreatedBy     struct {
		DisplayName string `json:"displayName"`
		URL         string `json:"url"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
		} `json:"_links"`
		ID         string `json:"id"`
		UniqueName string `json:"uniqueName"`
		ImageURL   string `json:"imageUrl"`
		Descriptor string `json:"descriptor"`
	} `json:"createdBy"`
	Owner struct {
		DisplayName string `json:"displayName"`
		URL         string `json:"url"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
		} `json:"_links"`
		ID         string `json:"id"`
		UniqueName string `json:"uniqueName"`
		ImageURL   string `json:"imageUrl"`
		Descriptor string `json:"descriptor"`
	} `json:"owner"`
	ID       int64  `json:"id"`
	Scope    string `json:"scope"`
	Name     string `json:"name"`
	IsHosted bool   `json:"isHosted"`
	PoolType string `json:"poolType"`
	Size     int    `json:"size"`
	IsLegacy bool   `json:"isLegacy"`
	Options  string `json:"options"`
}

type AgentPoolAgentList struct {
	Count int              `json:"count"`
	List  []AgentPoolAgent `json:"value"`
}

type AgentPoolAgent struct {
	Id                int64
	Enabled           bool
	MaxParallelism    int64
	Name              string
	OsDescription     string
	ProvisioningState string
	Status            string
	Version           string
	CreatedOn         time.Time
	AssignedRequest   JobRequest
}

type JobRequest struct {
	RequestId    int64
	Demands      []string
	QueueTime    time.Time
	AssignTime   *time.Time
	ReceiveTime  time.Time
	LockedUntil  time.Time
	ServiceOwner string
	HostId       string
	ScopeId      string
	PlanType     string
	PlanId       string
	JobId        string
	Definition   struct {
		Id    int64
		Name  string
		Links Links `json:"_links"`
	}
}

func (c *AzureDevopsClient) ListAgentPools() (list AgentPoolList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"/_apis/distributedtask/pools?api-version=%s",
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
		return
	}

	return
}

func (c *AzureDevopsClient) ListAgentPoolAgents(agentPoolId int64) (list AgentPoolAgentList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"/_apis/distributedtask/pools/%v/agents?includeCapabilities=false&includeAssignedRequest=true",
		fmt.Sprintf("%d", agentPoolId),
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

	return
}

type AgentPoolJobList struct {
	Count int          `json:"count"`
	List  []JobRequest `json:"value"`
}

func (c *AzureDevopsClient) ListAgentPoolJobs(agentPoolId int64) (list AgentPoolJobList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"/_apis/distributedtask/pools/%v/jobrequests",
		fmt.Sprintf("%d", agentPoolId),
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

	return
}
