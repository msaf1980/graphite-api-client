package graphiteapi

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/msaf1980/graphite-api-client/types"
)

type EvalCmp int8

const (
	EvalEq EvalCmp = iota
	EvalLt
	EvalLe
	EvalGt
	EvalGe
)

var (
	ErrCmpTargetEmpy   = errors.New("empty target")
	ErrCmpInvalid      = errors.New("invalid comparator")
	ErrCmpValueInvalid = errors.New("invalid value for comparator")
)

type RenderEval struct {
	eval          string
	q             *RenderQuery
	maxNullPoints int
	cmp           EvalCmp
	v             float64
}

func splitEval(eval string) (string, EvalCmp, float64, error) {
	var (
		target string
		cmp    EvalCmp

		start int
		end   int
	)
	if start = strings.Index(eval, "<="); start > 0 {
		cmp = EvalLe
		end = start + 2
	} else if start = strings.Index(eval, "<"); start > 0 {
		cmp = EvalLt
		end = start + 1
	} else if start = strings.Index(eval, ">="); start > 0 {
		cmp = EvalGe
		end = start + 2
	} else if start = strings.Index(eval, ">"); start > 0 {
		cmp = EvalGt
		end = start + 1
	} else if start = strings.Index(eval, "=="); start > 0 {
		cmp = EvalEq
		end = start + 2
	} else {
		return "", EvalEq, 0.0, ErrCmpInvalid
	}

	target = strings.TrimSpace(eval[0:start])
	if len(target) == 0 {
		return "", EvalEq, 0.0, ErrCmpTargetEmpy
	}

	if end < len(eval) {
		vStr := strings.TrimSpace(eval[end:])
		v, err := strconv.ParseFloat(vStr, 64)
		if err != nil {
			return "", EvalEq, 0.0, ErrCmpValueInvalid
		}

		return target, cmp, v, nil
	}

	return "", EvalEq, 0.0, ErrCmpValueInvalid
}

func NewRenderEval(base, from, until, eval string, maxNullPoints int) (*RenderEval, error) {
	if target, cmp, v, err := splitEval(eval); err == nil {
		return &RenderEval{
			eval:          eval,
			q:             NewRenderQuery(base, from, until, []string{target}),
			cmp:           cmp,
			v:             v,
			maxNullPoints: maxNullPoints,
		}, nil
	} else {
		return nil, err
	}
}

func (e *RenderEval) SetBasicAuth(username, password string) {
	e.q.SetBasicAuth(username, password)
}

func (e *RenderEval) Eval(ctx context.Context) ([]types.EvalResult, error) {
	if series, err := e.q.Request(ctx); err == nil {
		results := make([]types.EvalResult, len(series))
		for i := 0; i < len(series); i++ {
			results[i].Name = series[i].Target
			results[i].T, results[i].V, results[i].IsAbsent = GetLastNonNullValue(&series[i], e.maxNullPoints)
			switch e.cmp {
			case EvalLt:
				results[i].Success = results[i].V < e.v
			case EvalLe:
				results[i].Success = results[i].V <= e.v
			case EvalGt:
				results[i].Success = results[i].V > e.v
			case EvalGe:
				results[i].Success = results[i].V >= e.v
			default:
				results[i].Success = results[i].V == e.v
			}
		}

		return results, nil
	} else {
		return nil, err
	}
}
