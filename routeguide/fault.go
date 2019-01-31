package routeguide

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const FaultMsg = "grpc server unavailable"

// GetFault returns the fault object associated with the given API.
func GetFault(api string) error {
	return status.Errorf(codes.Unavailable, fmt.Sprintf("%s. path: %s", FaultMsg, api))
}
