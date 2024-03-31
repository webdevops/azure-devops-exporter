package AzureDevopsClient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	resty "github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	AZURE_DEVOPS_SCOPE = "499b84ac-1321-427f-aa17-267ca6975798/.default"
)

type AzureDevopsClient struct {
	logger *zap.SugaredLogger

	// RequestCount has to be the first words
	// in order to be 64-aligned on 32-bit architectures.
	RequestCount   uint64
	RequestRetries int

	organization *string
	collection   *string

	// we can either use a PAT token for authentication
	accessToken *string

	// azure auth
	azcreds *azidentity.DefaultAzureCredential

	HostUrl *string

	ApiVersion string

	restClient     *resty.Client
	restClientVsrm *resty.Client

	semaphore   chan bool
	concurrency int64

	delayUntil *time.Time

	LimitProject                      int64
	LimitBuildsPerProject             int64
	LimitBuildsPerDefinition          int64
	LimitReleasesPerDefinition        int64
	LimitDeploymentPerDefinition      int64
	LimitReleaseDefinitionsPerProject int64
	LimitReleasesPerProject           int64

	prometheus struct {
		apiRequest *prometheus.HistogramVec
	}
}

type EntraIdToken struct {
	TokenType    *string `json:"token_type"`
	ExpiresIn    *int64  `json:"expires_in"`
	ExtExpiresIn *int64  `json:"ext_expires_in"`
	AccessToken  *string `json:"access_token"`
}

type EntraIdErrorResponse struct {
	Error            *string `json:"error"`
	ErrorDescription *string `json:"error_description"`
}

func NewAzureDevopsClient(logger *zap.SugaredLogger) *AzureDevopsClient {
	c := AzureDevopsClient{
		logger: logger,
	}
	c.Init()

	return &c
}

func (c *AzureDevopsClient) Init() {
	collection := "DefaultCollection"
	c.collection = &collection
	c.RequestCount = 0
	c.SetRetries(3)
	c.SetConcurrency(10)

	c.LimitBuildsPerProject = 100
	c.LimitBuildsPerDefinition = 10
	c.LimitReleasesPerDefinition = 100
	c.LimitDeploymentPerDefinition = 100
	c.LimitReleaseDefinitionsPerProject = 100
	c.LimitReleasesPerProject = 100

	c.prometheus.apiRequest = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "azure_devops_api_request",
			Help:    "AzureDevOps API requests",
			Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"endpoint", "organization", "method", "statusCode"},
	)

	prometheus.MustRegister(c.prometheus.apiRequest)
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

	if c.restClientVsrm != nil {
		c.restClientVsrm.SetRetryCount(c.RequestRetries)
	}
}

func (c *AzureDevopsClient) SetUserAgent(v string) {
	c.rest().SetHeader("User-Agent", v)
	c.restVsrm().SetHeader("User-Agent", v)
}

func (c *AzureDevopsClient) SetApiVersion(apiversion string) {
	c.ApiVersion = apiversion
}

func (c *AzureDevopsClient) SetOrganization(url string) {
	c.organization = &url
}

func (c *AzureDevopsClient) SetAccessToken(token string) {
	c.accessToken = &token
}

func (c *AzureDevopsClient) UseAzAuth() error {
	opts := azidentity.DefaultAzureCredentialOptions{}
	cred, err := azidentity.NewDefaultAzureCredential(&opts)
	if err != nil {
		return err
	}

	c.azcreds = cred
	return nil
}

func (c *AzureDevopsClient) SupportsPatAuthentication() bool {
	return c.accessToken != nil && len(*c.accessToken) > 0
}

func (c *AzureDevopsClient) rest() *resty.Client {
	var client, err = c.restWithAuthentication("dev.azure.com")

	if err != nil {
		c.logger.Fatalf("could not create a rest client: %v", err)
	}

	return client
}

func (c *AzureDevopsClient) restVsrm() *resty.Client {
	var client, err = c.restWithAuthentication("vsrm.dev.azure.com")

	if err != nil {
		c.logger.Fatalf("could not create a rest client: %v", err)
	}

	return client
}

