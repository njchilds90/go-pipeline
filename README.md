# go-pipeline

[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/go-pipeline.svg)](https://pkg.go.dev/github.com/njchilds90/go-pipeline)
[![CI](https://github.com/njchilds90/go-pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-pipeline/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/njchilds90/go-pipeline)](https://goreportcard.com/report/github.com/njchilds90/go-pipeline)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Chain composable transformation stages, filters, and reducers into clean, readable data pipelines with zero external dependencies.

---

## Why go-pipeline?

Go developers constantly re-implement the same scattered `for` loop patterns: transform this slice, filter that one, fold another. There is no standard, idiomatic way to express a named, multi-step data transformation as a first-class value.

`go-pipeline` fills that gap:

- **Type-safe** via Go generics — no `interface{}` casting
- **Named stages** for readable, auditable transformation chains
- **Structured errors** that identify exactly which stage and item failed
- **Immutable pipelines** — each `Map`, `Filter`, and `Reduce` returns a new value
- **Context-aware** — every stage receives `context.Context` for cancellation support
- **Zero external dependencies** — just Go

---

## Installation
```bash
go get github.com/njchilds90/go-pipeline@v1.0.0
```

---

## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/njchilds90/go-pipeline"
)

func main() {
    p := pipeline.New[string]().
        Map("trim-spaces", func(_ context.Context, s string) (string, error) {
            return strings.TrimSpace(s), nil
        }).
        Filter("non-empty", func(_ context.Context, s string) (bool, error) {
            return s != "", nil
        }).
        Map("to-upper", func(_ context.Context, s string) (string, error) {
            return strings.ToUpper(s), nil
        })

    result, err := p.Run(context.Background(), []string{"  hello  ", "", " world "})
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Items)
    // Output: [HELLO WORLD]
}
```

---

## Core Concepts

### Pipeline

A `Pipeline[Value]` is an ordered sequence of named stages. Pipelines are immutable: each call to `Map`, `Filter`, or `Reduce` returns a new `Pipeline` without modifying the original.
```go
base := pipeline.New[int]().Map("double", doubleFn)
withFilter := base.Filter("positive-only", positiveFn) // base is unchanged
```

### Map

Apply a transformation to every item in the current slice.
```go
p := pipeline.New[int]().Map("square", func(_ context.Context, v int) (int, error) {
    return v * v, nil
})
```

### Filter

Remove items that do not satisfy a predicate.
```go
p := pipeline.New[int]().Filter("even-only", func(_ context.Context, v int) (bool, error) {
    return v%2 == 0, nil
})
```

### Reduce

Fold all items into a single accumulated value. The output of a Reduce stage is a one-element slice containing the final accumulator.
```go
p := pipeline.New[int]().Reduce("sum", 0, func(_ context.Context, acc, v int) (int, error) {
    return acc + v, nil
})
```

---

## Chaining Stages
```go
double := func(_ context.Context, v int) (int, error) { return v * 2, nil }
positive := func(_ context.Context, v int) (bool, error) { return v > 0, nil }
sum := func(_ context.Context, acc, v int) (int, error) { return acc + v, nil }

p := pipeline.New[int]().
    Map("double", double).
    Filter("positive-only", positive).
    Reduce("sum", 0, sum)

result, err := p.Run(context.Background(), []int{-1, 1, 2, 3})
// After double:        [-2, 2, 4, 6]
// After positive-only: [2, 4, 6]
// After sum:           [12]
fmt.Println(result.Items[0]) // 12
```

---

## Structured Errors

When a stage fails, `Run` returns a `*StageError` that identifies exactly where the failure occurred.
```go
result, err := p.Run(ctx, input)
if err != nil {
    var stageErr *pipeline.StageError
    if errors.As(err, &stageErr) {
        fmt.Printf("stage %q failed at item index %d: %v\n",
            stageErr.StageName, stageErr.ItemIndex, stageErr.Cause)
    }
}
```

---

## Stage Reports

Every `Result` includes a per-stage report with input and output counts — useful for logging and debugging.
```go
for _, report := range result.Stages {
    fmt.Printf("stage %q: %d items in, %d items out\n",
        report.StageName, report.InputCount, report.OutputCount)
}
```

---

## Context and Cancellation

Every stage function receives the `context.Context` passed to `Run`. Check it to support cancellation:
```go
p := pipeline.New[int]().Map("check-ctx", func(ctx context.Context, v int) (int, error) {
    if err := ctx.Err(); err != nil {
        return 0, err
    }
    return v * 2, nil
})
```

---

## Works with Any Type

Because `go-pipeline` uses Go generics, it works with any type — strings, structs, maps, or your own domain types.
```go
type Record struct {
    Name  string
    Score int
}

p := pipeline.New[Record]().
    Filter("high-scorers", func(_ context.Context, r Record) (bool, error) {
        return r.Score >= 80, nil
    }).
    Map("normalize-name", func(_ context.Context, r Record) (Record, error) {
        r.Name = strings.TrimSpace(r.Name)
        return r, nil
    })
```

---

## Introspecting a Pipeline
```go
p := pipeline.New[int]().Map("double", doubleFn).Filter("positive-only", positiveFn)

fmt.Println(p.Len())        // 2
fmt.Println(p.StageNames()) // [double positive-only]
```

---

## License

MIT — see [LICENSE](LICENSE)
