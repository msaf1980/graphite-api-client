package graphiteapi

import "math"

// GetLastNonNullValue searches for the latest non null value, and skips at most maxNullPoints.
// If the last maxNullPoints values are all absent, returns absent
func GetLastNonNullValue(pp *Series, maxNullPoints int) (t int32, v float64, absent bool) {
	l := len(pp.Values)

	if l == 0 {
		// there is values, we should return absent
		v = math.NaN()
		t = pp.StopTime
		absent = true
		return t, v, absent
	}

	for i := 0; i < maxNullPoints && i < l; i++ {
		if math.IsNaN(pp.Values[l-1-i]) {
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
