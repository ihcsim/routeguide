package routeguide

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"

	pb "github.com/ihcsim/routeguide/proto"
)

// NewServer returns a new route guide server that exposes 4 GRPC APIs.
func NewServer(faultPercent float64) (pb.RouteGuideServer, error) {
	r := &routeGuideServer{
		savedFeatures: []*pb.Feature{},
		routeNotes:    make(map[string][]*pb.RouteNote),
		faultPercent:  faultPercent,
	}

	err := json.Unmarshal(featuresData, &r.savedFeatures)
	if err != nil {
		return nil, err
	}

	return r, err
}

type routeGuideServer struct {
	savedFeatures []*pb.Feature
	routeNotes    map[string][]*pb.RouteNote
	mutex         sync.Mutex
	faultPercent  float64
}

// GetFeature obtains the feature at a given position.
func (r *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	log.Printf("[GetFeature] (req) %+v\n", point)
	for _, feature := range r.savedFeatures {
		if proto.Equal(feature.Location, point) {
			log.Printf("[GetFeature] (resp) %+v\n", feature)
			return feature, nil
		}
	}

	return &pb.Feature{}, nil
}

// ListFeatures obtains the features available within the given rectangle.
// The results are streamed rather than returned immediately as the rectangle
// may cover a large area and contain a large number of features.
func (r *routeGuideServer) ListFeatures(rectangle *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	log.Printf("[ListFeatures] (req) %+v\n", rectangle)
	for _, feature := range r.savedFeatures {
		if inRange(feature, rectangle) {
			log.Printf("[ListFeatures] (resp) %+v\n", feature)
			if err := stream.Send(feature); err != nil {
				return err
			}
		}
	}

	return nil
}

// RecordRoute accepts a stream of points on a route being traversed, returning a
// route summary when traversal is completed.
func (r *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var (
		summary   = &pb.RouteSummary{}
		startTime = time.Now()
		lastPoint *pb.Point
	)

	for {
		point, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		log.Printf("[RecordRoute] (req) %+v\n", point)
		summary.PointCount++

		if _, err := r.GetFeature(context.Background(), point); err != nil {
			return err
		}
		summary.FeatureCount++

		if lastPoint != nil {
			summary.Distance += dist(point, lastPoint)
		}

		lastPoint = point
	}

	summary.ElapsedTime = int32(time.Now().Sub(startTime).Seconds())
	log.Printf("[RecordRoute] (resp) %+v\n", summary)
	if err := stream.SendAndClose(summary); err != nil {
		return err
	}

	return nil
}

// RouteChat accepts a stream of route notes sent while a route is being traversed,
// while receiving other route notes.
func (r *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
	for {
		note, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		log.Printf("[RouteChat] (req) %+v\n", note)

		key := fmt.Sprintf("(%d,%d)", note.Location.GetLatitude(), note.Location.GetLongitude())

		// clone the server's route notes to not block other clients
		r.mutex.Lock()
		note.Message = strings.Replace(note.Message, "ack=0", "ack=1", -1)
		r.routeNotes[key] = append(r.routeNotes[key], note)
		clone := make([]*pb.RouteNote, len(r.routeNotes[key]))
		copy(clone, r.routeNotes[key])
		r.mutex.Unlock()

		for _, note := range clone {
			log.Printf("[RouteChat] (resp) %+v\n", note)
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}

func inRange(feature *pb.Feature, rectangle *pb.Rectangle) bool {
	var (
		top    = math.Max(float64(rectangle.Lo.GetLatitude()), float64(rectangle.Hi.GetLatitude()))
		bottom = math.Min(float64(rectangle.Lo.GetLatitude()), float64(rectangle.Hi.GetLatitude()))
		right  = math.Max(float64(rectangle.Lo.GetLongitude()), float64(rectangle.Hi.GetLongitude()))
		left   = math.Min(float64(rectangle.Lo.GetLongitude()), float64(rectangle.Hi.GetLongitude()))
	)

	return float64(feature.Location.GetLatitude()) <= top && float64(feature.Location.GetLatitude()) >= bottom && float64(feature.Location.GetLongitude()) <= right && float64(feature.Location.GetLongitude()) >= left
}

func dist(a *pb.Point, b *pb.Point) int32 {
	var (
		operand1 = math.Pow(float64(a.GetLatitude())-float64(b.GetLatitude()), 2)
		operand2 = math.Pow(float64(a.GetLongitude())-float64(b.GetLongitude()), 2)
	)
	return int32(math.Sqrt(operand1 - operand2))
}
