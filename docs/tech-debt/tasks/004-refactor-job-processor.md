# TASK-004 — Refactor JobProcessor to Interface-Based Design

## Goal
Decouple `JobProcessor` from specific processor implementations by introducing a `SourceProcessor` interface and using a map-based registry strategy, strictly adhering to the Open/Closed Principle.

## Scope
- `internal/processing/job_processor.go`
- `internal/processing` implementations (`url`, `pdf`, `text`).

## Checklist
- [ ] Define `SourceProcessor` interface in `internal/processing/interfaces.go`.
    - `ProcessWithSteps(ctx, ...)`
- [ ] Make `URLProcessor`, `PDFProcessor`, `TextProcessor` implement this interface.
- [ ] Refactor `JobProcessor` struct:
    - Replace individual fields (`urlProcessor`) with `processors map[string]SourceProcessor`.
- [ ] Update `NewJobProcessor` to accept the map or register processors.
- [ ] Update `processWithResume` to lookup processor by `source.SourceType`.
- [ ] Remove the hardcoded `switch` statement.
- [ ] Fix any broken tests in `internal/processing`.

## Edge Cases
- Unknown source type handling (should return specific error).
- Nil processor in map.

## Files Likely to Change
- `internal/processing/job_processor.go`
- `internal/processing/interfaces.go` (NEW)
