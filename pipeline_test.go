package pipeline_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/njchilds90/go-pipeline"
)

func TestNew_EmptyPipeline(t *testing.T) {
	p := pipeline.New[int]()
	if p.Len() != 0 {
		t.Fatalf("expected 0 stages, got %d", p.Len())
	}
}

func TestPipeline_Map(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "double all values",
			input:    []int{1, 2, 3},
			expected: []int{2, 4, 6},
		},
		{
			name:     "empty input",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "single value",
			input:    []int{5},
			expected: []int{10},
		},
	}

	double := func(_ context.Context, v int) (int, error) { return v * 2, nil }

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := pipeline.New[int]().Map("double", double)
			result, err := p.Run(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Items) != len(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, result.Items)
			}
			for i, v := range result.Items {
				if v != tc.expected[i] {
					t.Errorf("index %d: expected %d, got %d", i, tc.expected[i], v)
				}
			}
		})
	}
}

func TestPipeline_Filter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "keep positive values",
			input:    []int{-2, -1, 0, 1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "all filtered out",
			input:    []int{-3, -2, -1},
			expected: []int{},
		},
		{
			name:     "none filtered out",
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
	}

	positive := func(_ context.Context, v int) (bool, error) { return v > 0, nil }

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := pipeline.New[int]().Filter("positive-only", positive)
			result, err := p.Run(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Items) != len(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, result.Items)
			}
			for i, v := range result.Items {
				if v != tc.expected[i] {
					t.Errorf("index %d: expected %d, got %d", i, tc.expected[i], v)
				}
			}
		})
	}
}

func TestPipeline_Reduce(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		seed     int
		expected int
	}{
		{
			name:     "sum to zero seed",
			input:    []int{1, 2, 3, 4},
			seed:     0,
			expected: 10,
		},
		{
			name:     "sum with non-zero seed",
			input:    []int{1, 2, 3},
			seed:     10,
			expected: 16,
		},
		{
			name:     "empty input returns seed",
			input:    []int{},
			seed:     42,
			expected: 42,
		},
	}

	sum := func(_ context.Context, acc, v int) (int, error) { return acc + v, nil }

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := pipeline.New[int]().Reduce("sum", tc.seed, sum)
			result, err := p.Run(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Items) != 1 {
				t.Fatalf("expected 1 item from reduce, got %d", len(result.Items))
			}
			if result.Items[0] != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result.Items[0])
			}
		})
	}
}

