package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	AzureDevops "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	"github.com/webdevops/azure-devops-exporter/config"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	Author = "webdevops.io"
)

var (
	argparser *flags.Parser
	opts      config.Opts

	AzureDevopsClient *AzureDevops.AzureDevopsClient

	collectorGeneralList   map[string]*CollectorGeneral
	collectorProjectList   map[string]*CollectorProject
	collectorAgentPoolList map[string]*CollectorAgentPool
	collectorQueryList     map[string]*CollectorQuery

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

func main() {
	initArgparser()

	log.Infof("starting azure-devops-exporter v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	log.Info(string(opts.GetJson()))

	log.Infof("init AzureDevOps connection")
	initAzureDevOpsConnection()

	log.Info("init metrics collection")
	initMetricCollector()

	log.Infof("starting http server on %s", opts.ServerBind)
	startHttpServer()
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	// verbose level
	if opts.Logger.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// debug level
	if opts.Logger.Debug {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
		log.SetFormatter(&log.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	}

	// json log format
	if opts.Logger.LogJson {
		log.SetReportCaller(true)
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
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
}

// Init and build Azure authorzier
func initAzureDevOpsConnection() {
	AzureDevopsClient = AzureDevops.NewAzureDevopsClient()
	if opts.AzureDevops.Url != nil {
		AzureDevopsClient.HostUrl = opts.AzureDevops.Url
	}

	log.Infof("using organization: %v", opts.AzureDevops.Organisation)
	log.Infof("using apiversion: %v", opts.AzureDevops.ApiVersion)
	log.Infof("using concurrency: %v", opts.Request.ConcurrencyLimit)
	log.Infof("using retries: %v", opts.Request.Retries)

	AzureDevopsClient.SetOrganization(opts.AzureDevops.Organisation)
	AzureDevopsClient.SetAccessToken(opts.AzureDevops.AccessToken)
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

func getAzureDevOpsProjects() (list AzureDevops.ProjectList) {
	rawList, err := AzureDevopsClient.ListProjects()

	if err != nil {
		panic(err)
	}

	list = rawList

	// whitelist
	if len(opts.AzureDevops.FilterProjects) > 0 {
		rawList = list
		list = AzureDevops.ProjectList{}
		for _, project := range rawList.List {
			if arrayStringContains(opts.AzureDevops.FilterProjects, project.Id) {
				list.List = append(list.List, project)
			}
		}
	}

	// blacklist
	if len(opts.AzureDevops.BlacklistProjects) > 0 {
		// filter ignored azure devops projects
		rawList = list
		list = AzureDevops.ProjectList{}
		for _, project := range rawList.List {
			if !arrayStringContains(opts.AzureDevops.BlacklistProjects, project.Id) {
				list.List = append(list.List, project)
			}
		}
	}

	return
}

func initMetricCollector() {
	var collectorName string
	collectorGeneralList = map[string]*CollectorGeneral{}
	collectorProjectList = map[string]*CollectorProject{}
	collectorAgentPoolList = map[string]*CollectorAgentPool{}
	collectorQueryList = map[string]*CollectorQuery{}

	projectList := getAzureDevOpsProjects()

	collectorName = "General"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		collectorGeneralList[collectorName] = NewCollectorGeneral(collectorName, &MetricsCollectorGeneral{})
		collectorGeneralList[collectorName].SetAzureProjects(&projectList)
		collectorGeneralList[collectorName].SetScrapeTime(*opts.Scrape.TimeLive)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Project"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorProject{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeLive)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "AgentPool"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		collectorAgentPoolList[collectorName] = NewCollectorAgentPool(collectorName, &MetricsCollectorAgentPool{})
		collectorAgentPoolList[collectorName].SetAzureProjects(&projectList)
		collectorAgentPoolList[collectorName].AgentPoolIdList = opts.AzureDevops.FilterAgentPoolId
		collectorAgentPoolList[collectorName].SetScrapeTime(*opts.Scrape.TimeLive)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "LatestBuild"
	if opts.Scrape.TimeLive.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorLatestBuild{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeLive)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Repository"
	if opts.Scrape.TimeRepository.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorRepository{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeRepository)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "PullRequest"
	if opts.Scrape.TimePullRequest.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorPullRequest{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimePullRequest)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Build"
	if opts.Scrape.TimeBuild.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorBuild{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeBuild)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Release"
	if opts.Scrape.TimeRelease.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorRelease{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeRelease)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Deployment"
	if opts.Scrape.TimeDeployment.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorDeployment{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeDeployment)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Stats"
	if opts.Scrape.TimeStats.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorStats{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.Scrape.TimeStats)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "ResourceUsage"
	if opts.Scrape.TimeResourceUsage.Seconds() > 0 {
		collectorGeneralList[collectorName] = NewCollectorGeneral(collectorName, &MetricsCollectorResourceUsage{})
		collectorGeneralList[collectorName].SetAzureProjects(&projectList)
		collectorGeneralList[collectorName].SetScrapeTime(*opts.Scrape.TimeResourceUsage)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Query"
	if opts.Scrape.TimeQuery.Seconds() > 0 {
		collectorQueryList[collectorName] = NewCollectorQuery(collectorName, &MetricsCollectorQuery{})
		collectorQueryList[collectorName].SetAzureProjects(&projectList)
		collectorQueryList[collectorName].QueryList = opts.AzureDevops.QueriesWithProjects
		collectorQueryList[collectorName].SetScrapeTime(*opts.Scrape.TimeQuery)
	} else {
		log.Infof("collector[%s]: disabled", collectorName)
	}

	for _, collector := range collectorGeneralList {
		collector.Run()
	}

	for _, collector := range collectorProjectList {
		collector.Run()
	}

	for _, collector := range collectorAgentPoolList {
		collector.Run()
	}

	for _, collector := range collectorQueryList {
		collector.Run()
	}

	// background auto update of projects
	if opts.Scrape.TimeProjects.Seconds() > 0 {
		go func() {
			// initial sleep
			time.Sleep(*opts.Scrape.TimeProjects)

			for {
				log.Info("daemon: updating project list")

				projectList := getAzureDevOpsProjects()
				log.Infof("daemon: found %v projects", projectList.Count)

				for _, collector := range collectorGeneralList {
					collector.SetAzureProjects(&projectList)
				}

				for _, collector := range collectorProjectList {
					collector.SetAzureProjects(&projectList)
				}

				for _, collector := range collectorAgentPoolList {
					collector.SetAzureProjects(&projectList)
				}

				log.Info("daemon: skipping Query collectors; they don't use projects")
				time.Sleep(*opts.Scrape.TimeProjects)
			}
		}()
	}
}

// start and handle prometheus handler
func startHttpServer() {
	// healthz
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			log.Error(err)
		}
	})

	http.Handle("/metrics", promhttp.Handler())
	log.Error(http.ListenAndServe(opts.ServerBind, nil))
}
