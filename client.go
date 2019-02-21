package routeguide

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	pb "github.com/ihcsim/routeguide/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	APIGetFeature   = "getfeature"
	APIListFeatures = "listfeatures"
	APIRecordRoute  = "recordroute"
	APIRouteChat    = "routechat"

	unknownServerName = "unknown"
)

// Client knows how to communicate with the GRPC server.
type Client struct {
	GRPC pb.RouteGuideClient
}

// GetFeature interacts with the GetFeature API on the GRPC server.
func (c *Client) GetFeature(ctx context.Context) error {
	var (
		header metadata.MD
		point  = randPoint()
	)
	log.Printf("[GetFeature] (req) %+v\n", point)

	feature, err := c.GRPC.GetFeature(ctx, point, grpc.Header(&header))
	if err != nil {
		return err
	}

	output(APIGetFeature, header, feature)
	return nil
}

// ListFeatures interacts with the ListFeatures API on the GRPC server.
func (c *Client) ListFeatures(ctx context.Context) error {
	rectangle := &pb.Rectangle{
		Lo: randPoint(),
		Hi: randPoint(),
	}
	log.Printf("[ListFeatures] (req) %+v, %+v\n", rectangle.Lo, rectangle.Hi)

	stream, err := c.GRPC.ListFeatures(ctx, rectangle)
	if err != nil {
		return err
	}

	for {
		feature, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		header, err := stream.Header()
		if err != nil {
			return err
		}

		output(APIListFeatures, header, feature)
	}

	return nil
}

// RecourdRoute interacts with the RecordRoute API on the GRPC server.
func (c *Client) RecordRoute(ctx context.Context) error {
	stream, err := c.GRPC.RecordRoute(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		point := randPoint()
		log.Printf("[RecordRoute] (req) %+v\n", point)

		if err := stream.Send(point); err != nil {
			return nil
		}
	}

	summary, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	header, err := stream.Header()
	if err != nil {
		return err
	}

	output(APIRecordRoute, header, summary)
	return nil
}

// RouteChat interacts with the RouteChat API on the GRPC server.
func (c *Client) RouteChat(ctx context.Context) error {
	stream, err := c.GRPC.RouteChat(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		var (
			point = randPoint()
			msg   = fmt.Sprintf("[%s] ack=0 msg='message #%d'", time.Now(), i)
			note  = &pb.RouteNote{
				Location: point,
				Message:  msg,
			}
		)
		log.Printf("[RouteChat] (req) %+v\n", note)

		if err := stream.Send(note); err != nil {
			return err
		}

		resp, err := stream.Recv()
		if err != nil {
			return err
		}

		header, err := stream.Header()
		if err != nil {
			return err
		}

		output(APIRouteChat, header, resp)
	}

	return nil
}

func output(api string, metadata metadata.MD, content proto.Message) {
	server := unknownServerName
	if serverName, ok := metadata["server"]; ok && len(serverName) > 0 {
		server = serverName[0]
	}
	log.Printf("[%s] {resp) (server=%s) %+v\n", strings.Title(api), server, content)
}
