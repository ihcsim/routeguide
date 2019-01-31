# grpc-101
This repository contains a GRPC programming exercise which is based on the [basic tutorial](https://grpc.io/docs/tutorials/basic/go.html) found in the official GRPC documentation.

This exercise is developed using the following software:

* Minikube v0.33.1
* Go 1.10.2

The GRPC server exposes the 4 interfaces as described in the official tutorial to demonstrate unary, client streaming, server streaming and full duplex RPCs.

API            | Description
-------------- | -----------
`GetFeature`   | Obtains the feature at a given position.
`ListFeatures` | Obtains the features available within the given rectangle, via server-side streaming.
`RecordRoute`  | Accepts a stream of points from the client and returns a summary of the route traversed.
`RouteChat`    | Accepts a stream of route notes from the client and returns another stream of notes to the client.

The GRPC server uses interceptors to return faulty responses. The `FAULT_PERCENT` environment variable can be used to adjust percentage of requests to be failed by the interceptors.

A GRPC client is included to call the APIs exposed by the GRPC server. It can be started in two modes:

Mode       | Description
---------- | -----------
`FIREHOSE` | The client issues random calls to all 4 APIs in a continuous loop. Each loop performs one call.
`REPEATN`  | The client issues sequential calls to all 4 server-side APIs. Each API is called N (default to 20) times. The client will exit once all the APIs are called.


To run the server and client locally:
```
$ make run_server
$ make run_client
```

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
