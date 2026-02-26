package pipeline

import "context"

// Pipeline is an ordered sequence of named transformation stages applied
// to a slice of items. Use [New] to create a pipeline, then chain [Map],
// [Filter], and [Reduce] calls to add stages. Call [Run] to execute the
// pipeline against an input slice.
//
// Pipeline is immutable after construction: each call to Map, Filter, or
// Reduce returns a new Pipeline value. The original is never modified.
//
// Example (doubling and filtering a slice of integers):
//
//	p := pipeline.New[int]().
//	    Map("double", func(_ context.Context, v int) (int, error) {
//	        return v * 2, nil
//	    }).
//	    Filter("keep-large", func(_ context.Context, v int) (bool, error) {
//	        return v > 4, nil
//	    })
//
//	result, err := p.Run(context.Background(), []int{1, 2, 3, 4, 5})
//	// result.Items == []int{6, 8, 10}
type Pipeline[Value any] struct {
	stages []stage[Value]
}

// New creates and returns an empty [Pipeline] with no stages.
//
// Example:
//
//	p := pipeline.New[string]()
func New[Value any]() Pipeline[Value] {
	return Pipeline[Value]{}
}

// Map adds a named transformation stage to the pipeline. The provided
// function is applied to every item in the current slice. If the function
// returns an error for any item, the pipeline halts and returns a
// [StageError] wrapping the original error.
//
// Map returns a new Pipeline; the receiver is not modified.
//
// Example:
//
//	p := pipeline.New[string]().Map("to-upper", func(_ context.Context, s string) (string, error) {
//	    return strings.ToUpper(s), nil
//	})
func (p Pipeline[Value]) Map(name string, fn MapFunc[Value]) Pipeline[Value] {
	return p.appendStage(stage[Value]{name: name, kind: stageKindMap, mapFn: fn})
}

// Filter adds a named filter stage to the pipeline. Items for which the
// predicate returns false are removed from the slice. If the predicate
// returns an error for any item, the pipeline halts and returns a
// [StageError] wrapping the original error.
//
// Filter returns a new Pipeline; the receiver is not modified.
//
// Example:
//
//	p := pipeline.New[int]().Filter("positive-only", func(_ context.Context, v int) (bool, error) {
//	    return v > 0, nil
//	})
func (p Pipeline[Value]) Filter(name string, fn FilterFunc[Value]) Pipeline[Value] {
	return p.appendStage(stage[Value]{name: name, kind: stageKindFilter, filtFn: fn})
}

// Reduce adds a named reduction stage to the pipeline. All items in the
// current slice are folded into a single accumulated value starting from
// seed. The result of the reduce stage is a one-element slice containing
// the final accumulator value.
//
// Reduce returns a new Pipeline; the receiver is not modified.
//
// Example:
//
//	p := pipeline.New[int]().Reduce("sum", 0, func(_ context.Context, acc, v int) (int, error) {
//	    return acc + v, nil
//	})
func (p Pipeline[Value]) Reduce(name string, seed Value, fn ReduceFunc[Value]) Pipeline[Value] {
	return p.appendStage(stage[Value]{name: name, kind: stageKindReduce, redFn: fn, seed: seed})
}

// Len returns the number of stages currently registered in the pipeline.
//
// Example:
//
//	p := pipeline.New[int]().Map("double", fn).Filter("keep-positive", fn2)
//	p.Len() // returns 2
func (p Pipeline[Value]) Len() int {
	return len(p.stages)
}

// StageNames returns the names of all registered stages in execution order.
//
// Example:
//
//	p := pipeline.New[int]().Map("double", fn).Filter("keep-positive", fn2)
//	p.StageNames() // returns []string{"double", "keep-positive"}
func (p Pipeline[Value]) StageNames() []string {
	names := make([]string, len(p.stages))
	for i, s := range p.stages {
		names[i] = s.name
	}
	return names
}

// Run executes the pipeline against the provided input slice and returns
// a [Result] containing the transformed items and per-stage metadata.
//
// If any stage returns an error, Run halts immediately and returns nil
// and the [StageError]. The context is forwarded to every stage function
// and can be used for cancellation or deadline propagation.
//
// Run does not modify the input slice.
//
// Example:
//
//	result, err := p.Run(context.Background(), []int{1, 2, 3})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Items)
func (p Pipeline[Value]) Run(ctx context.Context, input []Value) (*Result[Value], error) {
	current := make([]Value, len(input))
	copy(current, input)

	reports := make([]StageReport, 0, len(p.stages))

	for _, s := range p.stages {
		inputCount := len(current)

		var (
			next []Value
			err  error
		)

		switch s.kind {
		case stageKindMap:
			next, err = s.applyMap(ctx, current)
		case stageKindFilter:
			next, err = s.applyFilter(ctx, current)
		case stageKindReduce:
			next, err = s.applyReduce(ctx, current)
		}

		if err != nil {
			return nil, err
		}

		reports = append(reports, StageReport{
			StageName:   s.name,
			InputCount:  inputCount,
			OutputCount: len(next),
		})

		current = next
	}

	return &Result[Value]{
		Items:  current,
		Stages: reports,
	}, nil
}

// appendStage returns a new Pipeline with the given stage appended.
func (p Pipeline[Value]) appendStage(s stage[Value]) Pipeline[Value] {
	newStages := make([]stage[Value], len(p.stages)+1)
	copy(newStages, p.stages)
	newStages[len(p.stages)] = s
	return Pipeline[Value]{stages: newStages}
}
