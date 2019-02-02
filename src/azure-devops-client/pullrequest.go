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
	Labels []PullRequestLabels

	Status       string `json:"status"`
	CreationDate time.Time
	ClosedDate   time.Time

	Links Links `json:"_links"`
}

type PullRequestReviewer struct {
	Vote        int64
	DisplayName string
}

type PullRequestLabels struct {
	Id string
	Name string
	Active bool
}

type PullRequestVoteSummary struct {
	Approved int64
	ApprovedSuggestions int64
	None int64
	WaitingForAuthor int64
	Rejected int64
	Count int64
}

func (v *PullRequest) GetVoteSummary() PullRequestVoteSummary {
	ret := PullRequestVoteSummary{}

	for _, reviewer := range v.Reviewers {
		ret.Count++
		switch reviewer.Vote {
		case 10:
			ret.Approved++
		case 5:
			ret.ApprovedSuggestions++
		case 0:
			ret.None++
		case -5:
			ret.WaitingForAuthor++
		case -10:
			ret.Rejected++
		}
	}

	return ret
}

func (v *PullRequestVoteSummary) HumanizeString() (status string) {
	status = "None"

	if v.Rejected >= 1 {
		status = "Rejected"
	} else if v.WaitingForAuthor >= 1 {
		status = "WaitingForAuthor"
	} else if v.ApprovedSuggestions >= 1 {
		status = "ApprovedSuggestions"
	} else if v.Approved >= 1 {
		status = "Approved"
	}

	return
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
