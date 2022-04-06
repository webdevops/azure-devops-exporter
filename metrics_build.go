package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	prometheusCommon "github.com/webdevops/go-common/prometheus"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorBuild struct {
	CollectorProcessorProject

	prometheus struct {
		build       *prometheus.GaugeVec
		buildStatus *prometheus.GaugeVec

		buildDefinition *prometheus.GaugeVec

		buildStage *prometheus.GaugeVec
		buildPhase *prometheus.GaugeVec
		buildJob   *prometheus.GaugeVec
		buildTask  *prometheus.GaugeVec

		buildTimeProject *prometheus.SummaryVec
		jobTimeProject   *prometheus.SummaryVec
	}
}

func (m *MetricsCollectorBuild) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.build = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_info",
			Help: "Azure DevOps build",
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"buildID",
			"agentPoolID",
			"requestedBy",
			"buildNumber",
			"buildName",
			"sourceBranch",
			"sourceVersion",
			"status",
			"reason",
			"result",
			"url",
		},
	)
	prometheus.MustRegister(m.prometheus.build)

	m.prometheus.buildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_status",
			Help: "Azure DevOps build",
		},
		[]string{
			"projectID",
			"buildID",
			"buildDefinitionID",
			"buildNumber",
			"result",
			"type",
		},
	)
	prometheus.MustRegister(m.prometheus.buildStatus)

	m.prometheus.buildStage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_stage",
			Help: "Azure DevOps build stages",
		},
		[]string{
			"projectID",
			"buildID",
			"buildDefinitionID",
			"buildNumber",
			"name",
			"id",
			"identifier",
			"result",
			"type",
		},
	)
	prometheus.MustRegister(m.prometheus.buildStage)

	m.prometheus.buildPhase = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_phase",
			Help: "Azure DevOps build phases",
		},
		[]string{
			"projectID",
			"buildID",
			"buildDefinitionID",
			"buildNumber",
			"name",
			"id",
			"parentId",
			"identifier",
			"result",
			"type",
		},
	)
	prometheus.MustRegister(m.prometheus.buildPhase)

	m.prometheus.buildJob = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_job",
			Help: "Azure DevOps build jobs",
		},
		[]string{
			"projectID",
			"buildID",
			"buildDefinitionID",
			"buildNumber",
			"name",
			"id",
			"parentId",
			"identifier",
			"result",
			"type",
		},
	)
	prometheus.MustRegister(m.prometheus.buildJob)

	m.prometheus.buildTask = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_task",
			Help: "Azure DevOps build tasks",
		},
		[]string{
			"projectID",
			"buildID",
			"buildDefinitionID",
			"buildNumber",
			"name",
			"id",
			"parentId",
			"workerName",
			"result",
			"type",
		},
	)
	prometheus.MustRegister(m.prometheus.buildTask)

	m.prometheus.buildDefinition = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_definition_info",
			Help: "Azure DevOps build definition",
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"buildNameFormat",
			"buildDefinitionName",
			"path",
			"url",
		},
	)
	prometheus.MustRegister(m.prometheus.buildDefinition)
}

func (m *MetricsCollectorBuild) Reset() {
	m.prometheus.build.Reset()
	m.prometheus.buildDefinition.Reset()
	m.prometheus.buildStatus.Reset()
	m.prometheus.buildStage.Reset()
	m.prometheus.buildPhase.Reset()
	m.prometheus.buildJob.Reset()
	m.prometheus.buildTask.Reset()
}

func (m *MetricsCollectorBuild) Collect(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	m.collectDefinition(ctx, logger, callback, project)
	m.collectBuilds(ctx, logger, callback, project)
	m.collectBuildsTimeline(ctx, logger, callback, project)
}

func (m *MetricsCollectorBuild) collectDefinition(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListBuildDefinitions(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	buildDefinitonMetric := prometheusCommon.NewMetricsList()

	for _, buildDefinition := range list.List {
		buildDefinitonMetric.Add(prometheus.Labels{
			"projectID":           project.Id,
			"buildDefinitionID":   int64ToString(buildDefinition.Id),
			"buildNameFormat":     buildDefinition.BuildNameFormat,
			"buildDefinitionName": buildDefinition.Name,
			"path":                buildDefinition.Path,
			"url":                 buildDefinition.Links.Web.Href,
		}, 1)
	}

	callback <- func() {
		buildDefinitonMetric.GaugeSet(m.prometheus.buildDefinition)
	}
}

func (m *MetricsCollectorBuild) collectBuilds(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	minTime := time.Now().Add(-opts.Limit.BuildHistoryDuration)

	list, err := AzureDevopsClient.ListBuildHistory(project.Id, minTime)
	if err != nil {
		logger.Error(err)
		return
	}

	buildMetric := prometheusCommon.NewMetricsList()
	buildStatusMetric := prometheusCommon.NewMetricsList()

	for _, build := range list.List {
		buildMetric.AddInfo(prometheus.Labels{
			"projectID":         project.Id,
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildID":           int64ToString(build.Id),
			"buildNumber":       build.BuildNumber,
			"buildName":         build.Definition.Name,
			"agentPoolID":       int64ToString(build.Queue.Pool.Id),
			"requestedBy":       build.RequestedBy.DisplayName,
			"sourceBranch":      build.SourceBranch,
			"sourceVersion":     build.SourceVersion,
			"status":            build.Status,
			"reason":            build.Reason,
			"result":            build.Result,
			"url":               build.Links.Web.Href,
		})

		buildStatusMetric.AddBool(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"result":            build.Result,
			"type":              "succeeded",
		}, build.Result == "succeeded")

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"result":            build.Result,
			"type":              "queued",
		}, build.QueueTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"result":            build.Result,
			"type":              "started",
		}, build.StartTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"result":            build.Result,
			"type":              "finished",
		}, build.FinishTime)

		buildStatusMetric.AddDuration(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"result":            build.Result,
			"type":              "jobDuration",
		}, build.FinishTime.Sub(build.StartTime))
	}

	callback <- func() {
		buildMetric.GaugeSet(m.prometheus.build)
		buildStatusMetric.GaugeSet(m.prometheus.buildStatus)
	}
}

