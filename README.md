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

A GRPC client is included to initiate continuous random calls to the APIs.

To run the server:
```
$ make -C routeguide run_server
```

To run the client:
```
$ make -C routeguide run_client
```
