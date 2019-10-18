FROM golang:1.13 as build

WORKDIR /go/src/github.com/keremispirli/azure-devops-exporter/tree/addQueryMetrics

# Get deps (cached)
COPY ./go.mod /go/src/github.com/keremispirli/azure-devops-exporter/tree/addQueryMetrics
COPY ./go.sum /go/src/github.com/keremispirli/azure-devops-exporter/tree/addQueryMetrics
RUN go mod download

# Compile
COPY ./ /go/src/github.com/keremispirli/azure-devops-exporter/tree/addQueryMetrics
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /azure-devops-exporter \
    && chmod +x /azure-devops-exporter
RUN /azure-devops-exporter --help

#############################################
# FINAL IMAGE
#############################################
FROM gcr.io/distroless/static
COPY --from=build /azure-devops-exporter /
USER 1000
ENTRYPOINT ["/azure-devops-exporter"]
