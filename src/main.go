package main

import (
	AzureDevops "azure-devops-exporter/src/azure-devops-client"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)

const (
	Author  = "webdevops.io"
	Version = "0.4.0"
)

var (
	argparser         *flags.Parser
	args              []string
	Logger            *DaemonLogger
	ErrorLogger       *DaemonLogger
	AzureDevopsClient *AzureDevops.AzureDevopsClient

	collectorGeneralList   map[string]*CollectorGeneral
	collectorProjectList   map[string]*CollectorProject
	collectorAgentPoolList map[string]*CollectorAgentPool
)

var opts struct {
	// general settings
	Verbose []bool `   long:"verbose" short:"v"                   env:"VERBOSE"                       description:"Verbose mode"`

	// server settings
	ServerBind string `long:"bind"                                env:"SERVER_BIND"                   description:"Server address"                                    default:":8080"`

	// scrape time settings
	ScrapeTime            time.Duration  `long:"scrape-time"                  env:"SCRAPE_TIME"                    description:"Default scrape time (time.duration)"                       default:"30m"`
	ScrapeTimeProjects    *time.Duration `long:"scrape-time-projects"         env:"SCRAPE_TIME_PROJECTS"           description:"Scrape time for project metrics (time.duration)"`
	ScrapeTimeRepository  *time.Duration `long:"scrape-time-repository"       env:"SCRAPE_TIME_REPOSITORY"         description:"Scrape time for repository metrics (time.duration)"`
	ScrapeTimeBuild       *time.Duration `long:"scrape-time-build"            env:"SCRAPE_TIME_BUILD"              description:"Scrape time for build metrics (time.duration)"`
	ScrapeTimeRelease     *time.Duration `long:"scrape-time-release"          env:"SCRAPE_TIME_RELEASE"            description:"Scrape time for release metrics (time.duration)"`
	ScrapeTimePullRequest *time.Duration `long:"scrape-time-pullrequest"      env:"SCRAPE_TIME_PULLREQUEST"        description:"Scrape time for pullrequest metrics  (time.duration)"`
	ScrapeTimeLive        *time.Duration `long:"scrape-time-live"             env:"SCRAPE_TIME_LIVE"               description:"Scrape time for live metrics (time.duration)"              default:"30s"`

	// ignore settings
	AzureDevopsFilterProjects    []string `long:"azure-devops-filter-project"    env:"AZURE_DEVOPS_FILTER_PROJECT"    env-delim:" "   description:"Filter projects (UUIDs)"`
	AzureDevopsBlacklistProjects []string `long:"azure-devops-blacklist-project" env:"AZURE_DEVOPS_BLACKLIST_PROJECT" env-delim:" "   description:"Filter projects (UUIDs)"`
	AzureDevopsFilterAgentPoolId []int64  `long:"azure-devops-filter-agentpool"  env:"AZURE_DEVOPS_FILTER_AGENTPOOL"  env-delim:" "   description:"Filter of agent pool (IDs)"`

	// azure settings
	AzureDevopsAccessToken  string `long:"azure-devops-access-token"            env:"AZURE_DEVOPS_ACCESS_TOKEN"                      description:"Azure DevOps access token" required:"true"`
	AzureDevopsOrganisation string `long:"azure-devops-organisation"            env:"AZURE_DEVOPS_ORGANISATION"                      description:"Azure DevOps organization" required:"true"`
}

func main() {
	initArgparser()

	Logger = CreateDaemonLogger(0)
	ErrorLogger = CreateDaemonErrorLogger(0)

	// set verbosity
	Verbose = len(opts.Verbose) >= 1

	Logger.Messsage("Init Azure DevOps exporter v%s (written by %v)", Version, Author)

	Logger.Messsage("Init Azure connection")
	initAzureConnection()

	Logger.Messsage("Starting metrics collection")
	Logger.Messsage("  scape time: %v", opts.ScrapeTime)
	initMetricCollector()

	Logger.Messsage("Starting http server on %s", opts.ServerBind)
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

	if opts.ScrapeTimeLive == nil {
		opts.ScrapeTimeLive = &opts.ScrapeTime
	}
}

// Init and build Azure authorzier
func initAzureConnection() {
	AzureDevopsClient = AzureDevops.NewAzureDevopsClient()
	AzureDevopsClient.SetOrganization(opts.AzureDevopsOrganisation)
	AzureDevopsClient.SetAccessToken(opts.AzureDevopsAccessToken)
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

	projectList := getAzureDevOpsProjects()

	collectorName = "General"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorGeneralList[collectorName] = NewCollectorGeneral(collectorName, &MetricsCollectorGeneral{})
		collectorGeneralList[collectorName].SetAzureProjects(&projectList)
		collectorGeneralList[collectorName].Run(*opts.ScrapeTimeLive)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "Project"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorProject{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].Run(*opts.ScrapeTimeLive)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "AgentPool"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorAgentPoolList[collectorName] = NewCollectorAgentPool(collectorName, &MetricsCollectorAgentPool{})
		collectorAgentPoolList[collectorName].SetAzureProjects(&projectList)
		collectorAgentPoolList[collectorName].AgentPoolIdList = opts.AzureDevopsFilterAgentPoolId
		collectorAgentPoolList[collectorName].Run(*opts.ScrapeTimeLive)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "LatestBuild"
	if opts.ScrapeTimeLive.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorLatestBuild{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].Run(*opts.ScrapeTimeLive)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "Repository"
	if opts.ScrapeTimeRepository.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorRepository{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].Run(*opts.ScrapeTimeRepository)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "PullRequest"
	if opts.ScrapeTimePullRequest.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorPullRequest{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].Run(*opts.ScrapeTimePullRequest)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "Build"
	if opts.ScrapeTimeBuild.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorBuild{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].Run(*opts.ScrapeTimeBuild)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	collectorName = "Release"
	if opts.ScrapeTimeRelease.Seconds() > 0 {
		collectorProjectList[collectorName] = NewCollectorProject(collectorName, &MetricsCollectorRelease{})
		collectorProjectList[collectorName].SetAzureProjects(&projectList)
		collectorProjectList[collectorName].Run(*opts.ScrapeTimeRelease)
	} else {
		Logger.Messsage("collector[%s]: disabled", collectorName)
	}

	// background auto update of projects
	go func() {
		// initial sleep
		time.Sleep(*opts.ScrapeTimeProjects)

		for {
			Logger.Messsage("daemon: updating project list")

			projectList := getAzureDevOpsProjects()

			for _, collector := range collectorGeneralList {
				collector.SetAzureProjects(&projectList)
			}

			for _, collector := range collectorProjectList {
				collector.SetAzureProjects(&projectList)
			}

			for _, collector := range collectorAgentPoolList {
				collector.SetAzureProjects(&projectList)
			}

			Logger.Messsage("daemon: found %v projects", projectList.Count)
			time.Sleep(*opts.ScrapeTimeProjects)
		}
	}()
}

// start and handle prometheus handler
func startHttpServer() {
	http.Handle("/metrics", promhttp.Handler())
	ErrorLogger.Fatal(http.ListenAndServe(opts.ServerBind, nil))
}
