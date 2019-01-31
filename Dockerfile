FROM golang:1.11.5-alpine3.8 as builder
ARG PROJECT_PATH=/go/src/github.com/ihcsim/routeguide
WORKDIR ${PROJECT_PATH}
COPY . .
RUN go build -o /rg-server ${PROJECT_PATH}/cmd/server && \
    go build -o /rg-client ${PROJECT_PATH}/cmd/client

FROM alpine:3.8
COPY --from=builder /rg-server /rg-client /
CMD ["/rg-server"]
