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

	if err = httpDo(ctx, req, &pb_response); err != nil {
		return nil, err
	}

	for _, metrics := range pb_response.Metrics {
		data := make([]float64, len(metrics.Values))
		for i := range metrics.Values {
			if metrics.IsAbsent[i] {
				data[i] = math.NaN()
			} else {
				data[i] = metrics.Values[i]
			}
		}
		response[metrics.GetName()] = &Points{
			StartTime: metrics.StartTime,
			StopTime:  metrics.StopTime,
			StepTime:  metrics.StepTime,
		}
	}

	return response, nil
}

// func (t *RenderTarget) String() string {
// 	return t.str
// }

// func NewRenderTarget(seriesList string) *RenderTarget {
// 	return &RenderTarget{
// 		str: seriesList,
// 	}
// }

// func (t *RenderTarget) ApplyFunction(name string, args ...interface{}) *RenderTarget {
// 	tmp := make([]string, len(args)+1)
// 	tmp[0] = t.String()
// 	for i, a := range args {
// 		tmp[i+1] = fmt.Sprintf("%v", a)
// 	}
// 	t.str = fmt.Sprintf("%s(%s)", name, strings.Join(tmp, ","))
// 	return t
// }

// func (t *RenderTarget) ApplyFunctionWithoutSeries(name string, args ...interface{}) *RenderTarget {
// 	tmp := make([]string, len(args))
// 	for i, a := range args {
// 		tmp[i] = fmt.Sprintf("%v", a)
// 	}
// 	t.str = fmt.Sprintf("%s(%s)", name, strings.Join(tmp, ","))
// 	return t
// }

//
// function shortcuts, for code completion
//

// func (t *RenderTarget) SumSeries() *RenderTarget {
// 	return t.ApplyFunction("sumSeries")
// }

// func (t *RenderTarget) ConstantLine(value interface{}) *RenderTarget {
// 	return t.ApplyFunctionWithoutSeries("constantLine", value)
// }
