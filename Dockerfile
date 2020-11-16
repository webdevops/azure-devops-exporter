FROM golang:1.15 as build

WORKDIR /go/src/github.com/webdevops/azure-devops-exporter

# Get deps (cached)
COPY ./go.mod /go/src/github.com/webdevops/azure-devops-exporter
COPY ./go.sum /go/src/github.com/webdevops/azure-devops-exporter
COPY ./Makefile /go/src/github.com/webdevops/azure-devops-exporter
RUN make dependencies

# Compile
COPY ./ /go/src/github.com/webdevops/azure-devops-exporter
RUN make test
RUN make lint
RUN make build
RUN ./azure-devops-exporter --help

#############################################
# FINAL IMAGE
#############################################
FROM gcr.io/distroless/static
ENV LOG_JSON=1
COPY --from=build /go/src/github.com/webdevops/azure-devops-exporter/azure-devops-exporter /
USER 1000
ENTRYPOINT ["/azure-devops-exporter"]
