# Project Issues Report - Botla Backend

Generated on: 2026-01-16

This report summarizes issues found during a comprehensive scan of the Botla backend project.

## Summary

- **Total Linting Issues**: 60
- **Critical Security Issues**: 2 potential SQL injection vulnerabilities
- **Code Quality Issues**: Multiple unchecked errors, variable shadowing
- **Configuration Issues**: Exposed database ports in Docker Compose
- **TODO Items**: 1 pending implementation

## Detailed Findings

### 1. Linting Issues (60 total)

#### Error Checking (24 issues - errcheck)
- **Issue**: Multiple `rows.Close()` calls are not checking return values
- **Files**: `internal/repository/*.go` (chatbot_repo.go, organization_repo.go, etc.)
- **Impact**: Potential resource leaks if Close() fails
- **Severity**: Medium

#### Security Issues (18 issues - gosec)
- **SQL String Formatting (G201)**: 2 instances
  - `internal/integration/fixtures/env.go:552`: Dynamic column name in UPDATE query
  - `internal/repository/analytics_repo.go:370`: WHERE clause insertion
- **Integer Overflow (G115)**: 16 instances
  - Converting `int` to `uint64` without bounds checking
  - Files: `internal/repository/*.go`
- **Severity**: High (SQL injection), Medium (overflow)

#### Code Quality Issues (12 issues - govet)
- **Variable Shadowing**: 12 instances of variable shadowing (e.g., `err` variable)
- **Files**: `internal/repository/privacy_repo.go`, `internal/repository/user_repo.go`
- **Impact**: Potential bugs due to variable confusion
- **Severity**: Medium

#### Static Analysis Issues (3 issues - staticcheck)
- **Unused Field**: `internal/processing/sources_queue.go:28` - `loader` field unused
- **Always True Comparison**: `internal/processing/pdf_processor.go:94` - `err != nil` flagged as always true (likely false positive for stub implementation)
- **Severity**: Low

#### Other Issues
- **Deprecated Comment**: `internal/processing/suggestions.go:219` - Incorrect deprecation comment format
- **Unwrapped Errors**: 2 instances (`wrapcheck`) - External package errors not wrapped

### 2. TODO/FIXME Items

- **1 TODO found**:
  - `internal/services/handoff_service.go:198`: "TODO: Implement actual email sending when SMTP service is available"

### 3. Security Vulnerabilities

#### Potential SQL Injection
- **Location**: `internal/integration/fixtures/env.go:552`
- **Code**:
  ```go
  query := fmt.Sprintf(`
      UPDATE plan_limits
      SET %s = $1, updated_at = NOW()
      WHERE plan_id = (SELECT id FROM plans WHERE code = $2)
  `, field)
  ```
- **Risk**: If `field` parameter is user-controlled, could lead to SQL injection
- **Mitigation**: Use whitelisted column names or prepared statements

#### Integer Overflow Risks
- **Issue**: Converting `int` to `uint64` for pagination parameters
- **Example**: `Limit(uint64(limit))` where `limit` is `int`
- **Risk**: Negative `int` values become large `uint64` values
- **Mitigation**: Validate input ranges before conversion

### 4. Configuration Issues

#### Docker Compose Security
- **Issue**: Database and Qdrant ports exposed to host
  - PostgreSQL: `5432:5432`
  - Qdrant: `6333:6333`
- **Risk**: In production deployments, these should not be exposed externally
- **Mitigation**: Remove port mappings or restrict to internal networks

### 5. Code Quality Observations

#### Error Handling Patterns
- **Positive**: Extensive use of error wrapping with context (`fmt.Errorf("context: %w", err)`)
- **Issue**: Inconsistent checking of `rows.Close()` and `RowsAffected()`

#### Debug Code
- **Issue**: Debug print statements in production code
  - `internal/api/handlers/admin_plan.go:131`
  - `internal/repository/plan_repo.go:389-413`

### 6. Test Coverage

- **Issue**: CI fails due to linting issues, preventing test execution
- **Coverage Goal**: 90% (mentioned in AGENTS.md)
- **Current Status**: Unknown due to CI failure

## Recommendations

### Immediate Actions (High Priority)
1. **Fix SQL Injection Vulnerabilities**
   - Implement column name whitelisting in `updatePlanLimitField`
   - Review and fix dynamic SQL constructions

2. **Address Integer Overflow Issues**
   - Add bounds checking before `int` to `uint64` conversions
   - Validate pagination parameters

3. **Fix Resource Leaks**
   - Add error checking for all `rows.Close()` calls
   - Ensure proper cleanup in error paths

### Medium Priority
4. **Resolve Variable Shadowing**
   - Rename shadowed variables or use different scopes

5. **Clean Up Debug Code**
   - Remove or conditionalize debug print statements

6. **Fix Docker Security**
   - Remove unnecessary port exposures in production

### Low Priority
7. **Code Cleanup**
   - Remove unused fields
   - Fix deprecated comment formats
   - Address staticcheck warnings

## Next Steps

1. Run `make ci` after fixes to ensure tests pass
2. Implement missing email functionality for handoff service
3. Consider security audit for authentication and authorization logic
4. Review error handling patterns across the codebase for consistency

## Tools Used for Analysis

- golangci-lint (vet, staticcheck, gosec, errcheck, etc.)
- Manual code review
- Regex searches for TODO/FIXME patterns