func (m *MetricsCollectorBuild) collectBuildsTimeline(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	minTime := time.Now().Add(-opts.Limit.BuildHistoryDuration)
	list, err := AzureDevopsClient.ListBuildHistoryWithStatus(project.Id, minTime, "completed")
	if err != nil {
		logger.Error(err)
		return
	}

	buildStageMetric := prometheusCommon.NewMetricsList()
	buildPhaseMetric := prometheusCommon.NewMetricsList()
	buildJobMetric := prometheusCommon.NewMetricsList()
	buildTaskMetric := prometheusCommon.NewMetricsList()

	for _, build := range list.List {
		timelineRecordList, _ := AzureDevopsClient.ListBuildTimeline(project.Id, int64ToString(build.Id))
		for _, timelineRecord := range timelineRecordList.List {
			switch recordType := timelineRecord.RecordType; recordType {
			case "Stage":
				buildStageMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "errorCount",
				}, timelineRecord.ErrorCount)

				buildStageMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "warningCount",
				}, timelineRecord.WarningCount)

				buildStageMetric.AddBool(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "succeeded",
				}, timelineRecord.Result == "succeeded")

				buildStageMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "started",
				}, timelineRecord.StartTime)

				buildStageMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "finished",
				}, timelineRecord.FinishTime)

				buildStageMetric.AddDuration(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "duration",
				}, timelineRecord.FinishTime.Sub(timelineRecord.StartTime))

			case "Phase":
				buildPhaseMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "errorCount",
				}, timelineRecord.ErrorCount)

				buildPhaseMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "warningCount",
				}, timelineRecord.WarningCount)

				buildPhaseMetric.AddBool(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "succeeded",
				}, timelineRecord.Result == "succeeded")

				buildPhaseMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "started",
				}, timelineRecord.StartTime)

				buildPhaseMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "finished",
				}, timelineRecord.FinishTime)

				buildPhaseMetric.AddDuration(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "duration",
				}, timelineRecord.FinishTime.Sub(timelineRecord.StartTime))
			case "Job":
				buildJobMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "errorCount",
				}, timelineRecord.ErrorCount)

				buildJobMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "warningCount",
				}, timelineRecord.WarningCount)

				buildJobMetric.AddBool(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "succeeded",
				}, timelineRecord.Result == "succeeded")

				buildJobMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "started",
				}, timelineRecord.StartTime)

				buildJobMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "finished",
				}, timelineRecord.FinishTime)

				buildJobMetric.AddDuration(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"identifier":        timelineRecord.Identifier,
					"result":            timelineRecord.Result,
					"type":              "duration",
				}, timelineRecord.FinishTime.Sub(timelineRecord.StartTime))
			case "Task":
				buildTaskMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"workerName":        timelineRecord.WorkerName,
					"result":            timelineRecord.Result,
					"type":              "errorCount",
				}, timelineRecord.ErrorCount)

				buildTaskMetric.Add(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"workerName":        timelineRecord.WorkerName,
					"result":            timelineRecord.Result,
					"type":              "warningCount",
				}, timelineRecord.WarningCount)

				buildTaskMetric.AddBool(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"workerName":        timelineRecord.WorkerName,
					"result":            timelineRecord.Result,
					"type":              "succeeded",
				}, timelineRecord.Result == "succeeded")

				buildTaskMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"workerName":        timelineRecord.WorkerName,
					"result":            timelineRecord.Result,
					"type":              "started",
				}, timelineRecord.StartTime)

				buildTaskMetric.AddTime(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"workerName":        timelineRecord.WorkerName,
					"result":            timelineRecord.Result,
					"type":              "finished",
				}, timelineRecord.FinishTime)

				buildTaskMetric.AddDuration(prometheus.Labels{
					"projectID":         project.Id,
					"buildID":           int64ToString(build.Id),
					"buildDefinitionID": int64ToString(build.Definition.Id),
					"buildNumber":       build.BuildNumber,
					"name":              timelineRecord.Name,
					"id":                timelineRecord.Id,
					"parentId":          timelineRecord.ParentId,
					"workerName":        timelineRecord.WorkerName,
					"result":            timelineRecord.Result,
					"type":              "duration",
				}, timelineRecord.FinishTime.Sub(timelineRecord.StartTime))

			}

		}
	}

	callback <- func() {
		buildStageMetric.GaugeSet(m.prometheus.buildStage)
		buildPhaseMetric.GaugeSet(m.prometheus.buildPhase)
		buildJobMetric.GaugeSet(m.prometheus.buildJob)
		buildTaskMetric.GaugeSet(m.prometheus.buildTask)
	}
}
