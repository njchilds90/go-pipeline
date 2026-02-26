package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/njchilds90/go-pipeline"
)

func TestStageError_Error(t *testing.T) {
	cause := errors.New("underlying failure")
	e := &pipeline.StageError{
		StageName: "my-stage",
		ItemIndex: 3,
		Cause:     cause,
	}

	msg := e.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if !errors.Is(e, cause) {
		t.Error("expected errors.Is to find cause via Unwrap")
	}
}

func TestReduceStage_SingleItem(t *testing.T) {
	sum := func(_ context.Context, acc, v int) (int, error) { return acc + v, nil }
	p := pipeline.New[int]().Reduce("sum", 0, sum)

	result, err := p.Run(context.Background(), []int{7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Items[0] != 7 {
		t.Errorf("expected 7, got %d", result.Items[0])
	}
}

func TestReduceStage_Error(t *testing.T) {
	boom := errors.New("reduce-boom")
	failReduce := func(_ context.Context, acc, v int) (int, error) {
		if v == 5 {
			return 0, boom
		}
		return acc + v, nil
	}

	p := pipeline.New[int]().Reduce("fail-on-five", 0, failReduce)
	_, err := p.Run(context.Background(), []int{1, 2, 5, 3})
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
