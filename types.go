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
	Targets       []string
	From          string
	Until         string
	MaxDataPoints int
}

type Points struct {
	StartTime int32
	StopTime  int32
	StepTime  int32
	Values    []float64
	IsAbsent  []bool
}

// RenderResponse is response of `/render/` query
type RenderResponse map[string]*Points
