package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ihcsim/grpc-101/routeguide"
	pb "github.com/ihcsim/grpc-101/routeguide/proto"

	"google.golang.org/grpc"
)

func main() {
	port := "8080"
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Fail to listen for tcp traffic at port %s", port)
	}
	log.Printf("[main] listening at port %s\n", port)

	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	routeGuideServer, err := routeguide.NewServer()
	if err != nil {
		log.Fatalf("Fail to listen for tcp traffic at port %s", port)
	}

	pb.RegisterRouteGuideServer(grpcServer, routeGuideServer)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
