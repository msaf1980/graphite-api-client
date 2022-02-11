package graphiteapi

import (
	"math"
	"strconv"

	"github.com/buger/jsonparser"
)

func unmarshallSeries(data []byte, maxTargets, maxDataPoints int) ([]Series, error) {
	empty := []Series{}
	if len(data) == 0 {
		return empty, nil
	}
	var ie error = nil
	result := make([]Series, 0, maxTargets)
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		datapoints, e := unmarshallDatapoints(value, maxDataPoints)
		if e != nil {
			ie = e
			return
		}

		target, e := jsonparser.GetString(value, "target")
		if e != nil {
			ie = e
			return
		}

		result = append(result, Series{Target: target, DataPoints: datapoints})
	})

	if err != nil {
		return empty, err
	}
	if ie != nil {
		return empty, ie
	}
	return result, nil
}

func unmarshallDatapoints(data []byte, maxDataPoints int) ([]DataPoint, error) {
	empty, result := []DataPoint{}, []DataPoint{}
	rawData, _, _, err := jsonparser.Get(data, "datapoints")
	if err != nil {
		return empty, err
	}

	_, err = jsonparser.ArrayEach(rawData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		datapoint, e := unmarshallDatapoint(value)
		if e != nil {
			err = e
			return
		}
		result = append(result, datapoint)
	})
	if err != nil {
		return empty, err
	}
	return result, nil
}

func unmarshallDatapoint(data []byte) (DataPoint, error) {
	empty, result := DataPoint{}, DataPoint{}
	var err error
	position := 0
	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		if position == 0 {
			if dataType == jsonparser.Null {
				result.Value = math.NaN()
			} else {
				v, e := strconv.ParseFloat(string(value), 64)
				if e != nil {
					err = e
					return
				}
				result.Value = v
			}
		} else {
			ts, e := strconv.ParseInt(string(value), 10, 32)
			if err != nil {
				err = e
				return
			}
			result.Timestamp = ts
		}
		position++
	})
	if err != nil {
		return empty, err
	}
	return result, nil
}
