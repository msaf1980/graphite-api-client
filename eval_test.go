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

	"github.com/msaf1980/graphite-api-client/types"
)

func Test_splitEval(t *testing.T) {
	tests := []struct {
		eval        string
		wantTarget  string
		wantEvalCmp EvalCmp
		wantV       float64
		wantErr     bool
	}{
		{
			eval:        " scale(TEST.*, 60)>0",
			wantTarget:  "scale(TEST.*, 60)",
			wantEvalCmp: EvalGt,
			wantV:       0.0,
			wantErr:     false,
		},
		{
			eval:        "movingAverage(TEST.*, 5) >= 2",
			wantTarget:  "movingAverage(TEST.*, 5)",
			wantEvalCmp: EvalGe,
			wantV:       2.0,
			wantErr:     false,
		},
		{
			eval:        "movingAverage(TEST.*, 5)<2",
			wantTarget:  "movingAverage(TEST.*, 5)",
			wantEvalCmp: EvalLt,
			wantV:       2.0,
			wantErr:     false,
		},
		{
			eval:        "movingAverage(TEST.*, 5) <=3.1",
			wantTarget:  "movingAverage(TEST.*, 5)",
			wantEvalCmp: EvalLe,
			wantV:       3.1,
			wantErr:     false,
		},
		{
			eval:        "movingAverage(TEST.*, 5) ==3.1",
			wantTarget:  "movingAverage(TEST.*, 5)",
			wantEvalCmp: EvalEq,
			wantV:       3.1,
			wantErr:     false,
		},
		{
			eval:        " <=3.1",
			wantEvalCmp: EvalEq,
			wantV:       0.0,
			wantErr:     true,
		},
		{
			eval:        "movingAverage(TEST.*, 5) 3.1",
			wantEvalCmp: EvalEq,
			wantV:       0.0,
			wantErr:     true,
		},
		{
			eval:        "movingAverage(TEST.*, 5) == ",
			wantEvalCmp: EvalEq,
			wantV:       0.0,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.eval, func(t *testing.T) {
			gotTarget, gotEvalCmp, gotV, err := splitEval(tt.eval)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitEval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTarget, tt.wantTarget) {
				t.Errorf("splitEval() got = %v, want %v", gotTarget, tt.wantTarget)
			}
			if gotEvalCmp != tt.wantEvalCmp {
				t.Errorf("splitEval() got1 = %v, want %v", gotEvalCmp, tt.wantEvalCmp)
			}
			if gotV != tt.wantV {
				t.Errorf("splitEval() got2 = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func compareEvalResult(t *testing.T, res, want []types.EvalResult) {
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
			if res[i].T != want[i].T || res[i].Success != want[i].Success || res[i].IsAbsent != want[i].IsAbsent {
				t.Errorf("- [%d] = %+v", i, want[i])
				t.Errorf("+ [%d] = %+v", i, res[i])
			} else if res[i].V != want[i].V && !(math.IsNaN(res[i].V) && math.IsNaN(want[i].V)) {
				t.Errorf("- [%d] = %+v", i, want[i])
				t.Errorf("+ [%d] = %+v", i, res[i])
			}
		}
	}
}

func makeEvalTest(t *testing.T, from, until, eval string, maxNullPoints int, expectedQuery, result string, wantEval []types.EvalResult) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parsedQuery, _ := url.ParseQuery(expectedQuery)
		if !reflect.DeepEqual(r.URL.Query(), parsedQuery) {
			t.Errorf("Expected query is %+v but %+v got", parsedQuery, r.URL.Query())
		}
		if r.URL.Path != "/render/" {
			t.Errorf("Path should be `/render/` but %s found", r.URL.Path)
		}
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintln(w, result)
	}))
	defer ts.Close()

	var res []types.EvalResult
	base := "http://" + ts.Listener.Addr().String()
	e, err := NewRenderEval(base, from, until, eval, 0, maxNullPoints)
	if err != nil {
		t.Error(err)
	} else {
		res, err = e.Eval(context.Background())
		if err == nil {
			compareEvalResult(t, res, wantEval)
		} else {
			t.Error(err)
		}
	}
}

