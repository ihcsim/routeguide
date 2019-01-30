package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/ihcsim/grpc-101/routeguide"
	pb "github.com/ihcsim/grpc-101/routeguide/proto"

	"google.golang.org/grpc"
)

const (
	modeFirehose = "firehose"
	modeRepeatN  = "repeatN"

	defaultAddr    = ""
	defaultPort    = "8080"
	defaultTimeout = time.Second * 20
	defaultMode    = modeRepeatN
)

func main() {
	addr, exist := os.LookupEnv("SERVER_HOST")
	if !exist {
		addr = defaultAddr
	}

	port, exist := os.LookupEnv("SERVER_PORT")
	if !exist {
		port = defaultPort
	}

	var err error
	timeout := defaultTimeout
	timeoutEnv, exist := os.LookupEnv("GRPC_TIMEOUT")
	if exist {
		timeout, err = time.ParseDuration(timeoutEnv)
		if err != nil {
			log.Fatal(err)
		}
	}

	mode, exist := os.LookupEnv("MODE")
	if !exist {
		mode = defaultMode
	}

	var (
		opts   = []grpc.DialOption{grpc.WithInsecure()}
		client = routeguide.Client{}
	)

	log.Printf("[main] connecting to server at %s:%s", addr, port)
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", addr, port), opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	grpcClient := pb.NewRouteGuideClient(conn)
	client.GRPC = grpcClient

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	go func() {
		<-stop
		log.Println("[main] stopping")
		cancel()
	}()

	log.Printf("[main] running in %s mode", mode)
	switch mode {
	case modeFirehose:
		if err := firehose(ctx, client, timeout); err != nil && err != context.Canceled {
			log.Fatalf("[main] %s", err)
		}
	case modeRepeatN:
		if err := repeatN(ctx, client, timeout); err != nil && err != context.Canceled {
			log.Fatal("[main] %s", err)
		}
	}
	log.Println("[main] finished")
}

func firehose(ctx context.Context, client routeguide.Client, timeout time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// each API has a 25% of being invoked
			n := rand.Intn(10)
			if n < 3 {
				if err := client.GetFeature(ctx); err != nil && !isTriggeredFault(err, "GetFeature") {
					fmt.Println("test 1")
					return err
				}
			} else if n < 5 && n >= 3 {
				if err := client.ListFeatures(ctx); err != nil && !isTriggeredFault(err, "ListFeatures") {
					return err
				}
			} else if n < 7 && n >= 5 {
				if err := client.RecordRoute(ctx); err != nil && !isTriggeredFault(err, "RecordRoute") {
					return err
				}
			} else {
				if err := client.RouteChat(ctx); err != nil && !isTriggeredFault(err, "RouteChat") {
					return err
				}
			}

			time.Sleep(time.Second * 3)
		}
	}
}

func repeatN(ctx context.Context, client routeguide.Client, timeout time.Duration) error {
	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := client.GetFeature(ctx); err != nil && !isTriggeredFault(err, "GetFeature") {
			return err
		}
		time.Sleep(time.Second * 3)
	}

	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := client.ListFeatures(ctx); err != nil && !isTriggeredFault(err, "ListFeatures") {
			return (err)
		}

		time.Sleep(time.Second * 3)
	}

	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := client.RecordRoute(ctx); err != nil && !isTriggeredFault(err, "RecordRoute") {
			return (err)
		}

		time.Sleep(time.Second * 3)
	}

	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := client.RouteChat(ctx); err != nil && !isTriggeredFault(err, "RouteChat") {
			return (err)
		}

		time.Sleep(time.Second * 3)
	}

	return nil
}

func isTriggeredFault(err error, api string) bool {
	return err == routeguide.GetFault(api)
}
