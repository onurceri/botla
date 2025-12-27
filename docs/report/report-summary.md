Bu sistem:
Ciddi ve doğru kurgulanmış
MVP’yi geçmiş
Ölçeklenmeye yakın ama henüz hazır değil
En yüksek kaldıraç:
Async training
Service parçalama
Contract + test
EN KRİTİK 5 REFACTOR — TEKNİK ROADMAP
Kriterler:
Etki / efor oranı yüksek
Birbirini açan adımlar
Üretimde risk azaltan

🥇 1. Async Training Mimarisine Geçiş (FOUNDATIONAL)
Problem
Training pipeline (fetch → parse → chunk → embed) synchronous
Request lifecycle’a bağlı
Timeout / ölçeklenme riski
Refactor
Training’i request’ten kopar
Job-based asenkron akışa taşı
Kapsam
API
Service layer
Infra
Neden #1?
Bunu yapmadan performans, reliability ve UX düzelmez.

🥈 2. Service Layer Parçalama (Use-case Oriented)
Problem
Tek servis çok fazla sorumluluk üstleniyor
Test yazmak zor
Değişiklik riski yüksek
Refactor
TrainingOrchestrator
SourceIngestionService
EmbeddingService
GuardrailService
ProviderSelector

Kazanç
İzole test
Net sorumluluk
Daha güvenli refactor

🥉 3. Contract-Driven API (Frontend ↔ Backend)
Problem
Error code’lar güçlü ama “implicit”
Tip güvenliği yok
Refactor
OpenAPI spec
Paylaşılan enum’lar
Contract test
Kazanç
Frontend/backend bağımsız deploy
Kırılmalar erken yakalanır

🏅 4. Training Idempotency & Deduplication
Problem
Aynı kaynak tekrar işlenebilir
Refactor
Source hash
Job uniqueness
Safe retry
Kazanç
Daha az maliyet
Daha az veri çöpü

🎖️ 5. Observability Temeli (Minimum Viable)
Problem
Prod debug zor
Refactor
Request ID
Job ID
Structured logs
Basic metrics
Kazanç
Gece uykusu
Hızlı müdahale

⚙️ ASYNC TRAINING MİMARİSİ (NET ÇİZİM)
Aşağıda framework bağımsız, ama production-grade bir model var.

🧱 1️⃣ Yeni Temel Kavramlar
🧩 TrainingJob (Domain)
id
bot_id
status: PENDING | RUNNING | FAILED | COMPLETED
current_step
error_code?
created_at
updated_at
🧩 TrainingStep
FETCH_SOURCE
PARSE_CONTENT
CHUNK_TEXT
EMBED_CHUNKS
STORE_VECTORS

🔌 2️⃣ API Tasarımı
POST /bots/{id}/train
→ returns 202 Accepted
{
  job_id
}
GET /training-jobs/{job_id}
→ status, progress, error
(Opsiyonel) Webhook / SSE
Frontend progress takibi için

🧠 3️⃣ Backend Akış (Adım Adım)
Client
  ↓
API Controller
  ↓
TrainingJobService.create()
  ↓
JobQueue.enqueue(job_id)
  ↓
(Async Worker)
     ↓
  TrainingOrchestrator.run(job_id)
     ↓
  Step-by-step execution
     ↓
  State persist

⚙️ 4️⃣ Worker İç Yapısı
Her step idempotent olmalı:
for step in steps:
  if step already completed:
    continue
  try:
    execute(step)
    mark step complete
  except:
    mark job failed
    break

➡️ Retry güvenli
➡️ Partial failure tolere edilebilir

🏗️ 5️⃣ Infra Seçenekleri (Basit → Güçlü)
MVP
Redis queue
Background worker
DB polling
Scale
SQS / PubSub
Dedicated workers
Horizontal scaling

🔐 6️⃣ Güvenlik & Limitler
Max source size
Max concurrent jobs per bot
Rate limit
Timeout per step

🧪 7️⃣ Test Stratejisi (Async’e Özel)
Step unit test
Job state transition test
Retry simulation
Failure recovery test

🧠 Net Tavsiye (Koç Yorumu)
Sıralamayı bozma.
1️⃣ Async training
2️⃣ Service parçalama
3️⃣ Contract
4️⃣ Idempotency
5️⃣ Observability
Bunları yaparsan:
Sistem rahatlar
Yeni feature eklemek hızlanır
Widget eklemek çocuk oyuncağı olur