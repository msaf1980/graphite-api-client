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

// DataPoint describes concrete point of time series.
type DataPoint struct {
	Value     float64
	Timestamp int64
}

type Series struct {
	Target     string
	DataPoints []DataPoint
}
