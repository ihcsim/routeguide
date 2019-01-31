package routeguide

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/ihcsim/routeguide/proto"
)

// Client knows how to communicate with the GRPC server.
type Client struct {
	GRPC pb.RouteGuideClient
}

// GetFeature interacts with the GetFeature API on the GRPC server.
func (c *Client) GetFeature(ctx context.Context) error {
	point := randPoint()
	log.Printf("[GetFeature] (req) %+v\n", point)

	feature, err := c.GRPC.GetFeature(ctx, point)
	if err != nil {
		return err
	}

	log.Printf("[GetFeature] (resp) %+v\n", feature)
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

		log.Printf("[ListFeatures] (resp) %+v\n", feature)
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

	log.Printf("[RecordRoute] (resp) %+v\n", summary)
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

		log.Printf("[RouteChat] {resp) %+v\n", resp)
	}

	return nil
}
