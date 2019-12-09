package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	AzureDevops "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	Author  = "webdevops.io"
	Version = "0.8.0-beta10"
)

var (
	argparser         *flags.Parser
	args              []string
	Verbose           bool
	Logger            *DaemonLogger
	AzureDevopsClient *AzureDevops.AzureDevopsClient

	collectorGeneralList   map[string]*CollectorGeneral
	collectorProjectList   map[string]*CollectorProject
	collectorAgentPoolList map[string]*CollectorAgentPool
	collectorQueryList     map[string]*CollectorQuery
)

var opts struct {
	// general settings
	Verbose []bool `   long:"verbose" short:"v"                   env:"VERBOSE"                       description:"Verbose mode"`

	// server settings
	ServerBind string `long:"bind"                                env:"SERVER_BIND"                   description:"Server address"                                    default:":8080"`

	// scrape time settings
	ScrapeTime              time.Duration  `long:"scrape.time"                  env:"SCRAPE_TIME"                    description:"Default scrape time (time.duration)"                       default:"30m"`
	ScrapeTimeProjects      *time.Duration `long:"scrape.time.projects"         env:"SCRAPE_TIME_PROJECTS"           description:"Scrape time for project metrics (time.duration)"`
	ScrapeTimeRepository    *time.Duration `long:"scrape.time.repository"       env:"SCRAPE_TIME_REPOSITORY"         description:"Scrape time for repository metrics (time.duration)"`
	ScrapeTimeBuild         *time.Duration `long:"scrape.time.build"            env:"SCRAPE_TIME_BUILD"              description:"Scrape time for build metrics (time.duration)"`
	ScrapeTimeRelease       *time.Duration `long:"scrape.time.release"          env:"SCRAPE_TIME_RELEASE"            description:"Scrape time for release metrics (time.duration)"`
	ScrapeTimeDeployment    *time.Duration `long:"scrape.time.deployment"       env:"SCRAPE_TIME_DEPLOYMENT"         description:"Scrape time for deployment metrics (time.duration)"`
	ScrapeTimePullRequest   *time.Duration `long:"scrape.time.pullrequest"      env:"SCRAPE_TIME_PULLREQUEST"        description:"Scrape time for pullrequest metrics  (time.duration)"`
	ScrapeTimeStats         *time.Duration `long:"scrape.time.stats"            env:"SCRAPE_TIME_STATS"              description:"Scrape time for stats metrics  (time.duration)"`
	ScrapeTimeResourceUsage *time.Duration `long:"scrape.time.resourceusage"    env:"SCRAPE_TIME_RESOURCEUSAGE"      description:"Scrape time for resourceusage metrics  (time.duration)"`
	ScrapeTimeQuery         *time.Duration `long:"scrape.time.query"            env:"SCRAPE_TIME_QUERY"              description:"Scrape time for query results  (time.duration)"`
	ScrapeTimeLive          *time.Duration `long:"scrape.time.live"             env:"SCRAPE_TIME_LIVE"               description:"Scrape time for live metrics (time.duration)"              default:"30s"`

	// summary options
	StatsSummaryMaxAge *time.Duration `long:"stats.summary.maxage"         env:"STATS_SUMMARY_MAX_AGE"             description:"Stats Summary metrics max age (time.duration)"`

	// ignore settings
	AzureDevopsFilterProjects    []string `long:"whitelist.project"    env:"AZURE_DEVOPS_FILTER_PROJECT"    env-delim:" "   description:"Filter projects (UUIDs)"`
	AzureDevopsBlacklistProjects []string `long:"blacklist.project"    env:"AZURE_DEVOPS_BLACKLIST_PROJECT" env-delim:" "   description:"Filter projects (UUIDs)"`
	AzureDevopsFilterAgentPoolId []int64  `long:"whitelist.agentpool"  env:"AZURE_DEVOPS_FILTER_AGENTPOOL"  env-delim:" "   description:"Filter of agent pool (IDs)"`

	// query settings
	QueriesWithProjects []string `long:"list.query"    env:"AZURE_DEVOPS_QUERIES"    env-delim:" "   description:"Pairs of query and project UUIDs in the form: '<queryId>@<projectId>'"`

	// azure settings
	AzureDevopsUrl          *string `long:"azuredevops.url"                     env:"AZURE_DEVOPS_URL"             description:"Azure DevOps url (empty if hosted by microsoft)"`
	AzureDevopsAccessToken  string  `long:"azuredevops.access-token"            env:"AZURE_DEVOPS_ACCESS_TOKEN"    description:"Azure DevOps access token" required:"true"`
	AzureDevopsOrganisation string  `long:"azuredevops.organisation"            env:"AZURE_DEVOPS_ORGANISATION"    description:"Azure DevOps organization" required:"true"`
	AzureDevopsApiVersion   string  `long:"azuredevops.apiversion"              env:"AZURE_DEVOPS_APIVERSION"      description:"Azure DevOps API version"  default:"5.0"`

	RequestConcurrencyLimit int64 `long:"request.concurrency"                   env:"REQUEST_CONCURRENCY"     description:"Number of concurrent requests against dev.azure.com"  default:"10"`
	RequestRetries          int   `long:"request.retries"                       env:"REQUEST_RETRIES"         description:"Number of retried requests against dev.azure.com"     default:"3"`

	LimitBuildsPerDefinition          int64 `long:"limit.builds-per-definition"           env:"LIMIT_BUILDS_PER_DEFINITION"           description:"Limit builds per definition"      default:"10"`
	LimitReleasesPerDefinition        int64 `long:"limit.releases-per-definition"         env:"LIMIT_RELEASES_PER_DEFINITION"         description:"Limit releases per definition"    default:"100"`
	LimitDeploymentPerDefinition      int64 `long:"limit.deployments-per-definition"      env:"LIMIT_DEPLOYMENTS_PER_DEFINITION"      description:"Limit deployments per definition" default:"100"`
	LimitReleaseDefinitionsPerProject int64 `long:"limit.releasedefinitions-per-project"  env:"LIMIT_RELEASEDEFINITION_PER_PROJECT"   description:"Limit builds per definition"      default:"100"`
}

