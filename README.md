# routeguide
This repository contains a gRPC programming exercise which is based on the [basic tutorial](https://grpc.io/docs/tutorials/basic/go.html) found in the official gRPC documentation.

The objective of this exercise is to explore the following gRPC features:

* Unary RPC
* Client streaming RPC
* Server streaming RPC
* Full duplex RPC
* Interceptors (to return faulty responses)
* Health checks
* Load balancing
* gRPC Metadata

It is developed using the following software:

* Minikube v0.33.1
* Go 1.10.2
* Protoc 3.6.1
* dep v0.5.0
* Linkerd2 edge-19.1.3

## About The Applications
The gRPC server exposes the 4 interfaces:

API            | Description
-------------- | -----------
`GetFeature`   | Obtains the feature at a given position.
`ListFeatures` | Obtains the features available within the given rectangle, via server-side streaming.
`RecordRoute`  | Accepts a stream of points from the client and returns a summary of the route traversed.
`RouteChat`    | Accepts a stream of route notes from the client and returns another stream of notes to the client.

It also uses interceptors to return faulty responses.

The gRPC client is used to make RPC calls to the server APIs. It can be started in two modes:

Mode       | Description
---------- | -----------
`FIREHOSE` | The client issues random calls to all 4 APIs in an infinite loop.
`REPEATN`  | The client issues N calls to the selected API. The client will exit once it repeated N calls.


To run the server and client locally:
```
$ make server
$ make client
```
Ensure that your local hostname resolves to 127.0.0.1. If not, add it to your `/etc/hosts` file.

To try out the client-side round robin load balancing, start multiple instances of the servers:
```
$ SERVER_PORT=8080 make server &
$ SERVER_PORT=8081 make server &
$ SERVER_PORT=8082 make server &
$ make client
```
Notice the logs of the servers as requests are received from the client.

To build the Dockerfile on Minikube:
```
$ make image
```

To meshed the server and client on Minikube with the Linkerd 2 proxy:
```
$ make mesh
```

To build the protobuf files:
```
$ make proto
```

To delete all resources from Minikube:
```
$ make clean
```
