FROM golang:1.11.5-alpine3.8 as builder
ARG PROJECT_PATH=/go/src/github.com/ihcsim/routeguide
ARG GRPC_HEALTH_PROBE_VERSION=v0.2.0
WORKDIR ${PROJECT_PATH}
COPY . .
RUN go build -o /rg-server ${PROJECT_PATH}/cmd/server && \
    go build -o /rg-client ${PROJECT_PATH}/cmd/client && \
    wget -qO /grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /grpc_health_probe

FROM alpine:3.8
COPY --from=builder /rg-server /rg-client /grpc_health_probe /
RUN apk add --update --no-cache bash
CMD ["/rg-server"]
