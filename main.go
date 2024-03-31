package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"

	AzureDevops "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	"github.com/webdevops/azure-devops-exporter/config"
)

const (
	Author = "webdevops.io"

	cacheTag = "v1"
)

var (
	argparser *flags.Parser
	opts      config.Opts

	AzureDevopsClient           *AzureDevops.AzureDevopsClient
	AzureDevopsServiceDiscovery *azureDevopsServiceDiscovery

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

func main() {
	initLogger()
	initArgparser()

	logger.Infof("starting azure-devops-exporter v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	logger.Info(string(opts.GetJson()))

	logger.Infof("init AzureDevOps connection")
	initAzureDevOpsConnection()
	AzureDevopsServiceDiscovery = NewAzureDevopsServiceDiscovery()
	AzureDevopsServiceDiscovery.Update()

	logger.Info("init metrics collection")
	initMetricCollector()

	logger.Infof("starting http server on %s", opts.Server.Bind)
	startHttpServer()
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	// load accesstoken from file
	if opts.AzureDevops.AccessTokenFile != nil && len(*opts.AzureDevops.AccessTokenFile) > 0 {
		logger.Infof("reading access token from file \"%s\"", *opts.AzureDevops.AccessTokenFile)
		// load access token from file
		if val, err := os.ReadFile(*opts.AzureDevops.AccessTokenFile); err == nil {
			opts.AzureDevops.AccessToken = strings.TrimSpace(string(val))
		} else {
			logger.Fatalf("unable to read access token file \"%s\": %v", *opts.AzureDevops.AccessTokenFile, err)
		}
	}

	if len(opts.AzureDevops.AccessToken) == 0 && (len(opts.AzureDevops.TenantId) == 0 || len(opts.AzureDevops.ClientId) == 0) {
		logger.Fatalf("neither an Azure DevOps PAT token nor client credentials (tenant ID, client ID) for service principal authentication have been provided")
	}

	// ensure query paths and projects are splitted by '@'
	if opts.AzureDevops.QueriesWithProjects != nil {
		queryError := false
		for _, query := range opts.AzureDevops.QueriesWithProjects {
			if strings.Count(query, "@") != 1 {
				fmt.Println("Query path '", query, "' is malformed; should be '<query UUID>@<project UUID>'")
				queryError = true
			}
		}
		if queryError {
			os.Exit(1)
		}
	}

	// use default scrape time if null
	if opts.Scrape.TimeProjects == nil {
		opts.Scrape.TimeProjects = &opts.Scrape.Time
	}

	if opts.Scrape.TimeRepository == nil {
		opts.Scrape.TimeRepository = &opts.Scrape.Time
	}

	if opts.Scrape.TimePullRequest == nil {
		opts.Scrape.TimePullRequest = &opts.Scrape.Time
	}

	if opts.Scrape.TimeBuild == nil {
		opts.Scrape.TimeBuild = &opts.Scrape.Time
	}

	if opts.Scrape.TimeRelease == nil {
		opts.Scrape.TimeRelease = &opts.Scrape.Time
	}

	if opts.Scrape.TimeDeployment == nil {
		opts.Scrape.TimeDeployment = &opts.Scrape.Time
	}

	if opts.Scrape.TimeStats == nil {
		opts.Scrape.TimeStats = &opts.Scrape.Time
	}

	if opts.Scrape.TimeResourceUsage == nil {
		opts.Scrape.TimeResourceUsage = &opts.Scrape.Time
	}

	if opts.Stats.SummaryMaxAge == nil {
		opts.Stats.SummaryMaxAge = opts.Scrape.TimeStats
	}

	if opts.Scrape.TimeQuery == nil {
		opts.Scrape.TimeQuery = &opts.Scrape.Time
	}

	if v := os.Getenv("AZURE_DEVOPS_FILTER_AGENTPOOL"); v != "" {
		logger.Fatal("deprecated env var AZURE_DEVOPS_FILTER_AGENTPOOL detected, please use AZURE_DEVOPS_AGENTPOOL")
	}
}

// Init and build Azure authorzier
func initAzureDevOpsConnection() {
	AzureDevopsClient = AzureDevops.NewAzureDevopsClient(logger)
	if opts.AzureDevops.Url != nil {
		AzureDevopsClient.HostUrl = opts.AzureDevops.Url
	}

	logger.Infof("using organization: %v", opts.AzureDevops.Organisation)
	logger.Infof("using apiversion: %v", opts.AzureDevops.ApiVersion)
	logger.Infof("using concurrency: %v", opts.Request.ConcurrencyLimit)
	logger.Infof("using retries: %v", opts.Request.Retries)

	if opts.AzureDevops.TenantId != "" {
		if err := os.Setenv("AZURE_TENANT_ID", opts.AzureDevops.TenantId); err != nil {
			panic(err)
		}
	}

	if opts.AzureDevops.ClientId != "" {
		if err := os.Setenv("AZURE_CLIENT_ID", opts.AzureDevops.ClientId); err != nil {
			panic(err)
		}
	}

	if opts.AzureDevops.ClientSecret != "" {
		if err := os.Setenv("AZURE_CLIENT_SECRET", opts.AzureDevops.ClientSecret); err != nil {
			panic(err)
		}
	}

	AzureDevopsClient.SetOrganization(opts.AzureDevops.Organisation)
	if opts.AzureDevops.AccessToken != "" {
		AzureDevopsClient.SetAccessToken(opts.AzureDevops.AccessToken)
	} else {
		if err := AzureDevopsClient.UseAzAuth(); err != nil {
			logger.Fatalf(err.Error())
		}
	}
	AzureDevopsClient.SetApiVersion(opts.AzureDevops.ApiVersion)
	AzureDevopsClient.SetConcurrency(opts.Request.ConcurrencyLimit)
	AzureDevopsClient.SetRetries(opts.Request.Retries)
	AzureDevopsClient.SetUserAgent(fmt.Sprintf("azure-devops-exporter/%v", gitTag))

	AzureDevopsClient.LimitProject = opts.Limit.Project
	AzureDevopsClient.LimitBuildsPerProject = opts.Limit.BuildsPerProject
	AzureDevopsClient.LimitBuildsPerDefinition = opts.Limit.BuildsPerDefinition
	AzureDevopsClient.LimitReleasesPerDefinition = opts.Limit.ReleasesPerDefinition
	AzureDevopsClient.LimitDeploymentPerDefinition = opts.Limit.DeploymentPerDefinition
	AzureDevopsClient.LimitReleaseDefinitionsPerProject = opts.Limit.ReleaseDefinitionsPerProject
	AzureDevopsClient.LimitReleasesPerProject = opts.Limit.ReleasesPerProject
}

func initMetricCollector() {
	var collectorName string

	collectorName = "Project"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorProject{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeLive)
		c.SetCache(opts.GetCachePath("project.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "AgentPool"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorAgentPool{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeLive)
		c.SetCache(opts.GetCachePath("agentpool.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "LatestBuild"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorLatestBuild{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeLive)
		c.SetCache(opts.GetCachePath("latestbuild.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Repository"
	if opts.Scrape.TimeRepository.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorRepository{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeRepository)
		c.SetCache(opts.GetCachePath("latestbuild.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "PullRequest"
	if opts.Scrape.TimePullRequest.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorPullRequest{}, logger)
		c.SetScapeTime(*opts.Scrape.TimePullRequest)
		c.SetCache(opts.GetCachePath("pullrequest.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Build"
	if opts.Scrape.TimeBuild.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorBuild{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeBuild)
		c.SetCache(opts.GetCachePath("build.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Release"
	if opts.Scrape.TimeRelease.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorRelease{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeRelease)
		c.SetCache(opts.GetCachePath("release.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Deployment"
	if opts.Scrape.TimeDeployment.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorDeployment{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeDeployment)
		c.SetCache(opts.GetCachePath("deployment.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Stats"
	if opts.Scrape.TimeStats.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorStats{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeStats)
		c.SetCache(opts.GetCachePath("stats.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "ResourceUsage"
	if opts.Scrape.TimeResourceUsage.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorResourceUsage{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeResourceUsage)
		c.SetCache(opts.GetCachePath("resourceusage.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Query"
	if opts.Scrape.TimeQuery.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorQuery{}, logger)
		c.SetScapeTime(*opts.Scrape.TimeQuery)
		c.SetCache(opts.GetCachePath("query.json"), collector.BuildCacheTag(cacheTag, opts.AzureDevops))
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(zap.String("collector", collectorName)).Info("collector disabled")
	}
}

// start and handle prometheus handler
func startHttpServer() {
	mux := http.NewServeMux()

	// healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err)
		}
	})

	// readyz
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err)
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         opts.Server.Bind,
		Handler:      mux,
		ReadTimeout:  opts.Server.ReadTimeout,
		WriteTimeout: opts.Server.WriteTimeout,
	}
	logger.Fatal(srv.ListenAndServe())
}
