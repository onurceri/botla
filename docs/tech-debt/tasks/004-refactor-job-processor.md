# TASK-004 — Refactor JobProcessor to Interface-Based Design

## Goal
Decouple `JobProcessor` from specific processor implementations by introducing a `SourceProcessor` interface and using a map-based registry strategy, strictly adhering to the Open/Closed Principle.

## Scope
- `internal/processing/job_processor.go`
- `internal/processing` implementations (`url`, `pdf`, `text`).

## Checklist
- [x] Define `SourceProcessor` interface in `internal/processing/interfaces.go`.
    - `ProcessWithSteps(...)`
- [x] Make `URLProcessor`, `PDFProcessor`, `TextProcessor` implement this interface.
- [x] Refactor `JobProcessor` struct:
    - Replace individual fields (`urlProcessor`) with `processors map[string]SourceProcessor`.
- [x] Update `NewJobProcessor` to accept the map or register processors.
- [x] Update `processWithResume` to lookup processor by `source.SourceType`.
- [x] Remove the hardcoded `switch` statement.
- [x] Fix any broken tests in `internal/processing`.

## Edge Cases
- Unknown source type handling (should return specific error).
- Nil processor in map.

## Files Likely to Change
- `internal/processing/job_processor.go`
- `internal/processing/interfaces.go` (NEW)
