# Task 008: OpenAPI Specification Generation

**Priority:** 🟡 Medium (Developer Experience)  
**Phase:** 5 - API Contracts  
**Estimated Time:** 4-5 hours  
**Dependencies:** None  

---

## Problem Statement

Frontend and backend share error codes but there's no formal API contract:
- No type safety for API requests/responses
- Frontend types are manually maintained
- No documentation for API consumers
- Breaking changes discovered at runtime

---

## Objective

Generate OpenAPI 3.0 specification:
1. Document all API endpoints
2. Define request/response schemas
3. Include error codes and descriptions
4. Enable code generation for frontend types

---

## Implementation Details

### Step 1: Install go-swagger or oapi-codegen

Add to development dependencies:

```bash
# Option A: Use swaggo/swag for annotation-based generation
go install github.com/swaggo/swag/cmd/swag@latest

# Option B: Create OpenAPI spec manually (recommended for control)
```

### Step 2: Create OpenAPI Base File

**File:** `api/openapi.yaml` (NEW)

```yaml
openapi: 3.0.3
info:
  title: Botla API
  description: API for Botla chatbot platform
  version: 1.0.0
  contact:
    name: Botla Support
    email: support@botla.co

servers:
  - url: /api/v1
    description: API v1

tags:
  - name: Auth
    description: Authentication endpoints
  - name: Chatbots
    description: Chatbot management
  - name: Sources
    description: Data source management
  - name: Chat
    description: Chat endpoints
  - name: Training
    description: Training job management

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    Error:
      type: object
      required:
        - error
      properties:
        error:
          type: string
          description: Error code
          example: ERR_INVALID_CREDENTIALS
        message:
          type: string
          description: Human-readable error message
          example: Invalid email or password

    # Auth schemas
    LoginRequest:
      type: object
      required:
        - email
        - password
      properties:
        email:
          type: string
          format: email
          example: user@example.com
        password:
          type: string
          format: password
          minLength: 8

    LoginResponse:
      type: object
      properties:
        access_token:
          type: string
        refresh_token:
          type: string
        user:
          $ref: '#/components/schemas/User'

    User:
      type: object
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        full_name:
          type: string
        plan_code:
          type: string
          enum: [free, starter, pro, enterprise]

    # Chatbot schemas  
    Chatbot:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        description:
          type: string
        model:
          type: string
        language_code:
          type: string
        status:
          type: string
          enum: [active, inactive, training]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    CreateChatbotRequest:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        description:
          type: string
        language_code:
          type: string
          default: tr

    # Source schemas
    DataSource:
      type: object
      properties:
        id:
          type: string
          format: uuid
        chatbot_id:
          type: string
          format: uuid
        source_type:
          type: string
          enum: [url, pdf, text]
        source_url:
          type: string
          format: uri
        status:
          type: string
          enum: [pending, processing, completed, failed]
        chunk_count:
          type: integer
        error_message:
          type: string
        created_at:
          type: string
          format: date-time

    # Training Job schemas
    TrainingJob:
      type: object
      properties:
        job_id:
          type: string
          format: uuid
        source_id:
          type: string
          format: uuid
        status:
          type: string
          enum: [pending, running, completed, failed, cancelled]
        current_step:
          type: string
          enum: [fetch_source, parse_content, chunk_text, embed_chunks, store_vectors]
        progress_percent:
          type: integer
          minimum: 0
          maximum: 100
        error_code:
          type: string
        error_message:
          type: string
        started_at:
          type: string
          format: date-time
        completed_at:
          type: string
          format: date-time

    # Chat schemas
    ChatRequest:
      type: object
      required:
        - message
      properties:
        message:
          type: string
          minLength: 1
          maxLength: 4000
        conversation_id:
          type: string
          format: uuid

    ChatResponse:
      type: object
      properties:
        response:
          type: string
        conversation_id:
          type: string
          format: uuid
        message_id:
          type: string
          format: uuid
        sources:
          type: array
          items:
            $ref: '#/components/schemas/SourceCitation'

    SourceCitation:
      type: object
      properties:
        source_id:
          type: string
        title:
          type: string
        url:
          type: string
        relevance:
          type: number

  responses:
    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Forbidden:
      description: Permission denied
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    BadRequest:
      description: Invalid request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'

paths:
  # Auth endpoints
  /auth/register:
    post:
      tags: [Auth]
      summary: Register new user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password, full_name]
              properties:
                email:
                  type: string
                  format: email
                password:
                  type: string
                  minLength: 8
                full_name:
                  type: string
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '409':
          description: Email already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /auth/login:
    post:
      tags: [Auth]
      summary: Login user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'

  # Chatbot endpoints
  /chatbots:
    get:
      tags: [Chatbots]
      summary: List user's chatbots
      security:
        - bearerAuth: []
      responses:
        '200':
          description: List of chatbots
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Chatbot'
        '401':
          $ref: '#/components/responses/Unauthorized'
    post:
      tags: [Chatbots]
      summary: Create chatbot
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateChatbotRequest'
      responses:
        '201':
          description: Chatbot created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Chatbot'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /chatbots/{id}:
    get:
      tags: [Chatbots]
      summary: Get chatbot by ID
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Chatbot details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Chatbot'
        '404':
          $ref: '#/components/responses/NotFound'

  # Source endpoints
  /chatbots/{id}/sources:
    get:
      tags: [Sources]
      summary: List sources for chatbot
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: List of sources
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/DataSource'
    post:
      tags: [Sources]
      summary: Add source to chatbot
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                source_type:
                  type: string
                  enum: [url, pdf, text]
                source_url:
                  type: string
                  format: uri
                file:
                  type: string
                  format: binary
                text:
                  type: string
      responses:
        '201':
          description: Source created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: uuid
        '400':
          $ref: '#/components/responses/BadRequest'
        '403':
          $ref: '#/components/responses/Forbidden'
        '409':
          description: Duplicate URL or content
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /sources/{id}/job:
    get:
      tags: [Training]
      summary: Get training job status
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Job status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TrainingJob'
        '404':
          $ref: '#/components/responses/NotFound'

  /sources/{id}/job/retry:
    post:
      tags: [Training]
      summary: Retry failed job
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '202':
          description: Job queued for retry
          content:
            application/json:
              schema:
                type: object
                properties:
                  job_id:
                    type: string
                  message:
                    type: string
        '400':
          description: Job not in failed state
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  # Public chat endpoint
  /public/chat/{chatbot_id}:
    post:
      tags: [Chat]
      summary: Send chat message (public)
      parameters:
        - name: chatbot_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ChatRequest'
      responses:
        '200':
          description: Chat response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ChatResponse'
        '404':
          $ref: '#/components/responses/NotFound'
```

