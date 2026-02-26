// Package pipeline provides a type-safe, composable data transformation
// pipeline for Go. Chain transformation stages, filters, and reducers
// over any data type using Go generics, with zero external dependencies.
//
// Core concepts:
//
//   - A [Stage] is a single named transformation function.
//   - A [Pipeline] is an ordered sequence of stages applied to a slice of items.
//   - A [Result] carries the output of a pipeline run along with per-stage metadata.
//
// Example:
//
//	p := pipeline.New[int]().
//	    Map("double", func(_ context.Context, v int) (int, error) {
//	        return v * 2, nil
//	    }).
//	    Filter("keep-positive", func(_ context.Context, v int) (bool, error) {
//	        return v > 0, nil
//	    })
//
//	result, err := p.Run(context.Background(), []int{-1, 2, 3})
package pipeline

import "fmt"

// StageError represents an error that occurred within a named pipeline stage.
// It carries the name of the stage and the index of the item being processed
// so callers can locate the exact failure point.
type StageError struct {
	// StageName is the name given to the stage when it was registered.
	StageName string

	// ItemIndex is the zero-based index of the item in the input slice
	// that caused the error.
	ItemIndex int

	// Cause is the underlying error returned by the stage function.
	Cause error
}

// Error implements the error interface.
func (e *StageError) Error() string {
	return fmt.Sprintf("pipeline stage %q failed at item index %d: %v", e.StageName, e.ItemIndex, e.Cause)
}

// Unwrap returns the underlying cause so errors.Is and errors.As work correctly.
func (e *StageError) Unwrap() error {
	return e.Cause
}

// StageReport contains execution metadata for a single stage in a pipeline run.
type StageReport struct {
	// StageName is the registered name of the stage.
	StageName string

	// InputCount is the number of items entering this stage.
	InputCount int

	// OutputCount is the number of items exiting this stage after
	// transformation or filtering.
	OutputCount int
}

// Result is the output of a [Pipeline.Run] call. It contains the final
// transformed items and a per-stage execution report.
type Result[Output any] struct {
	// Items contains the final output slice after all stages have been applied.
	Items []Output

	// Stages contains one [StageReport] per stage, in execution order.
	Stages []StageReport
}
