package graphiteapi

import (
	"math"
	"strconv"
	"testing"
)

var testUnmarshallMetricsCases = []struct {
	Json   string
	Result []Series
	Err    error
}{
	{
		Json:   "",
		Result: []Series{},
		Err:    nil},
	{
		Json: "[{\"target\": \"main\", \"datapoints\": [[1, 1468339853], [1.1, 1468339854], [null, 1468339855]]}]",
		Result: []Series{
			{
				Target: "main",
				DataPoints: []DataPoint{
					{Value: 1.0, Timestamp: 1468339853},
					{Value: 1.1, Timestamp: 1468339854},
					{Value: math.NaN(), Timestamp: 1468339855},
				},
			},
		},
		Err: nil,
	},
}

func TestUnmarshallMetrics(t *testing.T) {
	for i, tt := range testUnmarshallMetricsCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res, err := unmarshallSeries([]byte(tt.Json), 0, 0)
			compareSeries(t, res, tt.Result)

			if err != tt.Err {
				t.Errorf("E %+v", err)
			}
		})
	}
}
