# Contributing to go-pipeline

Thank you for your interest in contributing. This is a small, focused library and contributions are welcome.

---

## Reporting Issues

Open a GitHub Issue with:

- A clear description of the problem or request
- A minimal reproducible example if reporting a bug
- The Go version and operating system you are using

---

## Submitting Changes

1. Fork the repository
2. Create a branch named `feature/your-feature` or `fix/your-fix`
3. Make your changes
4. Ensure all tests pass: `go test -race ./...`
5. Ensure `go vet ./...` reports no issues
6. Open a pull request against `main` with a clear description

---

## Code Standards

- All exported symbols must have GoDoc comments
- New behavior must be covered by table-driven tests
- No external runtime dependencies may be added
- Public interfaces must remain backward compatible within a major version
- Use `context.Context` as the first parameter in all stage function signatures

---

## Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/):

- Patch releases (`v1.0.x`) — bug fixes only
- Minor releases (`v1.x.0`) — backward-compatible new features
- Major releases (`vX.0.0`) — breaking changes

---

## License

By contributing you agree that your contributions will be licensed under the MIT License.
```

---

## 4. Release and Verification Instructions
```
CREATING THE v1.0.0 RELEASE VIA GITHUB UI
==========================================

STEP 1 — Create the tag and release:
1. On your repository page, click "Releases" in the right sidebar
   (or go to: https://github.com/njchilds90/go-pipeline/releases)
2. Click "Create a new release"
3. Click "Choose a tag"
4. Type exactly:  v1.0.0
5. Click "Create new tag: v1.0.0 on publish"
6. Set Release title to:  v1.0.0 — Initial Release
7. In the description box, paste:

## go-pipeline v1.0.0

Initial release of go-pipeline — a type-safe, composable data transformation
pipeline for Go using generics.

### What is included

- `Map`, `Filter`, and `Reduce` stages with named registration
- Immutable pipeline chaining
- Structured `StageError` with stage name, item index, and unwrappable cause
- Per-stage execution reports via `Result.Stages`
- Full `context.Context` propagation
- Zero external dependencies
- Table-driven tests with race detector
- GitHub Actions CI across Go 1.21, 1.22, and 1.23

### Install

    go get github.com/njchilds90/go-pipeline@v1.0.0

8. Leave "Set as the latest release" checked
9. Click "Publish release"


STEP 2 — Trigger pkg.go.dev indexing:
Visit this URL in your browser (this nudges the proxy to index immediately):

    https://sum.golang.org/lookup/github.com/njchilds90/go-pipeline@v1.0.0

Then visit:

    https://proxy.golang.org/github.com/njchilds90/go-pipeline/@v/v1.0.0.info


STEP 3 — Verify the package is live (~10 minutes after release):

    https://pkg.go.dev/github.com/njchilds90/go-pipeline


SEMANTIC VERSIONING GUIDANCE FOR FUTURE RELEASES
=================================================

v1.0.x  — Bug fixes only. No new exported symbols. No behavior changes.
v1.x.0  — New exported functions, types, or stage helpers that are
           fully backward compatible with v1.0.0.
v2.0.0  — Any breaking change to existing public signatures or behavior.
           (Requires updating the module path to .../go-pipeline/v2)

Recommended next minor additions (v1.1.0):
- Tap(name, fn)  — side-effect stage that observes items without transforming
- Batch(name, size, fn)  — process items in fixed-size groups
- FlatMap(name, fn) — map one item to zero or more output items
