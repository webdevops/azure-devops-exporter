Azure DevOps Exporter (VSTS)
============================

[![license](https://img.shields.io/github/license/webdevops/azure-devops-exporter.svg)](https://github.com/webdevops/azure-devops-exporter/blob/master/LICENSE)
[![Docker](https://img.shields.io/badge/docker-webdevops%2Fazure--devops--exporter-blue.svg?longCache=true&style=flat&logo=docker)](https://hub.docker.com/r/webdevops/azure-devops-exporter/)
[![Docker Build Status](https://img.shields.io/docker/build/webdevops/azure-devops-exporter.svg)](https://hub.docker.com/r/webdevops/azure-devops-exporter/)

Prometheus exporter for Azure DevOps (VSTS) for projects, builds, build times (elapsed and queue wait time), agent pool utilization and active pull requests.

Configuration
-------------

Normally no configuration is needed but can be customized using environment variables.

| Environment variable                  | DefaultValue                | Description                                                                      |
|---------------------------------------|-------------------------------------|--------------------------------------------------------------------------|
| `SCRAPE_TIME`                         | `30m`                               | Interval (time.Duration) between API calls                               |
| `SCRAPE_TIME_PROJECTS`                | not set, default see `SCRAPE_TIME`  | Interval for project metrics (list of projects for all scrapers)         |
| `SCRAPE_TIME_REPOSITORY`              | not set, default see `SCRAPE_TIME`  | Interval for repository metrics                                          |
| `SCRAPE_TIME_BUILD`                   | not set, default see `SCRAPE_TIME`  | Interval for build metrics                                               |
| `SCRAPE_TIME_RELEASE`                 | not set, default see `SCRAPE_TIME`  | Interval for release metrics                                             |
| `SCRAPE_TIME_DEPLOYMENT`              | not set, default see `SCRAPE_TIME`  | Interval for deployment metrics                                          |
| `SCRAPE_TIME_PULLREQUEST`             | not set, default see `SCRAPE_TIME`  | Interval for pullrequest metrics                                         |
| `SCRAPE_TIME_LIVE`                    | `30s`                               | Time (time.Duration) between API calls                                   |
| `SERVER_BIND`                         | `:8080`                             | IP/Port binding                                                          |
| `AZURE_DEVOPS_ORGANISATION`           | none                                | Azure DevOps organisation (subdomain)                                    |
| `AZURE_DEVOPS_ACCESS_TOKEN`           | none                                | Azure DevOps access token                                                |
| `AZURE_DEVOPS_FILTER_PROJECT`         | none                                | Whitelist project uuids                                                  |
| `AZURE_DEVOPS_BLACKLIST_PROJECT`      | none                                | Blacklist project uuids                                                  |
| `AZURE_DEVOPS_FILTER_AGENTPOOL`       | none                                | Whitelist for agentpool metrics                                          |
| `REQUEST_CONCURRENCY`                 | `10`                                | API request concurrency (number of calls at the same time)               |
| `REQUEST_RETRIES`                     | `3`                                 | API request retries in case of failure                                   |
| `LIMIT_BUILDS_PER_DEFINITION`         | `10`                                | Fetched builds per definition                                            |
| `LIMIT_RELEASES_PER_DEFINITION`       | `100`                               | Fetched releases per definition                                          |
| `LIMIT_DEPLOYMENTS_PER_DEFINITION`    | `100`                               | Fetched deployments per definition                                       |
| `LIMIT_RELEASEDEFINITION_PER_PROJECT` | `100`                               | Fetched builds per definition                                            |


Metrics
-------

| Metric                                          | Scraper       | Description                                                                          |
|-------------------------------------------------|------------------------------------------------------------------------------------------------------|
| `azure_devops_stats`                            | live          | General scraper stats                                                                |
| `azure_devops_agentpool_info`                   | live          | Agent Pool informations                                                              |
| `azure_devops_agentpool_builds`                 | live          | Count of builds (by status)                                                          |
| `azure_devops_agentpool_wait`                   | live          | Queue wait time per agent pool (summary vector)                                      |
| `azure_devops_agentpool_size`                   | live          | Queue size per agent pool                                                            |
| `azure_devops_agentpool_agent_info`             | live          | Agent information per agent pool                                                     |
| `azure_devops_agentpool_agent_status`           | live          | Status informations (eg. created date) for each agent in a agent pool                |
| `azure_devops_agentpool_agent_job`              | live          | Currently running jobs on each agent                                                 |
| `azure_devops_project_info`                     | live/projects | Project informations                                                                 |
| `azure_devops_build_latest_info`                | live          | Latest build information                                                             |
| `azure_devops_build_latest_status`              | live          | Latest build status informations                                                     |
| `azure_devops_pullrequest_info`                 | pullrequest   | Active PullRequests                                                                  |
| `azure_devops_pullrequest_status`               | pullrequest   | Status informations (eg. created date) for active PullRequests                       |
| `azure_devops_pullrequest_label`                | pullrequest   | Labels set on active PullRequests                                                    |
| `azure_devops_build_info`                       | build         | Build informations                                                                   |
| `azure_devops_build_status`                     | build         | Build status infos (queued, started, finished time)                                  |
| `azure_devops_build_definition_info`            | build         | Build definition info                                                                |
| `azure_devops_release_info`                     | release       | Release informations                                                                 |
| `azure_devops_release_artifact`                 | release       | Release artifcact informations                                                       |
| `azure_devops_release_environment`              | release       | Release environment list                                                             |
| `azure_devops_release_environment_status`       | release       | Release environment status informations                                              |
| `azure_devops_release_approval`                 | release       | Release environment approval list                                                    |
| `azure_devops_release_definition_info`          | release       | Release definition info                                                              |
| `azure_devops_release_definition_environment`   | release       | Release definition environment list                                                  |
| `azure_devops_repository_info`                  | repository    | Repository informations                                                              |
| `azure_devops_repository_stats`                 | repository    | Repository stats                                                                     |
| `azure_devops_repository_commits`               | repository    | Repository commit counter                                                            |
| `azure_devops_repository_pushes`                | repository    | Repository push counter                                                              |
| `azure_devops_deployment_info`                  | deployment    | Release deployment informations                                                      |
| `azure_devops_deployment_status`                | deployment    | Release deployment status informations                                               |


Usage
-----

```
Usage:
  azure-devops-exporter [OPTIONS]

Application Options:
  -v, --verbose                               Verbose mode [$VERBOSE]
      --bind=                                 Server address (default: :8080) [$SERVER_BIND]
      --scrape.time=                          Default scrape time (time.duration) (default: 30m) [$SCRAPE_TIME]
      --scrape.time.projects=                 Scrape time for project metrics (time.duration) [$SCRAPE_TIME_PROJECTS]
      --scrape.time.repository=               Scrape time for repository metrics (time.duration) [$SCRAPE_TIME_REPOSITORY]
      --scrape.time.build=                    Scrape time for build metrics (time.duration) [$SCRAPE_TIME_BUILD]
      --scrape.time.release=                  Scrape time for release metrics (time.duration) [$SCRAPE_TIME_RELEASE]
      --scrape.time.deployment=               Scrape time for deployment metrics (time.duration) [$SCRAPE_TIME_DEPLOYMENT]
      --scrape.time.pullrequest=              Scrape time for pullrequest metrics  (time.duration) [$SCRAPE_TIME_PULLREQUEST]
      --scrape.time.live=                     Scrape time for live metrics (time.duration) (default: 30s) [$SCRAPE_TIME_LIVE]
      --whitelist.project=                    Filter projects (UUIDs) [$AZURE_DEVOPS_FILTER_PROJECT]
      --blacklist.project=                    Filter projects (UUIDs) [$AZURE_DEVOPS_BLACKLIST_PROJECT]
      --whitelist.agentpool=                  Filter of agent pool (IDs) [$AZURE_DEVOPS_FILTER_AGENTPOOL]
      --azuredevops.access-token=             Azure DevOps access token [$AZURE_DEVOPS_ACCESS_TOKEN]
      --azuredevops.organisation=             Azure DevOps organization [$AZURE_DEVOPS_ORGANISATION]
      --request.concurrency=                  Number of concurrent requests against dev.azure.com (default: 10) [$REQUEST_CONCURRENCY]
      --request.retries=                      Number of retried requests against dev.azure.com (default: 3) [$REQUEST_RETRIES]
      --limit.builds-per-definition=          Limit builds per definition (default: 10) [$LIMIT_BUILDS_PER_DEFINITION]
      --limit.releases-per-definition=        Limit releases per definition (default: 100) [$LIMIT_RELEASES_PER_DEFINITION]
      --limit.deployments-per-definition=     Limit deployments per definition (default: 100) [$LIMIT_DEPLOYMENTS_PER_DEFINITION]
      --limit.releasedefinitions-per-project= Limit builds per definition (default: 100) [$LIMIT_RELEASEDEFINITION_PER_PROJECT]

Help Options:
  -h, --help                                  Show this help message
```
