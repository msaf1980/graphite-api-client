package graphiteapi

import (
	"context"
	"math"
	"net/http"
	"net/url"
	"strconv"

	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
)

// NewRenderQuery returns a RenderQuery instance
func NewRenderQuery(base, from, until string, targets []string) *RenderQuery {
	q := &RenderQuery{
		Base:          base,
		Targets:       targets,
		From:          from,
		Until:         until,
		MaxDataPoints: 0,
	}
	return q
}

func (q *RenderQuery) SetBasicAuth(username, password string) {
	q.User = username
	q.Password = password
}

func (q *RenderQuery) SetFrom(from string) *RenderQuery {
	q.From = from
	return q
}

func (q *RenderQuery) SetUntil(until string) *RenderQuery {
	q.Until = until
	return q
}

func (q *RenderQuery) SetTargets(targets []string) *RenderQuery {
	q.Targets = targets
	return q
}

func (q *RenderQuery) AddTarget(target string) *RenderQuery {
	q.Targets = append(q.Targets, target)
	return q
}

func (q *RenderQuery) SetMaxDataPoints(maxDataPoints int) *RenderQuery {
	q.MaxDataPoints = maxDataPoints
	return q
}

// URL implements Query interface
func (q *RenderQuery) URL() *url.URL {
	u, _ := url.Parse(q.Base + "/render/")
	v := url.Values{}

	// force set format to protobuf
	v.Set("format", "protobuf")

	for _, target := range q.Targets {
		v.Add("target", target)
	}

	if q.From != "" {
		v.Set("from", q.From)
	}

	if q.Until != "" {
		v.Set("until", q.Until)
	}

	if q.MaxDataPoints != 0 {
		v.Set("maxDataPoints", strconv.Itoa(q.MaxDataPoints))
	}

	u.RawQuery = v.Encode()

	return u
}

// Request implements Query interface
func (q *RenderQuery) Request(ctx context.Context) ([]Series, error) {
	var req *http.Request
	var err error

	pb_response := protov2.MultiFetchResponse{}

	if req, err = httpNewRequest("GET", q.URL().String(), nil); err != nil {
		return nil, err
	}

	if len(q.User) > 0 {
		req.SetBasicAuth(q.User, q.Password)
	}

	if err = httpDo(ctx, req, &pb_response); err != nil {
		return nil, err
	}

	response := make([]Series, len(pb_response.Metrics))

	i := 0
	for _, metrics := range pb_response.Metrics {
		for i := range metrics.Values {
			if metrics.IsAbsent[i] {
				metrics.Values[i] = math.NaN()
			}
		}
		response[i] = Series{
			Target:    metrics.Name,
			StartTime: metrics.StartTime,
			StopTime:  metrics.StopTime,
			StepTime:  metrics.StepTime,
			Values:    metrics.Values,
		}
		i++
	}

	return response, nil
}
