package main

import (
	"os"
	"fmt"
	"time"
	"net/http"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	AzureDevops "azure-devops-exporter/src/azure-devops-client"
)

const (
	Author  = "webdevops.io"
	Version = "0.1.0"
	AZURE_RESOURCEGROUP_TAG_PREFIX = "tag_"
)

var (
	argparser          *flags.Parser
	args               []string
	Logger             *DaemonLogger
	ErrorLogger        *DaemonLogger
	AzureDevopsClient  *AzureDevops.AzureDevopsClient
)

var opts struct {
	// general settings
	Verbose     []bool `       long:"verbose" short:"v"               env:"VERBOSE"                         description:"Verbose mode"`

	// server settings
	ServerBind  string `       long:"bind"                            env:"SERVER_BIND"                     description:"Server address"               default:":8080"`
	ScrapeTime  time.Duration `long:"scrape-time"                     env:"SCRAPE_TIME"                     description:"Scrape time (time.duration)"  default:"15m"`

	// azure settings
	AzureDevopsAccessToken string ` long:"azure-devops-access-token"  env:"AZURE_DEVOPS_ACCESS_TOKEN"       description:"Azure DevOps access token" required:"true"`
	AzureDevopsOrganisation string `long:"azure-devops-organisation"  env:"AZURE_DEVOPS_ORGANISATION"       description:"Azure DevOps organization" required:"true"`
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
	setupMetricsCollection()
	startMetricsCollection()

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
}

// Init and build Azure authorzier
func initAzureConnection() {
	AzureDevopsClient = AzureDevops.NewAzureDevopsClient()
	AzureDevopsClient.SetOrganization(opts.AzureDevopsOrganisation)
	AzureDevopsClient.SetAccessToken(opts.AzureDevopsAccessToken)
}

// start and handle prometheus handler
func startHttpServer() {
	http.Handle("/metrics", promhttp.Handler())
	ErrorLogger.Fatal(http.ListenAndServe(opts.ServerBind, nil))
}
