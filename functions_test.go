package graphiteapi

import (
	"math"
	"testing"
)

func TestGetLastNonNullValue(t *testing.T) {
	tests := []struct {
		name          string
		pp            Series
		maxNullPoints int
		wantT         int32
		wantV         float64
		wantAbsent    bool
	}{
		{
			name: "last (empty)",
			pp: Series{
				StartTime: 1643964240,
				StopTime:  1643964240,
				StepTime:  60,
			},
			maxNullPoints: 1,
			wantT:         1643964240,
			wantV:         math.NaN(),
			wantAbsent:    true,
		},
		{
			name: "last (NaN)",
			pp: Series{
				StartTime: 1643964240,
				StopTime:  1643964240,
				StepTime:  60,
				Values:    []float64{math.NaN()},
			},
			maxNullPoints: 1,
			wantT:         1643964240,
			wantV:         math.NaN(),
			wantAbsent:    true,
		},
		{
			name: "last (NaN)",
			pp: Series{
				StartTime: 1643964240,
				StopTime:  1643964240,
				StepTime:  60,
				Values:    []float64{math.NaN()},
			},
			maxNullPoints: 1,
			wantT:         1643964240,
			wantV:         math.NaN(),
			wantAbsent:    true,
		},
		{
			name: "last [10.0]",
			pp: Series{
				StartTime: 1643964240,
				StopTime:  1643964240,
				StepTime:  60,
				Values:    []float64{10.0},
			},
			maxNullPoints: 1,
			wantT:         1643964240,
			wantV:         10.0,
			wantAbsent:    false,
		},
		{
			name: "last [10.0, 5.0]",
			pp: Series{
				StartTime: 1643964180,
				StopTime:  1643964240,
				StepTime:  60,
				Values:    []float64{10.0, 5.0},
			},
			maxNullPoints: 1,
			wantT:         1643964240,
			wantV:         5.0,
			wantAbsent:    false,
		},
		{
			name: "last maxNullPoints=1 [10.0, NaN]",
			pp: Series{
				StartTime: 1643964180,
				StopTime:  1643964240,
				StepTime:  60,
				Values:    []float64{10.0, math.NaN()},
			},
			maxNullPoints: 1,
			wantT:         1643964240,
			wantV:         math.NaN(),
			wantAbsent:    true,
		},
		{
			name: "last maxNullPoints=2 [10.0, NaN]",
			pp: Series{
				StartTime: 1643964180,
				StopTime:  1643964240,
				StepTime:  60,
				Values:    []float64{10.0, math.NaN()},
			},
			maxNullPoints: 2,
			wantT:         1643964180,
			wantV:         10.0,
			wantAbsent:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, gotV, gotAbsent := GetLastNonNullValue(&tt.pp, tt.maxNullPoints)
			if gotT != tt.wantT {
				t.Errorf("GetLastNonNullValue() gotT = %v, want %v", gotT, tt.wantT)
			}
			if gotV != tt.wantV && !(math.IsNaN(gotV) && math.IsNaN(tt.wantV)) {
				t.Errorf("GetLastNonNullValue() gotV = %v, want %v", gotV, tt.wantV)
			}
			if gotAbsent != tt.wantAbsent {
				t.Errorf("GetLastNonNullValue() gotAbsent = %v, want %v", gotAbsent, tt.wantAbsent)
			}
		})
	}
}