func TestPipeline_Chained(t *testing.T) {
	double := func(_ context.Context, v int) (int, error) { return v * 2, nil }
	positive := func(_ context.Context, v int) (bool, error) { return v > 0, nil }
	sum := func(_ context.Context, acc, v int) (int, error) { return acc + v, nil }

	p := pipeline.New[int]().
		Map("double", double).
		Filter("positive-only", positive).
		Reduce("sum", 0, sum)

	result, err := p.Run(context.Background(), []int{-1, 1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After double: [-2, 2, 4, 6]
	// After filter: [2, 4, 6]
	// After sum: [12]
	if len(result.Items) != 1 || result.Items[0] != 12 {
		t.Errorf("expected [12], got %v", result.Items)
	}
	if len(result.Stages) != 3 {
		t.Errorf("expected 3 stage reports, got %d", len(result.Stages))
	}
}

func TestPipeline_StageReportCounts(t *testing.T) {
	double := func(_ context.Context, v int) (int, error) { return v * 2, nil }
	positive := func(_ context.Context, v int) (bool, error) { return v > 0, nil }

	p := pipeline.New[int]().
		Map("double", double).
		Filter("positive-only", positive)

	result, err := p.Run(context.Background(), []int{-1, 1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Stages[0].InputCount != 3 || result.Stages[0].OutputCount != 3 {
		t.Errorf("map stage: expected in=3 out=3, got in=%d out=%d",
			result.Stages[0].InputCount, result.Stages[0].OutputCount)
	}
	// After double: [-2, 2, 4] — filter removes -2, keeps 2 and 4
	if result.Stages[1].InputCount != 3 || result.Stages[1].OutputCount != 2 {
		t.Errorf("filter stage: expected in=3 out=2, got in=%d out=%d",
			result.Stages[1].InputCount, result.Stages[1].OutputCount)
	}
}

func TestPipeline_MapError(t *testing.T) {
	boom := errors.New("boom")
	failFn := func(_ context.Context, v int) (int, error) {
		if v == 2 {
			return 0, boom
		}
		return v, nil
	}

	p := pipeline.New[int]().Map("fail-on-two", failFn)
	_, err := p.Run(context.Background(), []int{1, 2, 3})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var stageErr *pipeline.StageError
	if !errors.As(err, &stageErr) {
		t.Fatalf("expected *pipeline.StageError, got %T", err)
	}
	if stageErr.StageName != "fail-on-two" {
		t.Errorf("expected stage name 'fail-on-two', got %q", stageErr.StageName)
	}
	if stageErr.ItemIndex != 1 {
		t.Errorf("expected item index 1, got %d", stageErr.ItemIndex)
	}
	if !errors.Is(err, boom) {
		t.Errorf("expected errors.Is to find boom via Unwrap")
	}
}

func TestPipeline_FilterError(t *testing.T) {
	boom := errors.New("filter-boom")
	failFilter := func(_ context.Context, v int) (bool, error) {
		if v == 3 {
			return false, boom
		}
		return true, nil
	}

	p := pipeline.New[int]().Filter("fail-on-three", failFilter)
	_, err := p.Run(context.Background(), []int{1, 2, 3})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var stageErr *pipeline.StageError
	if !errors.As(err, &stageErr) {
		t.Fatalf("expected *pipeline.StageError, got %T", err)
	}
	if stageErr.ItemIndex != 2 {
		t.Errorf("expected item index 2, got %d", stageErr.ItemIndex)
	}
}

func TestPipeline_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	checkCtx := func(c context.Context, v int) (int, error) {
		if err := c.Err(); err != nil {
			return 0, err
		}
		return v, nil
	}

	p := pipeline.New[int]().Map("check-context", checkCtx)
	_, err := p.Run(ctx, []int{1, 2, 3})
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestPipeline_StringType(t *testing.T) {
	toUpper := func(_ context.Context, s string) (string, error) {
		return strings.ToUpper(s), nil
	}
	nonEmpty := func(_ context.Context, s string) (bool, error) {
		return s != "", nil
	}

	p := pipeline.New[string]().
		Map("to-upper", toUpper).
		Filter("non-empty", nonEmpty)

	result, err := p.Run(context.Background(), []string{"hello", "", "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"HELLO", "WORLD"}
	if len(result.Items) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result.Items)
	}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestPipeline_DoesNotMutateInput(t *testing.T) {
	input := []int{1, 2, 3}
	original := []int{1, 2, 3}

	double := func(_ context.Context, v int) (int, error) { return v * 2, nil }
	p := pipeline.New[int]().Map("double", double)
	_, err := p.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, v := range input {
		if v != original[i] {
			t.Errorf("input was mutated at index %d: expected %d, got %d", i, original[i], v)
		}
	}
}

func TestPipeline_Len_And_StageNames(t *testing.T) {
	double := func(_ context.Context, v int) (int, error) { return v * 2, nil }
	positive := func(_ context.Context, v int) (bool, error) { return v > 0, nil }

	p := pipeline.New[int]().
		Map("double", double).
		Filter("positive-only", positive)

	if p.Len() != 2 {
		t.Errorf("expected 2 stages, got %d", p.Len())
	}

	names := p.StageNames()
	if names[0] != "double" || names[1] != "positive-only" {
		t.Errorf("unexpected stage names: %v", names)
	}
}

func TestPipeline_Immutability(t *testing.T) {
	double := func(_ context.Context, v int) (int, error) { return v * 2, nil }
	positive := func(_ context.Context, v int) (bool, error) { return v > 0, nil }

	base := pipeline.New[int]().Map("double", double)
	extended := base.Filter("positive-only", positive)

	if base.Len() != 1 {
		t.Errorf("base pipeline was mutated: expected 1 stage, got %d", base.Len())
	}
	if extended.Len() != 2 {
		t.Errorf("extended pipeline: expected 2 stages, got %d", extended.Len())
	}
}
