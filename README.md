Azure DevOps Exporter (VSTS)
============================

[![license](https://img.shields.io/github/license/webdevops/azure-devops-exporter.svg)](https://github.com/webdevops/azure-devops-exporter/blob/master/LICENSE)
[![Docker](https://img.shields.io/badge/docker-webdevops%2Fazure--devops--exporter-blue.svg?longCache=true&style=flat&logo=docker)](https://hub.docker.com/r/webdevops/azure-devops-exporter/)
[![Docker Build Status](https://img.shields.io/docker/build/webdevops/azure-devops-exporter.svg)](https://hub.docker.com/r/webdevops/azure-devops-exporter/)

Prometheus exporter for Azure DevOps (VSTS) for projects, builds, build times (elapsed and queue wait time), agent pool utilization and active pull requests.

Configuration
-------------

Normally no configuration is needed but can be customized using environment variables.

| Environment variable              | DefaultValue                | Description                                                       |
|-----------------------------------|-----------------------------|-------------------------------------------------------------------|
| `SCRAPE_TIME`                     | `2m`                        | Time (time.Duration) between API calls                            |
| `SERVER_BIND`                     | `:8080`                     | IP/Port binding                                                   |
| `AZURE_DEVOPS_ORGANISATION`       | none                        | Azure DevOps organisation (subdomain)                             |
| `AZURE_DEVOPS_ACCESS_TOKEN`       | none                        | Azure DevOps access token                                         |

Metrics
-------

| Metric                              | Description                                                                           |
|-------------------------------------|---------------------------------------------------------------------------------------|
| `azure_devops_project_info`         | Project informations                                                                  |
| `azure_devops_repository_info`      | Repository informations                                                               |
| `azure_devops_pullrequest_info`     | Active PullRequests                                                                   |
| `azure_devops_pullrequest_status`   | Status informations (eg. created date) for active PullRequests                        |
| `azure_devops_agentpool_info`       | Agent Pool informations                                                               |
| `azure_devops_agentpool_builds`     | Count of builds (by status)                                                           |
| `azure_devops_agentpool_wait`       | Queue wait time per agent pool (summary vector)                                       |
| `azure_devops_build_info`           | Build information (last build of definition)                                          |
| `azure_devops_build_status`         | Status informations of last build (queued, started, finished...)                      |
