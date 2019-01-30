package routeguide

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const faultMsg = "[%s] (fault) grpc server unavailable"

// GetFault returns the fault object associated with the given API.
func GetFault(api string) error {
	return status.Errorf(codes.Unavailable, faultMsg, api)
}
