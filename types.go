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
	Data      []float64
}

// RenderResponse is response of `/render/` query
type RenderResponse map[string]*Points

// QueryTarget represents a `target=` arg in `/render/` query
// type RenderTarget struct {
// 	str string
// }

// GetLastNonNullValue searches for the latest non null value, and skips at most maxNullPoints.
// If the last maxNullPoints values are all absent, returns absent
// func GetLastNonNullValue(m *RenderResponse, maxNullPoints int) (v float64, t int32, absent bool) {
// 	l := len(m.Values)

// 	if l == 0 {
// 		// there is values, we should return absent
// 		v = 0
// 		t = m.StopTime
// 		absent = true
// 		return v, t, absent
// 	}

// 	for i := 0; i < maxNullPoints && i < l; i++ {
// 		if m.IsAbsent[l-1-i] {
// 			continue
// 		}
// 		v = m.Values[l-1-i]
// 		t = m.StopTime - int32(i)*m.StepTime
// 		absent = false
// 		return v, t, absent
// 	}

// 	// if we get here, there are two cases
// 	//   * maxNullPoints == 0, we didn't even enter the loop above
// 	//   * maxNullPoints > 0, but we didn't find a non-null point in the loop
// 	// in both cases, we return the last point's info
// 	v = m.Values[l-1]
// 	t = m.StopTime
// 	absent = m.IsAbsent[l-1]
// 	return v, t, absent
// }
