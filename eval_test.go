package graphiteapi

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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
			gotTargets, gotEvalCmp, gotV, err := splitEval(tt.eval)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitEval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTargets, tt.wantTarget) {
				t.Errorf("splitEval() got = %v, want %v", gotTargets, tt.wantTarget)
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

func renderEvalTest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse form error", http.StatusBadRequest)
		return
	}
	targets := r.Form.Get("target")
	from := r.Form.Get("from")
	until := r.Form.Get("until")
	format := r.Form.Get("format")

	if format == "protobuf" {
		pb := V2PB{}
		pb.initBuffer()
		writer := bufio.NewWriterSize(w, 1024*1024)
		defer writer.Flush()

		if from == "-5m" || until == "now" {
			if targets == "TEST.*" {
				pb.writeBody(writer, "TEST.1", 1643964180, 1643964240, 60, []float64{10.0, 5.0})
				pb.writeBody(writer, "TEST.2", 1643964180, 1643964240, 60, []float64{1.0, 2.0})
			}
		}
	} else {
		http.Error(w, "invalid format", http.StatusBadRequest)
	}
}

func TestRenderEval_Eval(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/render/", renderEvalTest)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	initV2PB()

	base := "http://" + ts.Listener.Addr().String()

	ctx := context.Background()

	tests := []struct {
		eval          string
		from          string
		until         string
		maxNullPoints int

		wantResult []RenderEvalResult
		wantErr    bool
	}{
		{
			eval:          "TEST.*<2",
			from:          "-1m",
			until:         "now",
			maxNullPoints: 2,
			wantResult: []RenderEvalResult{
				{Name: "TEST.1", T: 1643964240, V: 5.0, Success: false, IsAbsent: false},
				{Name: "TEST.2", T: 1643964240, V: 2.0, Success: false, IsAbsent: false},
			},
			wantErr: false,
		},
		{
			eval:          "TEST.*<=2.0",
			from:          "-1m",
			until:         "now",
			maxNullPoints: 2,
			wantResult: []RenderEvalResult{
				{Name: "TEST.1", T: 1643964240, V: 5.0, Success: false, IsAbsent: false},
				{Name: "TEST.2", T: 1643964240, V: 2.0, Success: true, IsAbsent: false},
			},
			wantErr: false,
		},
		{
			eval:          "TEST.*>2.0",
			from:          "-1m",
			until:         "now",
			maxNullPoints: 2,
			wantResult: []RenderEvalResult{
				{Name: "TEST.1", T: 1643964240, V: 5.0, Success: true, IsAbsent: false},
				{Name: "TEST.2", T: 1643964240, V: 2.0, Success: false, IsAbsent: false},
			},
			wantErr: false,
		},
		{
			eval:          "TEST.*>=2.0",
			from:          "-1m",
			until:         "now",
			maxNullPoints: 2,
			wantResult: []RenderEvalResult{
				{Name: "TEST.1", T: 1643964240, V: 5.0, Success: true, IsAbsent: false},
				{Name: "TEST.2", T: 1643964240, V: 2.0, Success: true, IsAbsent: false},
			},
			wantErr: false,
		},
		{
			eval:          "TEST.*==2.0",
			from:          "-1m",
			until:         "now",
			maxNullPoints: 2,
			wantResult: []RenderEvalResult{
				{Name: "TEST.1", T: 1643964240, V: 5.0, Success: false, IsAbsent: false},
				{Name: "TEST.2", T: 1643964240, V: 2.0, Success: true, IsAbsent: false},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.eval, func(t *testing.T) {
			e, err := NewRenderEval(base, tt.from, tt.until, tt.eval, tt.maxNullPoints)
			if err == nil {
				gotResult, err := e.Eval(ctx)
				if (err != nil) != tt.wantErr {
					t.Errorf("RenderEval.Eval() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(gotResult, tt.wantResult) {
					t.Errorf("RenderEval.Eval() = %v, want %v", gotResult, tt.wantResult)
				}
			} else {
				t.Errorf("RenderEval.Eval() error = %v", err)
			}
		})
	}
}
