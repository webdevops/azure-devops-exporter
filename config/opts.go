package config

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Debug       bool `long:"log.debug"    env:"LOG_DEBUG"  description:"debug mode"`
			Development bool `long:"log.devel"    env:"LOG_DEVEL"  description:"development mode"`
			Json        bool `long:"log.json"     env:"LOG_JSON"   description:"Switch log output to json format"`
		}

		// scrape time settings
		Scrape struct {
			Time              time.Duration  `long:"scrape.time"                  env:"SCRAPE_TIME"                    description:"Default scrape time (time.duration)"                       default:"30m"`
			TimeProjects      *time.Duration `long:"scrape.time.projects"         env:"SCRAPE_TIME_PROJECTS"           description:"Scrape time for project metrics (time.duration)"`
			TimeRepository    *time.Duration `long:"scrape.time.repository"       env:"SCRAPE_TIME_REPOSITORY"         description:"Scrape time for repository metrics (time.duration)"`
			TimeBuild         *time.Duration `long:"scrape.time.build"            env:"SCRAPE_TIME_BUILD"              description:"Scrape time for build metrics (time.duration)"`
			TimeRelease       *time.Duration `long:"scrape.time.release"          env:"SCRAPE_TIME_RELEASE"            description:"Scrape time for release metrics (time.duration)"`
			TimeDeployment    *time.Duration `long:"scrape.time.deployment"       env:"SCRAPE_TIME_DEPLOYMENT"         description:"Scrape time for deployment metrics (time.duration)"`
			TimePullRequest   *time.Duration `long:"scrape.time.pullrequest"      env:"SCRAPE_TIME_PULLREQUEST"        description:"Scrape time for pullrequest metrics  (time.duration)"`
			TimeStats         *time.Duration `long:"scrape.time.stats"            env:"SCRAPE_TIME_STATS"              description:"Scrape time for stats metrics  (time.duration)"`
			TimeResourceUsage *time.Duration `long:"scrape.time.resourceusage"    env:"SCRAPE_TIME_RESOURCEUSAGE"      description:"Scrape time for resourceusage metrics  (time.duration)"`
			TimeQuery         *time.Duration `long:"scrape.time.query"            env:"SCRAPE_TIME_QUERY"              description:"Scrape time for query results  (time.duration)"`
			TimeLive          *time.Duration `long:"scrape.time.live"             env:"SCRAPE_TIME_LIVE"               description:"Scrape time for live metrics (time.duration)"              default:"30s"`
		}

		// summary options
		Stats struct {
			SummaryMaxAge *time.Duration `long:"stats.summary.maxage"         env:"STATS_SUMMARY_MAX_AGE"             description:"Stats Summary metrics max age (time.duration)"`
		}

		// azure settings
		Azure struct {
			TenantId     string `long:"azure.tenant-id"               env:"AZURE_TENANT_ID"                description:"Azure tenant ID for Service Principal authentication"`
			ClientId     string `long:"azure.client-id"               env:"AZURE_CLIENT_ID"                description:"Client ID for Service Principal authentication"`
			ClientSecret string `long:"azure.client-secret"           env:"AZURE_CLIENT_SECRET"            description:"Client secret for Service Principal authentication" json:"-"`
		}

		// azure settings
		AzureDevops struct {
			Url             *string `long:"azuredevops.url"                     env:"AZURE_DEVOPS_URL"               description:"Azure DevOps URL (empty if hosted by Microsoft)"`
			AccessToken     string  `long:"azuredevops.access-token"            env:"AZURE_DEVOPS_ACCESS_TOKEN"      description:"Azure DevOps access token" json:"-"`
			AccessTokenFile *string `long:"azuredevops.access-token-file"       env:"AZURE_DEVOPS_ACCESS_TOKEN_FILE" description:"Azure DevOps access token (from file)"`
			Organisation    string  `long:"azuredevops.organisation"            env:"AZURE_DEVOPS_ORGANISATION"      description:"Azure DevOps organization" required:"true"`
			ApiVersion      string  `long:"azuredevops.apiversion"              env:"AZURE_DEVOPS_APIVERSION"        description:"Azure DevOps API version"  default:"5.1"`

			// agentpool
			AgentPoolIdList *[]int64 `long:"azuredevops.agentpool"  env:"AZURE_DEVOPS_AGENTPOOL"  env-delim:" "   description:"Enable scrape metrics for agent pool (IDs)"`

			// ignore settings
			FilterProjects    []string `long:"whitelist.project"    env:"AZURE_DEVOPS_FILTER_PROJECT"    env-delim:" "   description:"Filter projects (UUIDs)"`
			BlacklistProjects []string `long:"blacklist.project"    env:"AZURE_DEVOPS_BLACKLIST_PROJECT" env-delim:" "   description:"Filter projects (UUIDs)"`

			FilterTimelineState  []string `long:"timeline.state"    env:"AZURE_DEVOPS_FILTER_TIMELINE_STATE"    env-delim:" "   description:"Filter timeline states (completed, inProgress, pending)" default:"completed"`
			FetchAllBuildsFilter []string `long:"builds.all.project"   env:"AZURE_DEVOPS_FETCH_ALL_BUILDS_FILTER_PROJECT"  env-delim:" "  description:"Fetch all builds (even if they are not finished)"`

			// query settings
			QueriesWithProjects []string `long:"list.query"    env:"AZURE_DEVOPS_QUERIES"    env-delim:" "   description:"Pairs of query and project UUIDs in the form: '<queryId>@<projectId>'"`

			// tag settings
			TagsSchema                *[]string `long:"tags.schema"             env:"AZURE_DEVOPS_TAG_SCHEMA"              env-delim:" "   description:"Tags to be extracted from builds in the format 'tagName:type' with following types: number, info, bool"`
			TagsBuildDefinitionIdList *[]int64  `long:"tags.build.definition"   env:"AZURE_DEVOPS_TAG_BUILD_DEFINITION"    env-delim:" "   description:"Build definition ids to query tags (IDs)"`
		}

		// cache settings
		Cache struct {
			Path string `long:"cache.path" env:"CACHE_PATH" description:"Cache path (to folder, file://path... or azblob://storageaccount.blob.core.windows.net/containername or k8scm://{namespace}/{configmap}})"`
		}

		Request struct {
			ConcurrencyLimit int64 `long:"request.concurrency"                   env:"REQUEST_CONCURRENCY"     description:"Number of concurrent requests against dev.azure.com"  default:"10"`
			Retries          int   `long:"request.retries"                       env:"REQUEST_RETRIES"         description:"Number of retried requests against dev.azure.com"     default:"3"`
		}

		ServiceDiscovery struct {
			RefreshDuration time.Duration `long:"servicediscovery.refresh"  env:"SERVICEDISCOVERY_REFRESH"  description:"Refresh duration for servicediscovery (time.duration)"  default:"30m"`
		}

		Limit struct {
			Project                      int64         `long:"limit.project"                         env:"LIMIT_PROJECT"                         description:"Limit number of projects"         default:"100"`
			BuildsPerProject             int64         `long:"limit.builds-per-project"              env:"LIMIT_BUILDS_PER_PROJECT"              description:"Limit builds per project"         default:"100"`
			BuildsPerDefinition          int64         `long:"limit.builds-per-definition"           env:"LIMIT_BUILDS_PER_DEFINITION"           description:"Limit builds per definition"      default:"10"`
			ReleasesPerProject           int64         `long:"limit.releases-per-project"            env:"LIMIT_RELEASES_PER_PROJECT"            description:"Limit releases per project"       default:"100"`
			ReleasesPerDefinition        int64         `long:"limit.releases-per-definition"         env:"LIMIT_RELEASES_PER_DEFINITION"         description:"Limit releases per definition"    default:"100"`
			DeploymentPerDefinition      int64         `long:"limit.deployments-per-definition"      env:"LIMIT_DEPLOYMENTS_PER_DEFINITION"      description:"Limit deployments per definition" default:"100"`
			ReleaseDefinitionsPerProject int64         `long:"limit.releasedefinitions-per-project"  env:"LIMIT_RELEASEDEFINITION_PER_PROJECT"   description:"Limit builds per definition"      default:"100"`
			BuildHistoryDuration         time.Duration `long:"limit.build-history-duration"          env:"LIMIT_BUILD_HISTORY_DURATION"          description:"Time (time.Duration) how long the exporter should look back for builds"      default:"48h"`
			ReleaseHistoryDuration       time.Duration `long:"limit.release-history-duration"        env:"LIMIT_RELEASE_HISTORY_DURATION"        description:"Time (time.Duration) how long the exporter should look back for releases"      default:"48h"`
		}

		Server struct {
			// general options
			Bind         string        `long:"server.bind"              env:"SERVER_BIND"           description:"Server address"        default:":8080"`
			ReadTimeout  time.Duration `long:"server.timeout.read"      env:"SERVER_TIMEOUT_READ"   description:"Server read timeout"   default:"5s"`
			WriteTimeout time.Duration `long:"server.timeout.write"     env:"SERVER_TIMEOUT_WRITE"  description:"Server write timeout"  default:"10s"`
		}
	}
)

func (o *Opts) GetCachePath(path string) (ret *string) {
	if o.Cache.Path != "" {
		tmp := o.Cache.Path + "/" + path
		ret = &tmp
	}

	return
}

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
