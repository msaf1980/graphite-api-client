package types

type EvalResult struct {
	Name     string
	T        int32
	V        float64
	Success  bool
	IsAbsent bool
}
