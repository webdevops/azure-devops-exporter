package AzureDevopsClient

import (
	"fmt"
	"errors"
	"github.com/go-resty/resty"
	"sync/atomic"
)

type AzureDevopsClient struct {
	organization *string
	collection   *string
	accessToken  *string

	restClient     *resty.Client
	restClientDev  *resty.Client
	restClientVsrm *resty.Client

	semaphore   chan bool
	concurrency int64

	RequestCount uint64
	RequestRetries int

	LimitBuildsPerDefinition int64
	LimitReleasesPerDefinition int64
	LimitDeploymentPerDefinition int64
	LimitReleaseDefinitionsPerProject int64
}

func NewAzureDevopsClient() *AzureDevopsClient {
	c := AzureDevopsClient{}
	c.Init()

	return &c
}

func (c *AzureDevopsClient) Init() {
	collection := "DefaultCollection"
	c.collection = &collection
	c.RequestCount = 0
	c.SetRetries(3)
	c.SetConcurrency(10)

	c.LimitBuildsPerDefinition = 10
	c.LimitReleasesPerDefinition = 100
	c.LimitDeploymentPerDefinition = 100
	c.LimitReleaseDefinitionsPerProject = 100
}

func (c *AzureDevopsClient) SetConcurrency(v int64) {
	c.concurrency = v
	c.semaphore = make(chan bool, c.concurrency)
}
func (c *AzureDevopsClient) SetRetries(v int) {
	c.RequestRetries = v

	if c.restClient != nil {
		c.restClient.SetRetryCount(c.RequestRetries)
	}

	if c.restClientDev != nil {
		c.restClientDev.SetRetryCount(c.RequestRetries)
	}

	if c.restClientVsrm != nil {
		c.restClientVsrm.SetRetryCount(c.RequestRetries)
	}
}

func (c *AzureDevopsClient) SetOrganization(url string) {
	c.organization = &url
}

func (c *AzureDevopsClient) SetAccessToken(token string) {
	c.accessToken = &token
}

func (c *AzureDevopsClient) rest() *resty.Client {
	if c.restClient == nil {
		c.restClient = resty.New()
		c.restClient.SetHostURL(fmt.Sprintf("https://%v.visualstudio.com/", *c.organization))
		c.restClient.SetHeader("Accept", "application/json")
		c.restClient.SetBasicAuth("", *c.accessToken)
		c.restClient.SetRetryCount(c.RequestRetries)
		c.restClient.OnBeforeRequest(c.restOnBeforeRequest)
		c.restClient.OnAfterResponse(c.restOnAfterResponse)
	}

	return c.restClient
}

func (c *AzureDevopsClient) restDev() *resty.Client {
	if c.restClientDev == nil {
		c.restClientDev = resty.New()
		c.restClientDev.SetHostURL(fmt.Sprintf("https://dev.azure.com/%v/", *c.organization))
		c.restClientDev.SetHeader("Accept", "application/json")
		c.restClientDev.SetBasicAuth("", *c.accessToken)
		c.restClientDev.SetRetryCount(c.RequestRetries)
		c.restClientDev.OnBeforeRequest(c.restOnBeforeRequest)
		c.restClientDev.OnAfterResponse(c.restOnAfterResponse)
	}

	return c.restClientDev
}

func (c *AzureDevopsClient) restVsrm() *resty.Client {
	if c.restClientVsrm == nil {
		c.restClientVsrm = resty.New()
		c.restClientVsrm.SetHostURL(fmt.Sprintf("https://vsrm.dev.azure.com/%v/", *c.organization))
		c.restClientVsrm.SetHeader("Accept", "application/json")
		c.restClientVsrm.SetBasicAuth("", *c.accessToken)
		c.restClientVsrm.SetRetryCount(c.RequestRetries)
		c.restClientVsrm.OnBeforeRequest(c.restOnBeforeRequest)
		c.restClientVsrm.OnAfterResponse(c.restOnAfterResponse)
	}

	return c.restClientVsrm
}

func (c *AzureDevopsClient) concurrencyLock() {
	c.semaphore <- true
}

func (c *AzureDevopsClient) concurrencyUnlock() {
	<-c.semaphore
}

func (c *AzureDevopsClient) restOnBeforeRequest(client *resty.Client, request *resty.Request) (err error) {
	atomic.AddUint64(&c.RequestCount, 1)
	return
}

func (c *AzureDevopsClient) restOnAfterResponse(client *resty.Client, response *resty.Response) (err error) {
	return
}

func (c *AzureDevopsClient) GetRequestCount() float64 {
	requestCount := atomic.LoadUint64(&c.RequestCount)
	return float64(requestCount)
}

func (c *AzureDevopsClient) GetCurrentConcurrency() float64 {
	return float64(len(c.semaphore))
}

func (c *AzureDevopsClient) checkResponse(response *resty.Response, err error) (error) {
	if err != nil {
		return err
	}

	if response != nil {
		// check status code
		statusCode := response.StatusCode()
		if statusCode != 200 {
			return errors.New(fmt.Sprintf("Response status code is %v (expected 200)", statusCode))
		}
	} else {
		return errors.New("Response is nil")
	}

	return nil
}