### Step 3: Add OpenAPI Serving Endpoint

**File:** `internal/api/handlers/openapi.go` (NEW)

```go
package handlers

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiSpec embed.FS

// ServeOpenAPI serves the OpenAPI specification
func ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	spec, err := openapiSpec.ReadFile("openapi.yaml")
	if err != nil {
		http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(spec)
}
```

**File:** `internal/api/router/router.go` (MODIFY)

```go
// Add OpenAPI endpoint
mux.HandleFunc("GET /api/openapi.yaml", handlers.ServeOpenAPI)
```

### Step 4: Generate TypeScript Types

Add a script to generate TypeScript types from OpenAPI:

**File:** `frontend/scripts/generate-types.sh` (NEW)

```bash
#!/bin/bash

# Install openapi-typescript if not present
if ! command -v npx openapi-typescript &> /dev/null; then
    npm install -D openapi-typescript
fi

# Generate types
npx openapi-typescript http://localhost:8080/api/openapi.yaml \
  -o src/types/api.generated.ts

echo "Types generated successfully!"
```

### Step 5: Add npm Script

**File:** `frontend/package.json` (MODIFY)

```json
{
  "scripts": {
    "generate-types": "bash scripts/generate-types.sh"
  }
}
```

---

## Tests to Write

### Validation Test

**File:** `api/openapi_test.go` (NEW)

```go
package api

import (
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPISpec_Valid(t *testing.T) {
	spec, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	// Validate the spec
	if err := doc.Validate(loader.Context); err != nil {
		t.Errorf("spec validation failed: %v", err)
	}
}

func TestOpenAPISpec_HasRequiredPaths(t *testing.T) {
	spec, _ := os.ReadFile("openapi.yaml")
	loader := openapi3.NewLoader()
	doc, _ := loader.LoadFromData(spec)

	requiredPaths := []string{
		"/auth/login",
		"/auth/register",
		"/chatbots",
		"/chatbots/{id}",
		"/chatbots/{id}/sources",
		"/sources/{id}/job",
	}

	for _, path := range requiredPaths {
		if doc.Paths.Find(path) == nil {
			t.Errorf("missing required path: %s", path)
		}
	}
}
```

---

## Verification Steps

1. **Validate OpenAPI spec:**
   ```bash
   # Install validator
   npm install -g @redocly/cli
   
   # Validate
   redocly lint api/openapi.yaml
   ```

2. **View documentation:**
   ```bash
   # Start server
   make be-run
   
   # Access spec
   curl http://localhost:8080/api/openapi.yaml
   
   # View with Swagger UI (optional)
   npx @redocly/cli preview-docs api/openapi.yaml
   ```

3. **Generate frontend types:**
   ```bash
   cd frontend
   npm run generate-types
   ```

---

## Acceptance Criteria

- [ ] OpenAPI spec is valid
- [ ] All public endpoints are documented
- [ ] Request/response schemas defined
- [ ] Error codes documented
- [ ] Spec served at /api/openapi.yaml
- [ ] TypeScript types can be generated
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `api/openapi.yaml` | CREATE |
| `internal/api/handlers/openapi.go` | CREATE |
| `internal/api/router/router.go` | MODIFY |
| `api/openapi_test.go` | CREATE |
| `frontend/scripts/generate-types.sh` | CREATE |
| `frontend/package.json` | MODIFY |
