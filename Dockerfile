FROM golang:1.11 as build

# golang deps
WORKDIR /tmp/app/
COPY ./src/glide.yaml /tmp/app/
COPY ./src/glide.lock /tmp/app/
RUN curl https://glide.sh/get | sh \
    && glide install

WORKDIR /go/src/azure-devops-exporter/src
COPY ./src /go/src/azure-devops-exporter/src
RUN mkdir /app/ \
    && cp -a /tmp/app/vendor ./vendor/ \
    && go build -o /app/azure-devops-exporter

#############################################
# FINAL IMAGE
#############################################
FROM alpine
RUN apk add --no-cache \
        libc6-compat \
    	ca-certificates
COPY --from=build /app/ /app/
USER 1000

CMD ["/app/azure-devops-exporter"]
