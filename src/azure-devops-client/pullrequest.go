package AzureDevopsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type PullRequestList struct {
	Count int           `json:"count"`
	List  []PullRequest `json:"value"`
}

type PullRequest struct {
	Id           int64 `json:"pullRequestId"`
	CodeReviewId int64 `json:"codeReviewId"`

	Title       string
	Description string
	Uri         string
	Url         string

	CreatedBy IdentifyRef

	SourceRefName string
	TargetRefName string

	Reviewers []PullRequestReviewer

	Status       string `json:"status"`
	CreationDate time.Time
	ClosedDate   time.Time

	Links Links `json:"_links"`
}

type PullRequestReviewer struct {
	Vote        int64
	DisplayName string
}

func (v *PullRequest) GetVoteSummary() map[string]int {
	ret := map[string]int{
		"approved":            0,
		"approvedSuggestions": 0,
		"none":                0,
		"waitingForAuthor":    0,
		"rejected":            0,
	}

	for _, reviewer := range v.Reviewers {
		switch reviewer.Vote {
		case 10:
			ret["approved"]++
		case 5:
			ret["approvedSuggestions"]++
		case 0:
			ret["none"]++
		case -5:
			ret["waitingForAuthor"]++
		case -10:
			ret["rejected"]++
		}
	}

	return ret
}

func (c *AzureDevopsClient) ListPullrequest(project, repositoryId string) (list PullRequestList, error error) {
	defer c.concurrencyUnlock()
	c.concurrencyLock()

	url := fmt.Sprintf(
		"%v/_apis/git/repositories/%v/pullrequests?api-version=4.1&searchCriteria.status=active",
		url.QueryEscape(project),
		url.QueryEscape(repositoryId),
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
