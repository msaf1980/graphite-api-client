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
func (q *RenderQuery) Request(ctx context.Context) (RenderResponse, error) {
	var req *http.Request
	var err error

	response := make(RenderResponse)
	pb_response := protov2.MultiFetchResponse{}

	if req, err = httpNewRequest("GET", q.URL().String(), nil); err != nil {
		return nil, err
	}

	if len(q.User) > 0 {
		q.SetBasicAuth(q.User, q.Password)
	}

	if err = httpDo(ctx, req, &pb_response); err != nil {
		return nil, err
	}

	for _, metrics := range pb_response.Metrics {
		response[metrics.GetName()] = &Points{
			StartTime: metrics.StartTime,
			StopTime:  metrics.StopTime,
			StepTime:  metrics.StepTime,
			Values:    metrics.Values,
			IsAbsent:  metrics.IsAbsent,
		}
	}

	return response, nil
}

// GetLastNonNullValue searches for the latest non null value, and skips at most maxNullPoints.
// If the last maxNullPoints values are all absent, returns absent
func GetLastNonNullValue(pp *Points, maxNullPoints int) (t int32, v float64, absent bool) {
	l := len(pp.Values)

	if l == 0 {
		// there is values, we should return absent
		v = 0
		t = pp.StopTime
		absent = true
		return t, v, absent
	}

	for i := 0; i < maxNullPoints && i < l; i++ {
		if pp.IsAbsent[l-1-i] {
			continue
		}
		v = pp.Values[l-1-i]
		t = pp.StopTime - int32(i)*pp.StepTime
		absent = false
		return t, v, absent
	}

	// if we get here, there are two cases
	//   * maxNullPoints == 0, we didn't even enter the loop above
	//   * maxNullPoints > 0, but we didn't find a non-null point in the loop
	// in both cases, we return the last point's info
	v = pp.Values[l-1]
	t = pp.StopTime
	absent = math.IsNaN(pp.Values[l-1])
	return t, v, absent
}
