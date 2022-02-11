package types

type EvalResult struct {
	Name     string
	T        int64
	V        float64
	Success  bool
	IsAbsent bool
}
