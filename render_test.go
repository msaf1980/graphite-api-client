package graphiteapi

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func compareSeries(t *testing.T, res, want []Series) {
	maxLen := len(res)
	if maxLen < len(want) {
		maxLen = len(want)
	}
	for i := 0; i < maxLen; i++ {
		if i >= len(res) {
			t.Errorf("- [%d] = %+v", i, want[i])
		} else if i >= len(want) {
			t.Errorf("+ [%d] = %+v", i, res[i])
		} else {
			maxWidth := len(res[i].DataPoints)
			if maxWidth < len(want[i].DataPoints) {
				maxWidth = len(want[i].DataPoints)
			}
			for j := 0; j < maxWidth; j++ {
				if j >= len(res[i].DataPoints) {
					t.Errorf("- [%d][%d] = %+v", i, j, want[i].DataPoints[j])
				} else if j >= len(want[i].DataPoints) {
					t.Errorf("+ [%d][%d] = %+v", i, j, res[i].DataPoints[j])
				} else if want[i].DataPoints[j].Value == res[i].DataPoints[j].Value {
					if want[i].DataPoints[j].Timestamp != res[i].DataPoints[j].Timestamp {
						t.Errorf("- [%d][%d] = %+v", i, j, want[i].DataPoints[j])
						t.Errorf("+ [%d][%d] = %+v", i, j, res[i].DataPoints[j])
					}
				} else if !math.IsNaN(want[i].DataPoints[j].Value) || !math.IsNaN(res[i].DataPoints[j].Value) {
					t.Errorf("- [%d][%d] = %+v", i, j, want[i].DataPoints[j])
					t.Errorf("+ [%d][%d] = %+v", i, j, res[i].DataPoints[j])
				}
			}
		}
	}
}

func TestNewClientFromString(t *testing.T) {
	urlString := "http://domain.tld/path"
	testRequest := NewRenderQuery(urlString, "-5min", "now", []string{"TEST.*", "TEST2.a*"}, 0)
	shouldUrl := "http://domain.tld/path/render/?format=json&from=-5min&target=TEST.%2A&target=TEST2.a%2A&until=now"
	gotUrl := testRequest.URL().String()
	if shouldUrl != gotUrl {
		t.Errorf("Resulting URL is %v, \n but should be %v", gotUrl, shouldUrl)
	}
}

func makeRenderTest(t *testing.T, query *RenderQuery, expectedQuery, result string, series []Series) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parsedQuery, _ := url.ParseQuery(expectedQuery)
		if !reflect.DeepEqual(r.URL.Query(), parsedQuery) {
			t.Errorf("Expected query is %+v but %+v got", parsedQuery, r.URL.Query())
		}
		if r.URL.Path != "/render/" {
			t.Errorf("Path should be `/render/` but %s found", r.URL.Path)
		}
		fmt.Fprintln(w, result)
	}))
	defer ts.Close()

	base := "http://" + ts.Listener.Addr().String()
	q := NewRenderQuery(base, query.From, query.Until, query.Targets, query.MaxDataPoints)
	res, err := q.Request(context.Background())
	if err == nil {
		compareSeries(t, res, series)
	} else {
		t.Error(err)
	}
}

var renderTestCases = []struct {
	Query         *RenderQuery
	ExpectedQuery string
	Result        string
	Series        []Series
}{
	{
		Query:         &RenderQuery{Targets: []string{"main1", "main2"}},
		ExpectedQuery: "format=json&target=main1&target=main2",
		Result:        "[{\"target\": \"main\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}]",
		Series: []Series{
			{
				Target: "main",
				DataPoints: []DataPoint{
					{Value: 1.1, Timestamp: 1468339853},
					{Value: 2.0, Timestamp: 1468339854},
					{Value: math.NaN(), Timestamp: 1468339855},
				},
			},
		},
	},
	{
		Query:         &RenderQuery{From: "1468339853", Until: "1468339854"},
		ExpectedQuery: "format=json&from=1468339853&until=1468339854",
		Result:        "[{\"target\": \"main\", \"datapoints\": [[1.1, 1468339853], [2.0, 1468339854], [null, 1468339855]]}]",
		Series: []Series{
			{
				Target: "main",
				DataPoints: []DataPoint{
					{Value: 1.1, Timestamp: 1468339853},
					{Value: 2.0, Timestamp: 1468339854},
					{Value: math.NaN(), Timestamp: 1468339855},
				},
			},
		},
	},
	{
		Query:         &RenderQuery{MaxDataPoints: 1},
		ExpectedQuery: "format=json&maxDataPoints=1",
		Result:        "[]",
		Series:        []Series{},
	},
}

func TestGraphiteClient_Query(t *testing.T) {
	for i, tc := range renderTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			makeRenderTest(t, tc.Query, tc.ExpectedQuery, tc.Result, tc.Series)
		})
	}
}
