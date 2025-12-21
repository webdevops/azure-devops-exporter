package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/webdevops/go-common/prometheus/collector"

	AzureDevops "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	"github.com/webdevops/azure-devops-exporter/config"
)

const (
	Author = "webdevops.io"

	cacheTag = "v1"
)

var (
	argparser *flags.Parser
	Opts      config.Opts

	AzureDevopsClient           *AzureDevops.AzureDevopsClient
	AzureDevopsServiceDiscovery *azureDevopsServiceDiscovery

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
	buildDate = "<unknown>"
)

func main() {
	initArgparser()
	initLogger()
	parseArguments()

	logger.Infof("starting azure-devops-exporter v%s (%s; %s; by %v at %v)", gitTag, gitCommit, runtime.Version(), Author, buildDate)
	logger.Info(string(Opts.GetJson()))
	initSystem()

	logger.Infof("init AzureDevOps connection")
	initAzureDevOpsConnection()
	AzureDevopsServiceDiscovery = NewAzureDevopsServiceDiscovery()
	AzureDevopsServiceDiscovery.Update()

	logger.Info("init metrics collection")
	initMetricCollector()

	logger.Infof("starting http server on %s", Opts.Server.Bind)
	startHttpServer()
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&Opts, flags.Default)
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
}

// parses and validates the arguments
func parseArguments() {
	// load accesstoken from file
	if Opts.AzureDevops.AccessTokenFile != nil && len(*Opts.AzureDevops.AccessTokenFile) > 0 {
		logger.Infof("reading access token from file \"%s\"", *Opts.AzureDevops.AccessTokenFile)
		// load access token from file
		if val, err := os.ReadFile(*Opts.AzureDevops.AccessTokenFile); err == nil {
			Opts.AzureDevops.AccessToken = strings.TrimSpace(string(val))
		} else {
			logger.Fatalf("unable to read access token file \"%s\": %v", *Opts.AzureDevops.AccessTokenFile, err)
		}
	}

	if len(Opts.AzureDevops.AccessToken) == 0 && (len(Opts.Azure.TenantId) == 0 || len(Opts.Azure.ClientId) == 0) {
		logger.Fatalf("neither an Azure DevOps PAT token nor client credentials (tenant ID, client ID) for service principal authentication have been provided")
	}

	// ensure query paths and projects are splitted by '@'
	if Opts.AzureDevops.QueriesWithProjects != nil {
		queryError := false
		for _, query := range Opts.AzureDevops.QueriesWithProjects {
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
	if Opts.Scrape.TimeProjects == nil {
		Opts.Scrape.TimeProjects = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimeRepository == nil {
		Opts.Scrape.TimeRepository = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimePullRequest == nil {
		Opts.Scrape.TimePullRequest = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimeBuild == nil {
		Opts.Scrape.TimeBuild = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimeRelease == nil {
		Opts.Scrape.TimeRelease = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimeDeployment == nil {
		Opts.Scrape.TimeDeployment = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimeStats == nil {
		Opts.Scrape.TimeStats = &Opts.Scrape.Time
	}

	if Opts.Scrape.TimeResourceUsage == nil {
		Opts.Scrape.TimeResourceUsage = &Opts.Scrape.Time
	}

	if Opts.Stats.SummaryMaxAge == nil {
		Opts.Stats.SummaryMaxAge = Opts.Scrape.TimeStats
	}

	if Opts.Scrape.TimeQuery == nil {
		Opts.Scrape.TimeQuery = &Opts.Scrape.Time
	}

	if v := os.Getenv("AZURE_DEVOPS_FILTER_AGENTPOOL"); v != "" {
		logger.Fatal("deprecated env var AZURE_DEVOPS_FILTER_AGENTPOOL detected, please use AZURE_DEVOPS_AGENTPOOL")
	}
}

// Init and build Azure authorzier
func initAzureDevOpsConnection() {
	AzureDevopsClient = AzureDevops.NewAzureDevopsClient(logger)
	if Opts.AzureDevops.Url != nil {
		AzureDevopsClient.HostUrl = Opts.AzureDevops.Url
	}

	logger.Infof("using organization: %v", Opts.AzureDevops.Organisation)
	logger.Infof("using apiversion: %v", Opts.AzureDevops.ApiVersion)
	logger.Infof("using concurrency: %v", Opts.Request.ConcurrencyLimit)
	logger.Infof("using retries: %v", Opts.Request.Retries)

	// ensure AZURE env vars are populated for azidentity
	if Opts.Azure.TenantId != "" {
		if err := os.Setenv("AZURE_TENANT_ID", Opts.Azure.TenantId); err != nil {
			panic(err)
		}
	}

	if Opts.Azure.ClientId != "" {
		if err := os.Setenv("AZURE_CLIENT_ID", Opts.Azure.ClientId); err != nil {
			panic(err)
		}
	}

	if Opts.Azure.ClientSecret != "" {
		if err := os.Setenv("AZURE_CLIENT_SECRET", Opts.Azure.ClientSecret); err != nil {
			panic(err)
		}
	}

	AzureDevopsClient.SetOrganization(Opts.AzureDevops.Organisation)
	if Opts.AzureDevops.AccessToken != "" {
		AzureDevopsClient.SetAccessToken(Opts.AzureDevops.AccessToken)
	} else {
		if err := AzureDevopsClient.UseAzAuth(); err != nil {
			logger.Fatal(err.Error())
		}
	}
	AzureDevopsClient.SetApiVersion(Opts.AzureDevops.ApiVersion)
	AzureDevopsClient.SetConcurrency(Opts.Request.ConcurrencyLimit)
	AzureDevopsClient.SetRetries(Opts.Request.Retries)
	AzureDevopsClient.SetUserAgent(fmt.Sprintf("azure-devops-exporter/%v", gitTag))

	AzureDevopsClient.LimitProject = Opts.Limit.Project
	AzureDevopsClient.LimitBuildsPerProject = Opts.Limit.BuildsPerProject
	AzureDevopsClient.LimitBuildsPerDefinition = Opts.Limit.BuildsPerDefinition
	AzureDevopsClient.LimitReleasesPerDefinition = Opts.Limit.ReleasesPerDefinition
	AzureDevopsClient.LimitDeploymentPerDefinition = Opts.Limit.DeploymentPerDefinition
	AzureDevopsClient.LimitReleaseDefinitionsPerProject = Opts.Limit.ReleaseDefinitionsPerProject
	AzureDevopsClient.LimitReleasesPerProject = Opts.Limit.ReleasesPerProject
}

func initMetricCollector() {
	var collectorName string

	collectorName = "Project"
	if Opts.Scrape.TimeLive.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorProject{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeLive)
		if err := c.SetCache(Opts.GetCachePath("project.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "AgentPool"
	if Opts.Scrape.TimeLive.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorAgentPool{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeLive)
		if err := c.SetCache(Opts.GetCachePath("agentpool.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "LatestBuild"
	if Opts.Scrape.TimeLive.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorLatestBuild{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeLive)
		if err := c.SetCache(Opts.GetCachePath("latestbuild.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Repository"
	if Opts.Scrape.TimeRepository.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorRepository{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeRepository)
		if err := c.SetCache(Opts.GetCachePath("latestbuild.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "PullRequest"
	if Opts.Scrape.TimePullRequest.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorPullRequest{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimePullRequest)
		if err := c.SetCache(Opts.GetCachePath("pullrequest.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Build"
	if Opts.Scrape.TimeBuild.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorBuild{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeBuild)
		if err := c.SetCache(Opts.GetCachePath("build.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Release"
	if Opts.Scrape.TimeRelease.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorRelease{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeRelease)
		if err := c.SetCache(Opts.GetCachePath("release.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Deployment"
	if Opts.Scrape.TimeDeployment.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorDeployment{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeDeployment)
		if err := c.SetCache(Opts.GetCachePath("deployment.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Stats"
	if Opts.Scrape.TimeStats.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorStats{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeStats)
		if err := c.SetCache(Opts.GetCachePath("stats.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "ResourceUsage"
	if Opts.Scrape.TimeResourceUsage.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorResourceUsage{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeResourceUsage)
		if err := c.SetCache(Opts.GetCachePath("resourceusage.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}

	collectorName = "Query"
	if Opts.Scrape.TimeQuery.Seconds() > 0 {
		c := collector.New(collectorName, &MetricsCollectorQuery{}, logger.Slog())
		c.SetScapeTime(*Opts.Scrape.TimeQuery)
		if err := c.SetCache(Opts.GetCachePath("query.json"), collector.BuildCacheTag(cacheTag, Opts.AzureDevops)); err != nil {
			logger.Fatal(err.Error())
		}
		if err := c.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	} else {
		logger.With(slog.String("collector", collectorName)).Info("collector disabled")
	}
}

// start and handle prometheus handler
func startHttpServer() {
	mux := http.NewServeMux()

	// healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err.Error())
		}
	})

	// readyz
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err.Error())
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         Opts.Server.Bind,
		Handler:      mux,
		ReadTimeout:  Opts.Server.ReadTimeout,
		WriteTimeout: Opts.Server.WriteTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal(err.Error())
	}
}
