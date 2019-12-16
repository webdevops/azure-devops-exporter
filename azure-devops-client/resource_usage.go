package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type (
	ResourceUsageBuild struct {
		DistributedTaskAgents *int `json:"distributedTaskAgents"`
		PaidPrivateAgentSlots *int `json:"paidPrivateAgentSlots"`
		TotalUsage            *int `json:"totalUsage"`
		XamlControllers       *int `json:"xamlControllers"`
	}

	ResourceUsageAgent struct {
		Data struct {
			Provider struct {
				IncludeResourceLimitsSection bool `json:"includeResourceLimitsSection"`
				IncludeConcurrentJobsSection bool `json:"includeConcurrentJobsSection"`

				ResourceUsages []ResourceUsageAgentUsageRow `json:"resourceUsages"`

				TaskHubLicenseDetails struct {
					FreeLicenseCount *float64 `json:"freeLicenseCount"`
					FreeHostedLicenseCount *float64 `json:"freeHostedLicenseCount"`
					EnterpriseUsersCount *float64 `json:"enterpriseUsersCount"`
					PurchasedLicenseCount *float64 `json:"purchasedLicenseCount"`
					PurchasedHostedLicenseCount *float64 `json:"purchasedHostedLicenseCount"`
					HostedLicensesArePremium bool `json:"hostedLicensesArePremium"`
					TotalLicenseCount *float64 `json:"totalLicenseCount"`
					HasLicenseCountEverUpdated bool `json:"hasLicenseCountEverUpdated"`
					MsdnUsersCount *float64 `json:"msdnUsersCount"`
					HostedAgentMinutesFreeCount *float64 `json:"hostedAgentMinutesFreeCount"`
					HostedAgentMinutesUsedCount *float64 `json:"hostedAgentMinutesUsedCount"`
					FailedToReachAllProviders bool `json:"failedToReachAllProviders"`
					TotalPrivateLicenseCount *float64 `json:"totalPrivateLicenseCount"`
					TotalHostedLicenseCount *float64 `json:"totalHostedLicenseCount"`
				} `json:"taskHubLicenseDetails"`

			} `json:"ms.vss-build-web.build-queue-hub-data-provider"`
		} `json:"data"`
	}

	ResourceUsageAgentUsageRow struct {
		ResourceLimit struct {
			ResourceLimitsData struct {
				FreeCount string `json:"freeCount"`
				PurchasedCount string `json:"purchasedCount"`
			} `json:"resourceLimitsData"`

			HostId string `json:"hostId"`
			ParallelismTag string `json:"parallelismTag"`
			IsHosted bool `json:"isHosted"`
			TotalCount float64 `json:"totalCount"`
			IsPremium bool `json:"IsPremium"`

		} `json:"resourceLimit"`
	}
)

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

func (c *AzureDevopsClient) GetResourceUsageAgent() (ret ResourceUsageAgent, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"/_apis/Contribution/dataProviders/query?api-version=%v",
		// FIXME: hardcoded api version
		url.QueryEscape("5.1-preview.1"),
	)

	payload := `{"contributionIds": ["ms.vss-build-web.build-queue-hub-data-provider"]}`

	req := c.rest().NewRequest()
	req.SetHeader("Content-Type", "application/json")
	req.SetBody(payload)
	response, err := req.Post(url)
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
