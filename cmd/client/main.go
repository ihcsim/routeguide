package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ihcsim/routeguide"
	pb "github.com/ihcsim/routeguide/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	modeFirehose = "firehose"
	modeRepeatN  = "repeatn"

	defaultServer       = ":8080"
	defaultTimeout      = time.Second * 20
	defaultWait         = time.Second * 3
	defaultMode         = modeRepeatN
	defaultAPI          = routeguide.APIGetFeature
	defaultN            = 10
	defaultServerAddr   = "127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082"
	defaultResolverType = "manual"
)

func main() {
	var (
		server    = flag.String("server", defaultServer, "Name or IP of the target server, including port number.")
		timeout   = flag.Duration("timeout", defaultTimeout, "Default connection timeout")
		mode      = flag.String("mode", defaultMode, "Default mode to start the client in. Supported values: repeatn firehose")
		api       = flag.String("api", defaultAPI, "In the repeatn mode, this indicates the remote API to target")
		n         = flag.Int("n", defaultN, "In the repeatn mode, this is the number of API calls to be repeated")
		enableLB  = flag.Bool("enable-load-balancing", false, "Set to true to enable client-side load balancing")
		serverIPs = flag.String("server-ipv4", defaultServerAddr, "If load balancing is enabled, this is a list of comma-separated server addresses used by the GRPC name resolver")
		resolver  = flag.String("resolver", defaultResolverType, "The resolver to use. Supported values: dns manual")

		opts = []grpc.DialOption{grpc.WithInsecure()}
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	go func() {
		<-stop
		log.Println("[main] stopping")
		cancel()
	}()

	if *enableLB {
		log.Printf("[main] load balancing scheme: %s", roundrobin.Name)
		opts = append(opts, grpc.WithBalancerName(roundrobin.Name))

		rt, err := ParseResolverType(*resolver)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[main] resolver type: %s", rt)
		registerResolver(rt, *serverIPs)
	}

	log.Printf("[main] connecting to server at %s", *server)
	conn, err := grpc.Dial(fmt.Sprintf("%s", *server), opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	grpcClient := pb.NewRouteGuideClient(conn)
	client := routeguide.Client{GRPC: grpcClient}

	log.Printf("[main] running in %s mode", *mode)
	switch strings.ToLower(*mode) {
	case modeFirehose:
		if err := firehose(ctx, client, *timeout); err != nil && err != context.Canceled {
			log.Fatalf("[main] %s", err)
		}
	case modeRepeatN:
		if err := repeatN(ctx, client, *timeout, *api, *n); err != nil && err != context.Canceled {
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

func repeatN(ctx context.Context, client routeguide.Client, timeout time.Duration, api string, n int) error {
	var call func(ctx context.Context) error

	switch strings.ToLower(api) {
	case routeguide.APIGetFeature:
		call = client.GetFeature
	case routeguide.APIListFeatures:
		call = client.ListFeatures
	case routeguide.APIRecordRoute:
		call = client.RecordRoute
	case routeguide.APIRouteChat:
		call = client.RouteChat
	default:
		return fmt.Errorf("Unsupported API %s", api)
	}

	log.Printf("[main] calling %s %d times", api, n)
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
