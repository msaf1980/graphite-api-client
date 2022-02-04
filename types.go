package graphiteapi

import (
	"context"
)

// Query is interface for all api request
type Query interface {
	URL() string
	Request(ctx context.Context) (Response, error)
}

// Response is interface for all api request response types
type Response interface {
	Unmarshal([]byte) error
}

// RenderQuery is used to build `/render/` query
type RenderQuery struct {
	Base          string // base url of graphite server
	User          string // user
	Password      string // password
	Targets       []string
	From          string
	Until         string
	MaxDataPoints int
}

type Series struct {
	Target    string
	StartTime int32
	StopTime  int32
	StepTime  int32
	Values    []float64
}
