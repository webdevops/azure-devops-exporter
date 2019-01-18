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
	response, err := c.restDev().R().Get(url)
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

	AssignedRequest struct {
		RequestId    int64
		Demands      []string
		QueueTime    time.Time
		AssignTime   time.Time
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
