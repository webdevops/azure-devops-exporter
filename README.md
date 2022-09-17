Azure DevOps Exporter (VSTS)
============================

[![license](https://img.shields.io/github/license/webdevops/azure-devops-exporter.svg)](https://github.com/webdevops/azure-devops-exporter/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fazure--devops--exporter-blue)](https://hub.docker.com/r/webdevops/azure-devops-exporter/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fazure--devops--exporter-blue)](https://quay.io/repository/webdevops/azure-devops-exporter)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/azure-devops-exporter)](https://artifacthub.io/packages/search?repo=azure-devops-exporter)

Prometheus exporter for Azure DevOps (VSTS) for projects, builds, build times (elapsed and queue wait time), agent pool utilization and active pull requests.

Configuration
-------------

```
Usage:
  azure-devops-exporter [OPTIONS]

Application Options:
      --debug                                 debug mode [$DEBUG]
  -v, --verbose                               verbose mode [$VERBOSE]
      --log.json                              Switch log output to json format [$LOG_JSON]
      --scrape.time=                          Default scrape time (time.duration) (default: 30m) [$SCRAPE_TIME]
      --scrape.time.projects=                 Scrape time for project metrics (time.duration) [$SCRAPE_TIME_PROJECTS]
      --scrape.time.repository=               Scrape time for repository metrics (time.duration) [$SCRAPE_TIME_REPOSITORY]
      --scrape.time.build=                    Scrape time for build metrics (time.duration) [$SCRAPE_TIME_BUILD]
      --scrape.time.release=                  Scrape time for release metrics (time.duration) [$SCRAPE_TIME_RELEASE]
      --scrape.time.deployment=               Scrape time for deployment metrics (time.duration) [$SCRAPE_TIME_DEPLOYMENT]
      --scrape.time.pullrequest=              Scrape time for pullrequest metrics  (time.duration) [$SCRAPE_TIME_PULLREQUEST]
      --scrape.time.stats=                    Scrape time for stats metrics  (time.duration) [$SCRAPE_TIME_STATS]
      --scrape.time.resourceusage=            Scrape time for resourceusage metrics  (time.duration) [$SCRAPE_TIME_RESOURCEUSAGE]
      --scrape.time.query=                    Scrape time for query results  (time.duration) [$SCRAPE_TIME_QUERY]
      --scrape.time.live=                     Scrape time for live metrics (time.duration) (default: 30s) [$SCRAPE_TIME_LIVE]
      --stats.summary.maxage=                 Stats Summary metrics max age (time.duration) [$STATS_SUMMARY_MAX_AGE]
      --azuredevops.url=                      Azure DevOps url (empty if hosted by microsoft) [$AZURE_DEVOPS_URL]
      --azuredevops.access-token=             Azure DevOps access token [$AZURE_DEVOPS_ACCESS_TOKEN]
      --azuredevops.access-token-file=        Azure DevOps access token (from file) [$AZURE_DEVOPS_ACCESS_TOKEN_FILE]
      --azuredevops.organisation=             Azure DevOps organization [$AZURE_DEVOPS_ORGANISATION]
      --azuredevops.apiversion=               Azure DevOps API version (default: 5.1) [$AZURE_DEVOPS_APIVERSION]
      --azuredevops.agentpool=                Enable scrape metrics for agent pool (IDs) [$AZURE_DEVOPS_AGENTPOOL]
      --whitelist.project=                    Filter projects (UUIDs) [$AZURE_DEVOPS_FILTER_PROJECT]
      --blacklist.project=                    Filter projects (UUIDs) [$AZURE_DEVOPS_BLACKLIST_PROJECT]
      --list.query=                           Pairs of query and project UUIDs in the form: '<queryId>@<projectId>'
                                              [$AZURE_DEVOPS_QUERIES]
      --cache.expiry=                         Internal cache expiry time (time.duration) (default: 30m) [$CACHE_EXPIRY]
      --request.concurrency=                  Number of concurrent requests against dev.azure.com (default: 10)
                                              [$REQUEST_CONCURRENCY]
      --request.retries=                      Number of retried requests against dev.azure.com (default: 3) [$REQUEST_RETRIES]
      --limit.project=                        Limit number of projects (default: 100) [$LIMIT_PROJECT]
      --limit.builds-per-project=             Limit builds per project (default: 100) [$LIMIT_BUILDS_PER_PROJECT]
      --limit.builds-per-definition=          Limit builds per definition (default: 10) [$LIMIT_BUILDS_PER_DEFINITION]
      --limit.releases-per-project=           Limit releases per project (default: 100) [$LIMIT_RELEASES_PER_PROJECT]
      --limit.releases-per-definition=        Limit releases per definition (default: 100) [$LIMIT_RELEASES_PER_DEFINITION]
      --limit.deployments-per-definition=     Limit deployments per definition (default: 100) [$LIMIT_DEPLOYMENTS_PER_DEFINITION]
      --limit.releasedefinitions-per-project= Limit builds per definition (default: 100) [$LIMIT_RELEASEDEFINITION_PER_PROJECT]
      --limit.build-history-duration=         Time (time.Duration) how long the exporter should look back for builds (default:
                                              48h) [$LIMIT_BUILD_HISTORY_DURATION]
      --limit.release-history-duration=       Time (time.Duration) how long the exporter should look back for releases (default:
                                              48h) [$LIMIT_RELEASE_HISTORY_DURATION]
      --server.bind=                          Server address (default: :8080) [$SERVER_BIND]
      --server.timeout.read=                  Server read timeout (default: 5s) [$SERVER_TIMEOUT_READ]
      --server.timeout.write=                 Server write timeout (default: 10s) [$SERVER_TIMEOUT_WRITE]

Help Options:
  -h, --help                                  Show this help message
```

Metrics
-------

| Metric                                         | Scraper       | Description                                                                             |
|------------------------------------------------|---------------|-----------------------------------------------------------------------------------------|
| `azure_devops_stats`                           | live          | General scraper stats                                                                   |
| `azure_devops_agentpool_info`                  | live          | Agent Pool informations                                                                 |
| `azure_devops_agentpool_size`                  | live          | Number of agents per agent pool                                                         |
| `azure_devops_agentpool_usage`                 | live          | Usage of agent pool (used agents; percent 0-1)                                          |
| `azure_devops_agentpool_queue_length`          | live          | Queue length per agent pool                                                             |
| `azure_devops_agentpool_agent_info`            | live          | Agent information per agent pool                                                        |
| `azure_devops_agentpool_agent_status`          | live          | Status informations (eg. created date) for each agent in a agent pool                   |
| `azure_devops_agentpool_agent_job`             | live          | Currently running jobs on each agent                                                    |
| `azure_devops_project_info`                    | live/projects | Project informations                                                                    |
| `azure_devops_build_latest_info`               | live          | Latest build information                                                                |
| `azure_devops_build_latest_status`             | live          | Latest build status informations                                                        |
| `azure_devops_pullrequest_info`                | pullrequest   | Active PullRequests                                                                     |
| `azure_devops_pullrequest_status`              | pullrequest   | Status informations (eg. created date) for active PullRequests                          |
| `azure_devops_pullrequest_label`               | pullrequest   | Labels set on active PullRequests                                                       |
| `azure_devops_build_info`                      | build         | Build informations                                                                      |
| `azure_devops_build_status`                    | build         | Build status infos (queued, started, finished time)                                     |
| `azure_devops_build_stage`                     | build         | Build stage infos (duration, errors, warnings, started, finished time)                  |
| `azure_devops_build_phase`                     | build         | Build phase infos (duration, errors, warnings, started, finished time)                  |
| `azure_devops_build_job`                       | build         | Build job infos (duration, errors, warnings, started, finished time)                    |
| `azure_devops_build_task`                      | build         | Build task infos (duration, errors, warnings, started, finished time)                   |
| `azure_devops_build_definition_info`           | build         | Build definition info                                                                   |
| `azure_devops_release_info`                    | release       | Release informations                                                                    |
| `azure_devops_release_artifact`                | release       | Release artifcact informations                                                          |
| `azure_devops_release_environment`             | release       | Release environment list                                                                |
| `azure_devops_release_environment_status`      | release       | Release environment status informations                                                 |
| `azure_devops_release_approval`                | release       | Release environment approval list                                                       |
| `azure_devops_release_definition_info`         | release       | Release definition info                                                                 |
| `azure_devops_release_definition_environment`  | release       | Release definition environment list                                                     |
| `azure_devops_repository_info`                 | repository    | Repository informations                                                                 |
| `azure_devops_repository_stats`                | repository    | Repository stats                                                                        |
| `azure_devops_repository_commits`              | repository    | Repository commit counter                                                               |
| `azure_devops_repository_pushes`               | repository    | Repository push counter                                                                 |
| `azure_devops_query_result`                    | live          | Latest results of given queries                                                         |
| `azure_devops_deployment_info`                 | deployment    | Release deployment informations                                                         |
| `azure_devops_deployment_status`               | deployment    | Release deployment status informations                                                  |
| `azure_devops_stats_agentpool_builds`          | stats         | Number of buildsper agentpool, project and result (counter)                             |
| `azure_devops_stats_agentpool_builds_wait`     | stats         | Build wait time per agentpool, project and result (summary)                             |
| `azure_devops_stats_agentpool_builds_duration` | stats         | Build duration per agentpool, project and result (summary)                              |
| `azure_devops_stats_project_builds`            | stats         | Number of builds per project, definition and result (counter)                           |
| `azure_devops_stats_project_builds_wait`       | stats         | Build wait time per project, definition and result (summary)                            |
| `azure_devops_stats_project_builds_success`    | stats         | Success rating of build per project and definition (summary)                            |
| `azure_devops_stats_project_builds_duration`   | stats         | Build duration per project, definition and result (summary)                             |
| `azure_devops_stats_project_release_duration`  | stats         | Release environment duration per project, definition, environment and result (summary)  |
| `azure_devops_stats_project_release_success`   | stats         | Success rating of release environment per project, definition and environment (summary) |
| `azure_devops_resourceusage_build`             | resourceusage | Usage of limited and paid Azure DevOps resources (build)                                |
| `azure_devops_resourceusage_license`           | resourceusage | Usage of limited and paid Azure DevOps resources (license)                              |
| `azure_devops_api_request_*`                   |               | REST api request histogram (count, latency, statuscCodes)                               |


Prometheus queries
------------------

Last 3 failed releases per definition for one project
```
topk by(projectID,releaseDefinitionName,path) (3,
  azure_devops_release_environment{projectID="XXXXXXXXXXXXXXXX", status!="succeeded", status!="inProgress"}
  * on (projectID,releaseID,environmentID) group_left() (azure_devops_release_environment_status{type="created"})
  * on (projectID,releaseID) group_left(releaseName, releaseDefinitionID) (azure_devops_release_info)
  * on (projectID,releaseDefinitionID) group_left(path, releaseDefinitionName) (azure_devops_release_definition_info)
)
```

Agent pool usage (without PoolMaintenance)
```
count by(agentPoolID) (
  azure_devops_agentpool_agent_job{planType!="PoolMaintenance"}
  * on(agentPoolAgentID) group_left(agentPoolID) (azure_devops_agentpool_agent_info)
)
/ on (agentPoolID) group_left() (azure_devops_agentpool_size)
* on (agentPoolID) group_left(agentPoolName) (azure_devops_agentpool_info)
```

Current running jobs
```
label_replace(
    azure_devops_agentpool_agent_job{planType!="PoolMaintenance"}
    * on (agentPoolAgentID) group_left(agentPoolID,agentPoolAgentName) azure_devops_agentpool_agent_info
    * on (agentPoolID) group_left(agentPoolName) (azure_devops_agentpool_info)
  , "projectID", "$1", "scopeID", "^(.+)$"
)
* on (projectID) group_left(projectName) (azure_devops_project_info)
```

Agent pool size
```
azure_devops_agentpool_info
* on (agentPoolID) group_left() (azure_devops_agentpool_size)
```

Agent pool size (enabled and online)
```
azure_devops_agentpool_info
* on (agentPoolID) group_left() (
  count by(agentPoolID) (azure_devops_agentpool_agent_info{status="online",enabled="true"})
)
```
