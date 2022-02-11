package graphiteapi

import "math"

// GetLastNonNullValue searches for the latest non null value, and skips at most maxNullPoints.
// If the last maxNullPoints values are all absent, returns absent
func GetLastNonNullValue(pp *Series, maxNullPoints int) (t int64, v float64, absent bool) {
	l := len(pp.DataPoints)

	if l == 0 {
		// there is values, we should return absent
		v = math.NaN()
		t = 0
		absent = true
		return t, v, absent
	}

	for i := 0; i < maxNullPoints && i < l; i++ {
		if math.IsNaN(pp.DataPoints[l-1-i].Value) {
			continue
		}
		v = pp.DataPoints[l-1-i].Value
		t = pp.DataPoints[l-1-i].Timestamp
		absent = false
		return t, v, absent
	}

	// if we get here, there are two cases
	//   * maxNullPoints == 0, we didn't even enter the loop above
	//   * maxNullPoints > 0, but we didn't find a non-null point in the loop
	// in both cases, we return the last point's info
	v = pp.DataPoints[l-1].Value
	t = pp.DataPoints[l-1].Timestamp
	absent = math.IsNaN(v)
	return t, v, absent
}
