# Implementation Tasks - Botla System Improvements

Bu klasör, sistem analiz raporundaki tüm bulguları çözmek için oluşturulmuş detaylı task dosyalarını içerir.

## 📊 Genel Bakış

| Toplam Task | Tahmini Süre | Öncelik Dağılımı |
|-------------|--------------|------------------|
| 15 | ~50 saat | 🔴 Critical: 4 / 🟡 High: 7 / 🟢 Low: 4 |

---

## 🗂️ Fazlara Göre Tasklar

### Phase 1: Observability Foundation
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 001 | Request-ID Middleware | [001-request-id-middleware.md](./001-request-id-middleware.md) | 2-3h | 🔴 |
| 002 | Job State Table | [002-job-state-table.md](./002-job-state-table.md) | 2-3h | 🔴 |
| 003 | Job Progress API | [003-job-progress-api.md](./003-job-progress-api.md) | 2-3h | 🔴 |

### Phase 2: Async Training Improvements
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 004 | Integrate Job Tracking | [004-integrate-job-tracking.md](./004-integrate-job-tracking.md) | 3-4h | 🔴 |
| 005 | Step-Level Retry | [005-step-retry-mechanism.md](./005-step-retry-mechanism.md) | 3-4h | 🟡 |
| 009 | Worker Pool | [009-worker-pool.md](./009-worker-pool.md) | 4-5h | 🟡 |

### Phase 3: Idempotency & Deduplication
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 006 | PDF Deduplication | [006-pdf-deduplication.md](./006-pdf-deduplication.md) | 2-3h | 🟡 |

### Phase 4: Security Hardening
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 007 | SSRF Protection | [007-ssrf-protection.md](./007-ssrf-protection.md) | 3-4h | 🔴 |

### Phase 5: API Contracts
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 008 | OpenAPI Specification | [008-openapi-spec.md](./008-openapi-spec.md) | 4-5h | 🟡 |

### Phase 6: Service Layer Refactoring
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 015 | Frontend Domain Layer | [015-frontend-domain-layer.md](./015-frontend-domain-layer.md) | 4-5h | 🟢 |

### Phase 7: Test Coverage
| # | Task | Dosya | Süre | Öncelik |
|---|------|-------|------|---------|
| 010 | Async Pipeline Tests | [010-async-pipeline-tests.md](./010-async-pipeline-tests.md) | 4-5h | 🟡 |
| 011 | Multi-Tenant Tests | [011-multi-tenant-tests.md](./011-multi-tenant-tests.md) | 3-4h | 🔴 |
| 012 | Rate Limit Tests | [012-rate-limit-tests.md](./012-rate-limit-tests.md) | 2-3h | 🟡 |
| 013 | Frontend E2E Tests | [013-frontend-e2e-tests.md](./013-frontend-e2e-tests.md) | 4-5h | 🟢 |
| 014 | Widget Edge Case Tests | [014-widget-edge-case-tests.md](./014-widget-edge-case-tests.md) | 3-4h | 🟢 |

---

## 🎯 Önerilen Çalışma Sırası

**Kritik yolun öncelik sırası:**

```
001 → 002 → 003 → 004 → 005 → 007 → 006 → 008 → 009 → 010 → 011 → 012 → 013 → 014 → 015
```

### Sprint 1: Foundation (Hafta 1)
- [x] 001 - Request-ID Middleware
- [x] 002 - Job State Table  
- [x] 003 - Job Progress API
- [x] 004 - Integrate Job Tracking

### Sprint 2: Reliability (Hafta 2)
- [x] 005 - Step Retry Mechanism
- [x] 006 - PDF Deduplication
- [x] 007 - SSRF Protection

### Sprint 3: Scale & Quality (Hafta 3)
- [x] 008 - OpenAPI Spec
- [x] 009 - Worker Pool
- [x] 010 - Async Pipeline Tests
- [x] 011 - Multi-Tenant Tests

### Sprint 4: Polish (Hafta 4)
- [ ] 012 - Rate Limit Tests
- [ ] 013 - Frontend E2E Tests
- [ ] 014 - Widget Edge Case Tests
- [ ] 015 - Frontend Domain Layer

---

## ✅ Her Task'ın İçerdiği Bilgiler

Her task dosyası şunları içerir:
- **Problem Statement**: Ne sorun var
- **Objective**: Ne yapılacak
- **Implementation Details**: Nasıl yapılacak (kod örnekleri)
- **Tests to Write**: Yazılması gereken testler
- **Verification Steps**: Nasıl doğrulanacak
- **Acceptance Criteria**: Kabul kriterleri
- **Files Changed**: Değişen dosyalar listesi

---

## 🚀 Başlamadan Önce

1. Bu README'yi okuyun
2. İlk task'ı (001) açın
3. Sırayla ilerleyin
4. Her task'tan sonra testleri çalıştırın
5. Acceptance criteria'yı kontrol edin

**Sorularınız için:** Her task bağımsız, herhangi bir developer okuyabilir.
