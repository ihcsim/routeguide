FROM golang:1.11.5-alpine3.8

ARG PROJECT_PATH=/go/src/github.com/ihcsim/grpc-101/routeguide
WORKDIR ${PROJECT_PATH}
COPY . .
RUN go build -o route-guide-server ${PROJECT_PATH}/cmd/server && \
    go build -o route-guide-client ${PROJECT_PATH}/cmd/client
