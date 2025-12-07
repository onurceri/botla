# Plan 2.5: Temperature ve Max Tokens UI

## Özet

Mevcut Temperature ve MaxTokens ayarlarının frontend'de görünür hale getirilmesi.

---

## Mevcut Durum

| Dosya | Mevcut Özellik |
|-------|----------------|
| `internal/models/chatbot.go` | `Temperature`, `MaxTokens` alanları ✅ |
| `internal/rag/openai.go` | Parametreler kullanılıyor ✅ |
| Frontend | UI **yok** |

**Not:** Backend zaten hazır, sadece frontend gerekli.

---

## Uygulama Adımları

### Adım 1: Frontend - Model Ayarları Bileşeni

**Dosya:** `frontend/src/features/chatbot/ModelSettings.tsx` (YENİ veya mevcut settings'e ekle)

**UI:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🤖 Model Ayarları                                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Model:                                                      │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ gpt-4o-mini                                         ▼  │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Temperature:                                   0.7          │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ●─────────────────────●─────────────────────○           │ │
│ │ 0.0           Yaratıcı              Deterministik  2.0 │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ℹ️ Düşük: Daha tutarlı yanıtlar                             │
│    Yüksek: Daha yaratıcı yanıtlar                          │
│                                                             │
│ Max Tokens:                                   2048          │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ●─────────────────────────○                             │ │
│ │ 256                                               8192 │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ℹ️ Yanıttaki maksimum kelime sayısını belirler              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Adım 2: API Güncellemesi (Zaten çalışıyor, doğrulama)

```go
// PATCH /api/chatbots/:id
// Body:
{
    "temperature": 0.7,
    "max_tokens": 2048
}
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `frontend/src/features/chatbot/ModelSettings.tsx` | YENİ/GÜNCELLE | UI bileşeni |
| `frontend/src/features/chatbot/ChatbotSettings.tsx` | GÜNCELLE | Tab/section ekle |

---

## Test Planı

### Manuel Test

1. Chatbot → Ayarlar → Model Ayarları
2. Temperature slider'ını 0.2'ye çek
3. Max tokens'ı 1024 yap
4. Kaydet
5. Chat gönder
6. API loglarında doğru değerlerin kullanıldığını doğrula

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Frontend UI | 2-3 saat |
| Testler | 1 saat |
| **TOPLAM** | **~2-3 gün** |

---

## Bağımlılıklar

**Önceki:** Bağımsız - Backend zaten hazır

**Sonraki:** Bağımsız
