package AzureDevopsClient

import (
	"fmt"
	"github.com/go-resty/resty"
)

type AzureDevopsClient struct {
	organization *string
	collection *string
	accessToken *string

	restClient *resty.Client
	restClientDev *resty.Client
	restClientVsrm *resty.Client
}

func NewAzureDevopsClient() *AzureDevopsClient {
	c := AzureDevopsClient{}
	c.Init()

	return &c
}

func (c *AzureDevopsClient) Init() {
	collection := "DefaultCollection"
	c.collection = &collection
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
	}

	return c.restClient
}


func (c *AzureDevopsClient) restDev() *resty.Client {
	if c.restClientDev == nil {
		c.restClientDev = resty.New()
		c.restClientDev.SetHostURL(fmt.Sprintf("https://dev.azure.com/%v/", *c.organization))
		c.restClientDev.SetHeader("Accept", "application/json")
		c.restClientDev.SetBasicAuth("", *c.accessToken)
	}

	return c.restClientDev
}

func (c *AzureDevopsClient) restVsrm() *resty.Client {
	if c.restClientVsrm == nil {
		c.restClientVsrm = resty.New()
		c.restClientVsrm.SetHostURL(fmt.Sprintf("https://vsrm.dev.azure.com/%v/", *c.organization))
		c.restClientVsrm.SetHeader("Accept", "application/json")
		c.restClientVsrm.SetBasicAuth("", *c.accessToken)
	}

	return c.restClientVsrm
}
