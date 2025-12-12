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

---

### 4.2.4 Total Storage Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Upload files totaling 10MB | All succeed |
| 2 | Free: Upload 1 more MB | 403 Forbidden |

---

### 4.2.5 File Hash Computation
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload PDF | 201 Created |
| 2 | Query source | hash field populated |

---

### 4.2.6 Original Filename Stored
**Priority:** Low  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload "my-document.pdf" | 201 Created |
| 2 | Query source | original_filename = "my-document.pdf" |

---

### 4.2.7 Duplicate File Prevention
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload PDF A | 201 Created |
| 2 | Upload same PDF A again | 409 Conflict (same hash) |

---

### 4.2.8 Unsupported File Type
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload .exe file | 400 Bad Request |
| 2 | Upload .js file | 400 Bad Request |

---

### 4.2.9 OCR Based on Plan
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Upload image-based PDF | OCR not applied |
| 2 | Pro: Upload image-based PDF | OCR applied, text extracted |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "FileSource|Upload"
```
