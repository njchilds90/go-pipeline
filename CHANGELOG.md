# Changelog

All notable changes to go-pipeline will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-02-26

### Added

- `Pipeline[Value]` generic type representing an immutable, ordered sequence of named transformation stages
- `New[Value]()` constructor returning an empty pipeline
- `Map(name, fn)` — adds a named transformation stage applied to every item
- `Filter(name, fn)` — adds a named filter stage that removes items not matching a predicate
- `Reduce(name, seed, fn)` — adds a named reduction stage that folds items into a single accumulated value
- `Run(ctx, input)` — executes all stages in order and returns a `*Result`
- `Len()` — returns the number of registered stages
- `StageNames()` — returns stage names in execution order
- `Result[Value]` struct with `Items` (final output slice) and `Stages` (per-stage reports)
- `StageReport` struct with `StageName`, `InputCount`, and `OutputCount`
- `StageError` struct implementing `error` and `Unwrap()` for structured failure identification
- Full `context.Context` propagation to every stage function
- Immutable pipeline design — `Map`, `Filter`, and `Reduce` never modify the receiver
- Input slice protection — `Run` copies the input before processing
- Table-driven test suite with race detector coverage
- GitHub Actions workflow testing Go 1.21, 1.22, and 1.23
- GoDoc examples on all exported functions
