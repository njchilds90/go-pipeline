package pipeline

import "context"

// MapFunc is a pure transformation function used by [Pipeline.Map].
// It accepts a context and a single input value and returns a transformed
// output value or an error.
type MapFunc[Value any] func(ctx context.Context, value Value) (Value, error)

// FilterFunc is a predicate function used by [Pipeline.Filter].
// It accepts a context and a single input value and returns true if the
// item should be kept, or false if it should be removed from the pipeline.
type FilterFunc[Value any] func(ctx context.Context, value Value) (bool, error)

// ReduceFunc is a folding function used by [Pipeline.Reduce].
// It accepts the current accumulator, the current item, and returns
// the updated accumulator or an error.
type ReduceFunc[Value any] func(ctx context.Context, accumulator Value, current Value) (Value, error)

// stageKind identifies the type of operation a stage performs.
type stageKind int

const (
	stageKindMap    stageKind = iota
	stageKindFilter stageKind = iota
	stageKindReduce stageKind = iota
)

// stage is an internal representation of a single pipeline step.
type stage[Value any] struct {
	name   string
	kind   stageKind
	mapFn  MapFunc[Value]
	filtFn FilterFunc[Value]
	redFn  ReduceFunc[Value]
	seed   Value
}

// applyMap runs a map stage against a slice of items.
func (s *stage[Value]) applyMap(ctx context.Context, items []Value) ([]Value, error) {
	out := make([]Value, 0, len(items))
	for i, item := range items {
		transformed, err := s.mapFn(ctx, item)
		if err != nil {
			return nil, &StageError{StageName: s.name, ItemIndex: i, Cause: err}
		}
		out = append(out, transformed)
	}
	return out, nil
}

// applyFilter runs a filter stage against a slice of items.
func (s *stage[Value]) applyFilter(ctx context.Context, items []Value) ([]Value, error) {
	out := make([]Value, 0, len(items))
	for i, item := range items {
		keep, err := s.filtFn(ctx, item)
		if err != nil {
			return nil, &StageError{StageName: s.name, ItemIndex: i, Cause: err}
		}
		if keep {
			out = append(out, item)
		}
	}
	return out, nil
}

// applyReduce runs a reduce stage against a slice of items, collapsing
// them to a single accumulated value wrapped in a one-element slice.
func (s *stage[Value]) applyReduce(ctx context.Context, items []Value) ([]Value, error) {
	acc := s.seed
	for i, item := range items {
		next, err := s.redFn(ctx, acc, item)
		if err != nil {
			return nil, &StageError{StageName: s.name, ItemIndex: i, Cause: err}
		}
		acc = next
	}
	return []Value{acc}, nil
}