func main() {
	initArgparser()

	// set verbosity
	Verbose = len(opts.Verbose) >= 1

	Logger = NewLogger(log.Lshortfile, Verbose)
	defer Logger.Close()

	Logger.Infof("Init Azure DevOps exporter v%s (written by %v)", Version, Author)

	Logger.Infof("Init AzureDevOps connection")
	initAzureDevOpsConnection()

	Logger.Info("Starting metrics collection")
	Logger.Infof("set scape interval[Default]: %v", scrapeIntervalStatus(&opts.ScrapeTime))
	Logger.Infof("set scape interval[Live]: %v", scrapeIntervalStatus(opts.ScrapeTimeLive))
	Logger.Infof("set scape interval[Project]: %v", scrapeIntervalStatus(opts.ScrapeTimeProjects))
	Logger.Infof("set scape interval[Repository]: %v", scrapeIntervalStatus(opts.ScrapeTimeRepository))
	Logger.Infof("set scape interval[PullRequest]: %v", scrapeIntervalStatus(opts.ScrapeTimePullRequest))
	Logger.Infof("set scape interval[Build]: %v", scrapeIntervalStatus(opts.ScrapeTimeBuild))
	Logger.Infof("set scape interval[Release]: %v", scrapeIntervalStatus(opts.ScrapeTimeRelease))
	Logger.Infof("set scape interval[Deployment]: %v", scrapeIntervalStatus(opts.ScrapeTimeDeployment))
	Logger.Infof("set scape interval[Stats]: %v", scrapeIntervalStatus(opts.ScrapeTimeStats))
	Logger.Infof("set scape interval[ResourceUsage]: %v", scrapeIntervalStatus(opts.ScrapeTimeResourceUsage))
	Logger.Infof("set scape interval[Queries]: %v", scrapeIntervalStatus(opts.ScrapeTimeQuery))
	initMetricCollector()

	Logger.Infof("Starting http server on %s", opts.ServerBind)
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
			fmt.Println(err)
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	// ensure query paths and projects are splitted by '@'
	if opts.QueriesWithProjects != nil {
		queryError := false
		for _, query := range opts.QueriesWithProjects {
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
	if opts.ScrapeTimeProjects == nil {
		opts.ScrapeTimeProjects = &opts.ScrapeTime
	}

	if opts.ScrapeTimeRepository == nil {
		opts.ScrapeTimeRepository = &opts.ScrapeTime
	}

	if opts.ScrapeTimePullRequest == nil {
		opts.ScrapeTimePullRequest = &opts.ScrapeTime
	}

	if opts.ScrapeTimeBuild == nil {
		opts.ScrapeTimeBuild = &opts.ScrapeTime
	}

	if opts.ScrapeTimeRelease == nil {
		opts.ScrapeTimeRelease = &opts.ScrapeTime
	}

	if opts.ScrapeTimeDeployment == nil {
		opts.ScrapeTimeDeployment = &opts.ScrapeTime
	}

	if opts.ScrapeTimeStats == nil {
		opts.ScrapeTimeStats = &opts.ScrapeTime
	}

	if opts.ScrapeTimeResourceUsage == nil {
		opts.ScrapeTimeResourceUsage = &opts.ScrapeTime
	}

	if opts.StatsSummaryMaxAge == nil {
		opts.StatsSummaryMaxAge = opts.ScrapeTimeStats
	}

	if opts.ScrapeTimeQuery == nil {
		opts.ScrapeTimeQuery = &opts.ScrapeTime
	}
}

// Init and build Azure authorzier
func initAzureDevOpsConnection() {
	AzureDevopsClient = AzureDevops.NewAzureDevopsClient()
	if opts.AzureDevopsUrl != nil {
		AzureDevopsClient.HostUrl = opts.AzureDevopsUrl
	}

	Logger.Infof("using organization: %v", opts.AzureDevopsOrganisation)
	Logger.Infof("using apiversion: %v", opts.AzureDevopsApiVersion)
	Logger.Infof("using concurrency: %v", opts.RequestConcurrencyLimit)
	Logger.Infof("using retries: %v", opts.RequestRetries)

	AzureDevopsClient.SetOrganization(opts.AzureDevopsOrganisation)
	AzureDevopsClient.SetAccessToken(opts.AzureDevopsAccessToken)
	AzureDevopsClient.SetApiVersion(opts.AzureDevopsApiVersion)
	AzureDevopsClient.SetConcurrency(opts.RequestConcurrencyLimit)
	AzureDevopsClient.SetRetries(opts.RequestRetries)
	AzureDevopsClient.SetUserAgent(fmt.Sprintf("azure-devops-exporter/%v", Version))

	AzureDevopsClient.LimitBuildsPerDefinition = opts.LimitBuildsPerDefinition
	AzureDevopsClient.LimitReleasesPerDefinition = opts.LimitReleasesPerDefinition
	AzureDevopsClient.LimitDeploymentPerDefinition = opts.LimitDeploymentPerDefinition
	AzureDevopsClient.LimitReleaseDefinitionsPerProject = opts.LimitReleaseDefinitionsPerProject
}

func getAzureDevOpsProjects() (list AzureDevops.ProjectList) {
	rawList, err := AzureDevopsClient.ListProjects()

	if err != nil {
		panic(err)
	}

	list = rawList

	// whitelist
	if len(opts.AzureDevopsFilterProjects) > 0 {
		rawList = list
		list = AzureDevops.ProjectList{}
		for _, project := range rawList.List {
			if arrayStringContains(opts.AzureDevopsFilterProjects, project.Id) {
				list.List = append(list.List, project)
			}
		}
	}

	// blacklist
	if len(opts.AzureDevopsBlacklistProjects) > 0 {
		// filter ignored azure devops projects
		rawList = list
		list = AzureDevops.ProjectList{}
		for _, project := range rawList.List {
			if !arrayStringContains(opts.AzureDevopsBlacklistProjects, project.Id) {
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
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorGeneralList[collectorName] = NewCollectorGeneral(collectorName, &MetricsCollectorGeneral{})
		collectorGeneralList[collectorName].SetAzureProjects(&projectList)
		collectorGeneralList[collectorName].SetScrapeTime(*opts.ScrapeTimeLive)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Project"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorProject{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeLive)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "AgentPool"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorAgentPoolList[collectorName] = NewCollectorAgentPool(collectorName, &MetricsCollectorAgentPool{})
		collectorAgentPoolList[collectorName].SetAzureProjects(&projectList)
		collectorAgentPoolList[collectorName].AgentPoolIdList = opts.AzureDevopsFilterAgentPoolId
		collectorAgentPoolList[collectorName].SetScrapeTime(*opts.ScrapeTimeLive)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "LatestBuild"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorLatestBuild{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeLive)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Repository"
	if opts.ScrapeTimeRepository.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorRepository{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeRepository)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "PullRequest"
	if opts.ScrapeTimePullRequest.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorPullRequest{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimePullRequest)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Build"
	if opts.ScrapeTimeBuild.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorBuild{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeBuild)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Release"
	if opts.ScrapeTimeRelease.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorRelease{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeRelease)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Deployment"
	if opts.ScrapeTimeDeployment.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorDeployment{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeRelease)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Stats"
	if opts.ScrapeTimeStats.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorStats{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].SetScrapeTime(*opts.ScrapeTimeStats)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "ResourceUsage"
	if opts.ScrapeTimeStats.Seconds() > 0 {
		collectorGeneralList[collectorName] = NewCollectorGeneral(collectorName, &MetricsCollectorResourceUsage{})
		collectorGeneralList[collectorName].SetAzureProjects(&projectList)
		collectorGeneralList[collectorName].SetScrapeTime(*opts.ScrapeTimeStats)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
	}

	collectorName = "Query"
	if opts.ScrapeTimeQuery.Seconds() > 0 {
		collectorQueryList[collectorName] = NewCollectorQuery(collectorName, &MetricsCollectorQuery{})
		collectorQueryList[collectorName].SetAzureProjects(&projectList)
		collectorQueryList[collectorName].QueryList = opts.QueriesWithProjects
		collectorQueryList[collectorName].SetScrapeTime(*opts.ScrapeTimeLive)
	} else {
		Logger.Infof("collector[%s]: disabled", collectorName)
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
	if opts.ScrapeTimeProjects.Seconds() > 0 {
		go func() {
			// initial sleep
			time.Sleep(*opts.ScrapeTimeProjects)

			for {
				Logger.Info("daemon: updating project list")

				projectList := getAzureDevOpsProjects()
				Logger.Infof("daemon: found %v projects", projectList.Count)

				for _, collector := range collectorGeneralList {
					collector.SetAzureProjects(&projectList)
				}

				for _, collector := range collectorProjectList {
					collector.SetAzureProjects(&projectList)
				}

				for _, collector := range collectorAgentPoolList {
					collector.SetAzureProjects(&projectList)
				}

				Logger.Info("daemon: skipping Query collectors; they don't use projects")
				time.Sleep(*opts.ScrapeTimeProjects)
			}
		}()
	}
}

// start and handle prometheus handler
func startHttpServer() {
	http.Handle("/metrics", promhttp.Handler())
	Logger.Error(http.ListenAndServe(opts.ServerBind, nil))
}
