package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/ihcsim/grpc-101/routeguide"
	pb "github.com/ihcsim/grpc-101/routeguide/proto"

	"google.golang.org/grpc"
)

const (
	defaultPort         = "8080"
	defaultFaultPercent = 0.3
)

var faultPercent = defaultFaultPercent

func main() {
	port, exist := os.LookupEnv("SERVER_PORT")
	if !exist {
		port = defaultPort
	}

	var err error
	faultPercentEnv, exist := os.LookupEnv("FAULT_PERCENT")
	if exist {
		faultPercent, err = strconv.ParseFloat(faultPercentEnv, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("[main] fail to listen for tcp traffic at port %s", port)
	}
	log.Printf("[main] listening at port %s\n", port)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(triggerFaultUnaryInterceptor),
		grpc.StreamInterceptor(triggerFaultStreamInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)
	routeGuideServer, err := routeguide.NewServer(faultPercent)
	if err != nil {
		log.Fatalf("[main] fail to listen for tcp traffic at port %s", port)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	go func() {
		<-stop
		log.Println("[main] stopping")
		grpcServer.GracefulStop()
	}()

	pb.RegisterRouteGuideServer(grpcServer, routeGuideServer)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("[main] %s", err)
	}

	log.Println("[main] done")
}

func triggerFaultUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if n := rand.Float64(); n <= faultPercent {
		err := routeguide.GetFault(info.FullMethod)
		log.Printf("[interceptor] (fault) %+v\n", err)
		return nil, err
	}

	return handler(ctx, req)
}

func triggerFaultStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if n := rand.Float64(); n <= faultPercent {
		err := routeguide.GetFault(info.FullMethod)
		log.Printf("[interceptor] (fault) %+v\n", err)
		return err
	}

	return handler(srv, ss)
}
