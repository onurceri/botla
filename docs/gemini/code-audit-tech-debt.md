As Technical Lead, I have completed a code quality and technical debt audit of the **Botla.co** codebase. The codebase demonstrates a solid functional foundation but shows signs of "rapid-growth debt," particularly where architectural boundaries (Handlers vs. Services) have begun to blur and where "transitional" patterns have been left in place.

Below are the findings prioritized by their long-term impact on maintainability and system stability.

### 1. "Transitional" and Brittle Service Initialization

* **Status**: **RESOLVED**
* **Problem Description**: Core services (like `ChatService`) were being manually initialized within handler helpers using `nil` dependencies.
* **Resolution Verification**: `PublicHandlers` is now correctly initialized in `internal/api/router/router.go` with a fully instantiated `ChatService`. The "transitional" helper with `nil` arguments has been removed from `internal/api/handlers/public.go`.

### 2. Monolithic Frontend Components (The "God Component" Pattern)

* **Status**: **Active / High Priority**
* **Problem Description**: Critical frontend features are implemented as massive, monolithic files. The onboarding logic is particularly affected.
* **Why it increases maintenance cost**: Files exceeding 600 lines with complex state logic are difficult to test, prone to regression, and hard for new developers to parse. It prevents component reuse and makes unit testing individual steps of a wizard nearly impossible.
* 
**Evidence**: `frontend/src/components/onboarding/OnboardingWizard.tsx` is **22.83kb** and contains **632 lines** of code.


* **Refactoring Recommendation**: Decompose the `OnboardingWizard` into a state machine managed by a custom hook (e.g., `useOnboardingState`) and split each "step" of the wizard into individual, stateless functional components.

### 3. Tight Coupling and Poor Testability in Processing Pipelines

* **Status**: **Active**
* **Problem Description**: Core RAG (Retrieval-Augmented Generation) logic depends on package-level functions rather than interfaces, making them "hard to mock."
* **Why it increases maintenance cost**: It forces integration tests to use real infrastructure (or complex stubs) even for simple unit logic. This slows down the CI/CD pipeline and makes it difficult to simulate edge cases (like scraper failures).
* 
**Evidence**: The `URLProcessor` unit tests note that `scraper.ScrapeURLWithFallback` is *"package-level and hard to mock without more refactoring"*.


* **Refactoring Recommendation**: Define a `Scraper` interface. Update `URLProcessor` to accept this interface as a dependency. Move the logic in `scraper.ScrapeURLWithFallback` to a struct implementing this interface.

### 4. Unhandled Edge Cases in Security Middleware

* **Status**: **Active**
* **Problem Description**: Critical RBAC (Role-Based Access Control) middleware fails ungracefully on malformed input, leaking 500 errors instead of returning 400/404.
* **Why it increases maintenance cost**: It masks the true nature of errors and fills logs with noise. In a multi-tenant system, unhandled parsing errors can sometimes be exploited or lead to unexpected bypasses if the code continues execution.
* 
**Evidence**: Integration tests for RBAC note: *"Currently returns 500 (Internal Server Error) due to unhandled UUID parsing error... this should be fixed"*.


* **Refactoring Recommendation**: Implement a robust validation layer for all path parameters (UUIDs) before they reach the database or RBAC logic. Use a middleware that returns a standard `400 Bad Request` for malformed IDs.

### 5. High Duplication in Test Infrastructure

* **Status**: **Active**
* **Problem Description**: Test setup logic (database seeding, user creation, organization mocking) is repeated across dozens of test files.
* 
**Why it increases maintenance cost**: If the database schema for `users` or `plans` changes, developers must update setup functions in multiple files (e.g., `auth_test.go`, `organization_mgmt_test.go`, `source_create_test.go`).


* 
**Evidence**: Functions like `setupTestDB`, `createTestUser`, and `createTestOrg` are redefined in multiple handler and integration test files.


* 
**Refactoring Recommendation**: Move all shared testing utilities to the `internal/testdb` or a new `internal/testutils` package. Create a "Fixture Factory" that can generate valid models with sensible defaults for any test.



### 6. Masked Errors in API Handlers

* **Status**: **Active**
* **Problem Description**: API handlers sometimes swallow specific service errors and return generic error messages to the client.
* **Why it increases maintenance cost**: It makes debugging production issues significantly harder because the client-side error doesn't match the root cause, and the server logs may not be sufficiently granular.
* 
**Evidence**: Integration tests for handoffs note: *"currently handler masks errors unless we update it too"*.


* **Refactoring Recommendation**: Standardize error mapping. Create a centralized utility that maps internal service errors (e.g., `ErrHandoffExists`) to HTTP status codes and localized error messages.