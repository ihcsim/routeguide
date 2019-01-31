package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ihcsim/routeguide"
	pb "github.com/ihcsim/routeguide/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	modeFirehose = "firehose"
	modeRepeatN  = "repeatn"

	apiGetFeature   = "getfeature"
	apiListFeatures = "listfeatures"
	apiRecordRoute  = "recordroute"
	apiRouteChat    = "routechat"

	defaultAddr    = ""
	defaultPort    = "8080"
	defaultTimeout = time.Second * 20
	defaultWait    = time.Second * 3
	defaultMode    = modeRepeatN
	defaultAPI     = apiGetFeature
	defaultN       = 10
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
	switch strings.ToLower(mode) {
	case modeFirehose:
		if err := firehose(ctx, client, timeout); err != nil && err != context.Canceled {
			log.Fatalf("[main] %s", err)
		}
	case modeRepeatN:
		if err := repeatN(ctx, client, timeout); err != nil && err != context.Canceled {
			log.Fatalf("[main] %s", err)
		}
	default:
		log.Fatalf("[main] unknown mode %s", mode)
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
				if err := client.GetFeature(ctx); err != nil {
					if !isInjectedFault(err) {
						return err
					}
					log.Println(err)
				}
			} else if n < 5 && n >= 3 {
				if err := client.ListFeatures(ctx); err != nil {
					if !isInjectedFault(err) {
						return err
					}
					log.Println(err)
				}
			} else if n < 7 && n >= 5 {
				if err := client.RecordRoute(ctx); err != nil {
					if !isInjectedFault(err) {
						return err
					}
					log.Println(err)
				}
			} else {
				if err := client.RouteChat(ctx); err != nil {
					if !isInjectedFault(err) {
						return err
					}
					log.Println(err)
				}
			}

			time.Sleep(defaultWait)
		}
	}
}

func repeatN(ctx context.Context, client routeguide.Client, timeout time.Duration) error {
	var call func(ctx context.Context) error

	api, ok := os.LookupEnv("REMOTE_API")
	if !ok {
		api = defaultAPI
	}

	var err error
	n := defaultN
	nEnv, ok := os.LookupEnv("MAX_REPEAT")
	if ok {
		n, err = strconv.Atoi(nEnv)
		if err != nil {
			return err
		}
	}

	switch strings.ToLower(api) {
	case apiGetFeature:
		call = client.GetFeature
	case apiListFeatures:
		call = client.ListFeatures
	case apiRecordRoute:
		call = client.RecordRoute
	case apiRouteChat:
		call = client.RouteChat
	default:
		return fmt.Errorf("Unsupported API %s", api)
	}

	log.Printf("calling %s %d times", api, n)
	for i := 0; i < n; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := call(ctx); err != nil {
			if !isInjectedFault(err) {
				return err
			}
			log.Println(err)
		}

		time.Sleep(defaultWait)
	}

	return nil
}

func isInjectedFault(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}

	return s.Code() == codes.Unavailable && strings.Contains(s.Message(), routeguide.FaultMsg)
}
