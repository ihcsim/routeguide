package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"

	"github.com/ihcsim/routeguide"
	pb "github.com/ihcsim/routeguide/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	defaultPort     = 8080
	faultPercent    = 0.3
	pathHealthCheck = "/grpc.health.v1.Health/Check"
)

func main() {
	port := flag.Int("port", defaultPort, "Default port to listen on")
	help := flag.Bool("help", false, "Print usage")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("[main] fail to listen for tcp traffic at port %d", port)
	}
	log.Printf("[main] fault percentage: %.1f%%\n", faultPercent*100)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(triggerFaultUnaryInterceptor),
		grpc.StreamInterceptor(triggerFaultStreamInterceptor),
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	hostname = fmt.Sprintf("%s:%d", hostname, *port)
	log.Printf("[main] hostname: %s", hostname)

	grpcServer := grpc.NewServer(opts...)
	routeGuideServer, err := routeguide.NewServer(hostname)
	if err != nil {
		log.Fatalf("[main] fail to listen for tcp traffic at %s", hostname)
	}
	pb.RegisterRouteGuideServer(grpcServer, routeGuideServer)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("routeguide.RouteGuide", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		<-stop
		log.Println("[main] stopping")
		healthServer.Shutdown()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("[main] %s", err)
	}

	log.Println("[main] done")
}

func triggerFaultUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	percent := int(faultPercent * 100)
	if n := rand.Intn(100); info.FullMethod != pathHealthCheck && n > 0 && n <= percent {
		err := routeguide.GetFault(info.FullMethod)
		log.Printf("[interceptor] (fault) %+v\n", err)
		return nil, err
	}

	return handler(ctx, req)
}

func triggerFaultStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	percent := int(faultPercent * 100)
	if n := rand.Intn(100); info.FullMethod != pathHealthCheck && n > 0 && n <= percent {
		err := routeguide.GetFault(info.FullMethod)
		log.Printf("[interceptor] (fault) %+v\n", err)
		return err
	}

	return handler(srv, ss)
}
