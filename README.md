# grpc-101
This repository contains some GRPC-related programming exercises.

The `routeguide` folder is based on the [basic tutorial](https://grpc.io/docs/tutorials/basic/go.html) found in the official GRPC documentation.

The GRPC server exposes the 4 interfaces as described in the official tutorial to demonstrate unary, client streaming, server streaming and full duplex RPCs.

API            | Description
-------------- | -----------
`GetFeature`   | Obtains the feature at a given position.
`ListFeatures` | Obtains the features available within the given rectangle, via server-side streaming.
`RecordRoute`  | Accepts a stream of points from the client and returns a summary of the route traversed.
`RouteChat`    | Accepts a stream of route notes from the client and returns another stream of notes to the client.

A GRPC client is included to call the APIs exposed by the GRPC server. It can be started in two modes:

Mode       | Description
---------- | -----------
`FIREHOSE` | The client issues random calls to all 4 APIs in a continuous loop. Each loop performs one call.
`REPEATN`  | The client issues sequential calls to all 4 server-side APIs. Each API is called N (default to 20) times. The client will exit once all the APIs are called.

To run the server:
```
# run locally
$ make -C routeguide run_server

# mesh and run on mkube
$ make -C routeguide mesh_server
```

To run the client:
```
$ make -C routeguide run_client

# mesh and run on mkube
$ make -C routeguide mesh_client
```