func (c *AzureDevopsClient) restWithAuthentication(domain string) (*resty.Client, error) {
	if c.restClient == nil {
		c.restClient = c.restWithoutToken(domain)
	}

	if c.SupportsPatAuthentication() {
		c.restClient.SetBasicAuth("", *c.accessToken)
	} else {
		ctx := context.Background()
		opts := policy.TokenRequestOptions{
			Scopes: []string{AZURE_DEVOPS_SCOPE},
		}
		accessToken, err := c.azcreds.GetToken(ctx, opts)
		if err != nil {
			panic(err)
		}

		c.restClient.SetBasicAuth("", accessToken.Token)
	}

	return c.restClient, nil
}

func (c *AzureDevopsClient) restWithoutToken(domain string) *resty.Client {
	var restClient = resty.New()

	if c.HostUrl != nil {
		restClient.SetBaseURL(*c.HostUrl + "/" + *c.organization + "/")
	} else {
		restClient.SetBaseURL(fmt.Sprintf("https://%v/%v/", domain, *c.organization))
	}

	restClient.SetHeader("Accept", "application/json")
	restClient.SetRetryCount(c.RequestRetries)

	if c.delayUntil != nil {
		restClient.OnBeforeRequest(c.restOnBeforeRequestDelay)
	} else {
		restClient.OnBeforeRequest(c.restOnBeforeRequest)
	}

	restClient.OnAfterResponse(c.restOnAfterResponse)

	return restClient
}

func (c *AzureDevopsClient) concurrencyLock() {
	c.semaphore <- true
}

func (c *AzureDevopsClient) concurrencyUnlock() {
	<-c.semaphore
}

// PreRequestHook is a resty hook that is called before every request
// It checks that the delay is ok before requesting
func (c *AzureDevopsClient) restOnBeforeRequestDelay(client *resty.Client, request *resty.Request) (err error) {
	atomic.AddUint64(&c.RequestCount, 1)
	if c.delayUntil != nil {
		if time.Now().Before(*c.delayUntil) {
			time.Sleep(time.Until(*c.delayUntil))
		}
		c.delayUntil = nil
	}
	return
}

func (c *AzureDevopsClient) restOnBeforeRequest(client *resty.Client, request *resty.Request) (err error) {
	atomic.AddUint64(&c.RequestCount, 1)
	return
}

func (c *AzureDevopsClient) restOnAfterResponse(client *resty.Client, response *resty.Response) (err error) {
	requestUrl, _ := url.Parse(response.Request.URL)
	c.prometheus.apiRequest.With(prometheus.Labels{
		"endpoint":     requestUrl.Hostname(),
		"organization": *c.organization,
		"method":       strings.ToLower(response.Request.Method),
		"statusCode":   strconv.FormatInt(int64(response.StatusCode()), 10),
	}).Observe(response.Time().Seconds())
	return
}

func (c *AzureDevopsClient) GetRequestCount() float64 {
	requestCount := atomic.LoadUint64(&c.RequestCount)
	return float64(requestCount)
}

func (c *AzureDevopsClient) GetCurrentConcurrency() float64 {
	return float64(len(c.semaphore))
}

func (c *AzureDevopsClient) checkResponse(response *resty.Response, err error) error {
	if err != nil {
		return err
	}
	if response != nil {
		// check delay from usage quota
		if d := response.Header().Get("Retry-After"); d != "" {
			// convert string to int to time.Duration
			if dInt, err := strconv.Atoi(d); err != nil {
				dD := time.Now().Add(time.Duration(dInt) * time.Second)
				c.delayUntil = &dD
			}
		}
		// check status code
		statusCode := response.StatusCode()
		if statusCode != 200 {
			return fmt.Errorf("response status code is %v (expected 200), url: %v", statusCode, response.Request.URL)
		}
	} else {
		return errors.New("response is nil")
	}

	return nil
}
