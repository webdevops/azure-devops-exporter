package AzureDevopsClient

import (
	"fmt"
	"sync/atomic"
	"github.com/go-resty/resty"
)

type AzureDevopsClient struct {
	organization *string
	collection *string
	accessToken *string

	restClient *resty.Client
	restClientDev *resty.Client
	restClientVsrm *resty.Client

	semaphore chan bool
	concurrency int64

	RequestCount uint64
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
	c.SetConcurrency(10)
}

func (c *AzureDevopsClient) SetConcurrency(concurrency int64) {
	c.concurrency = concurrency
	c.semaphore = make(chan bool, c.concurrency)
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
		c.restClient.SetRetryCount(3)
		c.restClient.OnBeforeRequest(c.restOnBeforeRequest);
		c.restClient.OnAfterResponse(c.restOnAfterResponse);
	}

	return c.restClient
}

func (c *AzureDevopsClient) restDev() *resty.Client {
	if c.restClientDev == nil {
		c.restClientDev = resty.New()
		c.restClientDev.SetHostURL(fmt.Sprintf("https://dev.azure.com/%v/", *c.organization))
		c.restClientDev.SetHeader("Accept", "application/json")
		c.restClientDev.SetBasicAuth("", *c.accessToken)
		c.restClientDev.SetRetryCount(3)
		c.restClientDev.OnBeforeRequest(c.restOnBeforeRequest);
		c.restClientDev.OnAfterResponse(c.restOnAfterResponse);
	}

	return c.restClientDev
}

func (c *AzureDevopsClient) restVsrm() *resty.Client {
	if c.restClientVsrm == nil {
		c.restClientVsrm = resty.New()
		c.restClientVsrm.SetHostURL(fmt.Sprintf("https://vsrm.dev.azure.com/%v/", *c.organization))
		c.restClientVsrm.SetHeader("Accept", "application/json")
		c.restClientVsrm.SetBasicAuth("", *c.accessToken)
		c.restClientVsrm.SetRetryCount(3)
		c.restClientVsrm.OnBeforeRequest(c.restOnBeforeRequest)
		c.restClientVsrm.OnAfterResponse(c.restOnAfterResponse)
	}

	return c.restClientVsrm
}

func (c *AzureDevopsClient) restOnBeforeRequest(client *resty.Client, request *resty.Request) (err error) {
	c.semaphore <- true
	atomic.AddUint64(&c.RequestCount, 1)
	return
}

func (c *AzureDevopsClient) restOnAfterResponse(client *resty.Client, response *resty.Response) (err error) {
	<-c.semaphore
	return
}

func (c *AzureDevopsClient) GetRequestCount() float64 {
	requestCount := atomic.LoadUint64(&c.RequestCount)
	return float64(requestCount)
}
