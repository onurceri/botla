# 4.2 PDF/File Source Creation Test Plan

## Overview
This test plan covers file upload functionality including size limits and storage tracking.

---

## Test Cases

### 4.2.1 Upload Valid PDF
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST multipart/form-data with PDF | 201 Created |
| 2 | Source status | "pending" |
| 3 | File stored in S3/storage | File exists |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. POST valid PDF file.
  2. Verify 201.
  3. Verify status "pending".

---

### 4.2.2 File Size Limit Enforcement
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Upload 5MB PDF | 201 Created |
| 2 | Free: Upload 6MB PDF | 413 Payload Too Large |
| 3 | Pro: Upload 20MB PDF | 201 Created |
| 4 | Pro: Upload 21MB PDF | 413 Payload Too Large |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Free User, Pro User.
- **Steps:**
  1. Free: Upload 6MB -> 413.
  2. Pro: Upload 21MB -> 413.

---

### 4.2.3 Files Per Bot Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Upload 1 PDF | 201 Created |
| 2 | Free: Upload 2nd PDF | 403 Forbidden |
| 3 | Pro: Upload 20 PDFs | All succeed |
| 4 | Pro: Upload 21st PDF | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Free User, Pro User.
- **Steps:**
  1. Free: Upload 1 -> 201. Upload 2nd -> 403.
  2. Pro: Upload 20 -> 201. Upload 21st -> 403.

---

### 4.2.4 Total Storage Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Upload files totaling 10MB | All succeed |
| 2 | Free: Upload 1 more MB | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Free User.
  - Manually set storage usage to 10MB.
- **Steps:**
  1. Upload 1MB file. Expect 403/402.

---

### 4.2.5 File Hash Computation
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload PDF | 201 Created |
| 2 | Query source | hash field populated |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Upload PDF.
  2. Query DB for `hash` column. Verify non-empty.

---

### 4.2.6 Original Filename Stored
**Priority:** Low  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload "my-document.pdf" | 201 Created |
| 2 | Query source | original_filename = "my-document.pdf" |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Upload `test.pdf`.
  2. Verify DB `original_filename` is `test.pdf`.

---

### 4.2.7 Duplicate File Prevention
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload PDF A | 201 Created |
| 2 | Upload same PDF A again | 409 Conflict (same hash) |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Upload file A. Expect 201.
  2. Upload file A again. Expect 409.

---

### 4.2.8 Unsupported File Type
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload .exe file | 400 Bad Request |
| 2 | Upload .js file | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Steps:**
  1. Upload `test.exe`. Expect 400.
  2. Upload `test.js`. Expect 400.

---

### 4.2.9 OCR Based on Plan
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Upload image-based PDF | OCR not applied |
| 2 | Pro: Upload image-based PDF | OCR applied, text extracted |

**Implementation Plan:**
- **Test File:** `internal/integration/source_file_upload_test.go`
- **Setup:**
  - Free User, Pro User.
- **Steps:**
  1. Free: Upload image PDF. Verify empty/minimal text.
  2. Pro: Upload image PDF. Verify text extracted (mocked or real depending on capability).

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "FileSource|Upload"
```