var evalTestCases = []struct {
	From          string
	Until         string
	Eval          string
	MaxNullPoints int
	ExpectedQuery string
	Result        string
	WantEval      []types.EvalResult
}{
	{
		Eval:          "main* > 1.1",
		MaxNullPoints: 1,
		ExpectedQuery: "format=json&target=main*",
		Result: "[{\"target\": \"main1\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}," +
			"{\"target\": \"main2\", \"datapoints\": [[1.0, 1468339853], [0.3, 1468339854], [null, 1468339855]]}]",
		WantEval: []types.EvalResult{
			{
				Name:     "main1",
				T:        1468339855,
				V:        math.NaN(),
				Success:  false,
				IsAbsent: true,
			},
			{
				Name:     "main2",
				T:        1468339855,
				V:        math.NaN(),
				Success:  false,
				IsAbsent: true,
			},
		},
	},
	{
		Eval:          "main* > 0.3",
		MaxNullPoints: 2,
		ExpectedQuery: "format=json&target=main*",
		Result: "[{\"target\": \"main1\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}," +
			"{\"target\": \"main2\", \"datapoints\": [[1.0, 1468339853], [0.3, 1468339854], [null, 1468339855]]}]",
		WantEval: []types.EvalResult{
			{
				Name:     "main1",
				T:        1468339854,
				V:        2.0,
				Success:  true,
				IsAbsent: false,
			},
			{
				Name:     "main2",
				T:        1468339854,
				V:        0.3,
				Success:  false,
				IsAbsent: false,
			},
		},
	},
	{
		Eval:          "main* >= 0.3",
		MaxNullPoints: 2,
		ExpectedQuery: "format=json&target=main*",
		Result: "[{\"target\": \"main1\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}," +
			"{\"target\": \"main2\", \"datapoints\": [[1.0, 1468339853], [0.3, 1468339854], [null, 1468339855]]}]",
		WantEval: []types.EvalResult{
			{
				Name:     "main1",
				T:        1468339854,
				V:        2.0,
				Success:  true,
				IsAbsent: false,
			},
			{
				Name:     "main2",
				T:        1468339854,
				V:        0.3,
				Success:  true,
				IsAbsent: false,
			},
		},
	},
	{
		Eval:          "main* == 0.3",
		MaxNullPoints: 2,
		ExpectedQuery: "format=json&target=main*",
		Result: "[{\"target\": \"main1\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}," +
			"{\"target\": \"main2\", \"datapoints\": [[1.0, 1468339853], [0.3, 1468339854], [null, 1468339855]]}]",
		WantEval: []types.EvalResult{
			{
				Name:     "main1",
				T:        1468339854,
				V:        2.0,
				Success:  false,
				IsAbsent: false,
			},
			{
				Name:     "main2",
				T:        1468339854,
				V:        0.3,
				Success:  true,
				IsAbsent: false,
			},
		},
	},
	{
		Eval:          "main* < 0.3",
		MaxNullPoints: 2,
		ExpectedQuery: "format=json&target=main*",
		Result: "[{\"target\": \"main1\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}," +
			"{\"target\": \"main2\", \"datapoints\": [[1.0, 1468339853], [0.3, 1468339854], [null, 1468339855]]}]",
		WantEval: []types.EvalResult{
			{
				Name:     "main1",
				T:        1468339854,
				V:        2.0,
				Success:  false,
				IsAbsent: false,
			},
			{
				Name:     "main2",
				T:        1468339854,
				V:        0.3,
				Success:  false,
				IsAbsent: false,
			},
		},
	},
	{
		Eval:          "main* <= 0.3",
		MaxNullPoints: 2,
		ExpectedQuery: "format=json&target=main*",
		Result: "[{\"target\": \"main1\", \"datapoints\": [[1.1, 1468339853], [2, 1468339854], [null, 1468339855]]}," +
			"{\"target\": \"main2\", \"datapoints\": [[1.0, 1468339853], [0.3, 1468339854], [null, 1468339855]]}]",
		WantEval: []types.EvalResult{
			{
				Name:     "main1",
				T:        1468339854,
				V:        2.0,
				Success:  false,
				IsAbsent: false,
			},
			{
				Name:     "main2",
				T:        1468339854,
				V:        0.3,
				Success:  true,
				IsAbsent: false,
			},
		},
	},
}

func TestGraphiteClient_Eval(t *testing.T) {
	for i, tt := range evalTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			makeEvalTest(t, tt.From, tt.Until, tt.Eval, tt.MaxNullPoints, tt.ExpectedQuery, tt.Result, tt.WantEval)
		})
	}
}
