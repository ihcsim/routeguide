# routeguide
This repository contains a GRPC programming exercise which is based on the [basic tutorial](https://grpc.io/docs/tutorials/basic/go.html) found in the official GRPC documentation.

This exercise is developed using the following software:

* Minikube v0.33.1
* Go 1.10.2
* Protoc 3.6.1
* dep v0.5.0
* Linkerd2 edge-19.1.3

The objective is to explore the following GRPC features:

* Unary RPC
* Client streaming RPC
* Server streaming RPC
* Full duplex RPC
* Interceptors (to return faulty responses)
* Health checks
* Lad balancing

The GRPC server exposes the 4 interfaces as described in the official tutorial:

API            | Description
-------------- | -----------
`GetFeature`   | Obtains the feature at a given position.
`ListFeatures` | Obtains the features available within the given rectangle, via server-side streaming.
`RecordRoute`  | Accepts a stream of points from the client and returns a summary of the route traversed.
`RouteChat`    | Accepts a stream of route notes from the client and returns another stream of notes to the client.

The GRPC server uses interceptors to return faulty responses. The `-fault-percent` flag can be used to adjust percentage of requests to be failed by the interceptors.

A GRPC client is included to call the APIs exposed by the GRPC server. It can be started in two modes:

Mode       | Description
---------- | -----------
`FIREHOSE` | The client issues random calls to all 4 APIs in a continuous loop. Each loop performs one call.
`REPEATN`  | The client issues sequential calls to all 4 server-side APIs. Each API is called N (default to 20) times. The client will exit once all the APIs are called.


To run the server and client locally:
```
$ make server
$ make client
```

To try the client-side round robin load balancing, start multiple instances of the servers:
```
$ SERVER_PORT=8080 make server
$ SERVER_PORT=8081 make server
$ SERVER_PORT=8082 make server
$ make client
```
Notice the logs of the servers as requests are received from the client.

To build the Dockerfile on Minikube:
```
$ make image
```

To deploy the server and client to Minikube:
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
