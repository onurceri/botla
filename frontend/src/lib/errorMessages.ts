import axios from 'axios'

type ApiErrorResponse = {
  error?: string
  code?: string
  message?: string
  details?: unknown
}

export const AUTO_SAVE_DEFAULT_FALLBACK = 'Kaydetme başarısız'
export const AUTO_SAVE_RETRY_SUFFIX = ' - Tekrar deneniyor...'

export const SAVE_INDICATOR_MESSAGES = {
  saving: 'Kaydediliyor...',
  saved: 'Kaydedildi',
  retrying: 'Tekrar deneniyor...',
  failed: 'Kaydedilemedi',
} as const

const exactMap: Record<string, string> = {
  BAD_REQUEST: 'Geçersiz istek.',
  UNAUTHORIZED: 'Yetkisiz işlem. Lütfen tekrar giriş yapın.',
  FORBIDDEN: 'Bu işlem için yetkiniz yok.',
  NOT_FOUND: 'İstenen kayıt bulunamadı.',
  CONFLICT: 'Bu işlem çakışmaya neden oldu.',
  TOO_MANY_REQUESTS: 'Çok fazla istek. Lütfen daha sonra tekrar deneyin.',
  PAYMENT_REQUIRED: 'Bu özellik için planınızı yükseltmeniz gerekiyor.',
  INTERNAL_ERROR: 'Sunucu hatası. Lütfen tekrar deneyin.',
  SERVICE_UNAVAILABLE: 'Servis şu anda kullanılamıyor. Lütfen tekrar deneyin.',
  METHOD_NOT_ALLOWED: 'Bu işlem desteklenmiyor.',
  REQUEST_ENTITY_TOO_LARGE: 'İstek boyutu çok büyük.',

  ERR_MONTHLY_TOKENS_EXCEEDED: 'Aylık token limitinize ulaştınız.',
  ERR_NAME_AND_ACTION_TYPE_REQUIRED: 'Aksiyon adı ve tipi zorunludur.',
  ERR_PDF_LIMIT_REACHED: 'PDF limitinize ulaştınız.',
  ERR_FILE_TOO_LARGE: 'Dosya boyutu çok büyük.',
  ERR_READD_COOLDOWN_ACTIVE: 'Bu kaynağı tekrar eklemek için bir süre bekleyin.',
  ERR_DUPLICATE_URL: 'Bu URL zaten eklenmiş.',
  ERR_ONLY_URL_REFRESH: 'Bu işlem yalnızca URL kaynakları için geçerlidir.',
  ERR_SOURCE_ALREADY_PROCESSING: 'Kaynak zaten işleniyor.',
  ERR_PLAN_REFRESH_UNAVAILABLE: 'Yenileme özelliği planınızda aktif değil.',
  ERR_MONTHLY_REFRESH_EXCEEDED: 'Aylık yenileme limitinize ulaştınız.',
  ERR_REFRESH_COOLDOWN_ACTIVE: 'Yenileme için bekleme süresi aktif.',
  ERR_INVALID_REQUEST_BODY: 'Geçersiz istek içeriği.',
  ERR_NO_URLS_PROVIDED: 'URL listesi boş olamaz.',
  ERR_URL_LIMIT_REACHED: 'URL limitinize ulaştınız.',
  ERR_MONTHLY_INGESTION_EXCEEDED: 'Aylık içerik işleme limitinize ulaştınız.',
  ERR_SITEMAP_PARSE_FAILED: 'Sitemap okunamadı.',
  ERR_INVALID_STATUS: 'Geçersiz durum.',
  ERR_MAX_CHATBOTS_EXCEEDED: 'Bot limitinize ulaştınız.',
  ERR_TEXT_TOO_LONG: 'Metin çok uzun.',

  internal_error: 'Sunucu hatası. Lütfen tekrar deneyin.',
  unauthorized: 'Yetkisiz işlem. Lütfen tekrar giriş yapın.',
  method_not_allowed: 'Bu işlem desteklenmiyor.',
  invalid_chatbot_id: 'Geçersiz bot kimliği.',
  chatbot_not_found: 'Chatbot bulunamadı.',

  'No refresh token': 'Oturum süresi doldu. Lütfen tekrar giriş yapın.',
  'invalid request body': 'Geçersiz istek içeriği.',
  'valid email is required': 'Geçerli bir e-posta adresi gereklidir.',
  'failed to save email': 'E-posta kaydedilemedi.',
  'failed to create handoff request': 'Destek talebi oluşturulamadı.',

  'cannot delete the last workspace in the organization':
    'Organizasyondaki son çalışma alanı silinemez',
  'cannot delete the last organization': 'Son organizasyon silinemez',
  'cannot remove the last owner': 'Son sahip çıkarılamaz',
}

const includesMap: Array<{ includes: string; tr: string }> = [
  {
    includes: 'cannot demote the last owner',
    tr: 'Organizasyonun son sahibinin rolü değiştirilemez',
  },
  { includes: 'cannot promote yourself', tr: 'Kendi rolünüzü yükseltemezsiniz' },
  {
    includes: 'only owners can assign owner role',
    tr: 'Sahip rolünü yalnızca sahipler atayabilir',
  },
  { includes: 'invalid role', tr: 'Geçersiz rol seçildi' },
]

export function translateKnownErrorMessage(raw: string): string | null {
  const msg = raw.trim()
  if (!msg) return null

  const exact = exactMap[msg]
  if (exact) return exact

  const lower = msg.toLowerCase()
  const lowerExact = exactMap[lower]
  if (lowerExact) return lowerExact

  for (const rule of includesMap) {
    if (lower.includes(rule.includes)) return rule.tr
  }

  return null
}

function coerceApiErrorResponse(data: unknown): ApiErrorResponse | null {
  if (!data || typeof data !== 'object') return null
  const d = data as Record<string, unknown>
  return {
    error: typeof d.error === 'string' ? d.error : undefined,
    code: typeof d.code === 'string' ? d.code : undefined,
    message: typeof d.message === 'string' ? d.message : undefined,
    details: d.details,
  }
}

function getAxiosLikeResponseData(err: unknown): unknown | null {
  if (!err || typeof err !== 'object') return null
  const e = err as Record<string, unknown>
  const response = e.response
  if (!response || typeof response !== 'object') return null
  const r = response as Record<string, unknown>
  return r.data ?? null
}

export function getTurkishErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data

    if (typeof data === 'string') {
      return translateKnownErrorMessage(data) || data || fallback
    }

    const apiErr = coerceApiErrorResponse(data)
    const code = apiErr?.code
    if (code) {
      const translatedCode = translateKnownErrorMessage(code)
      if (translatedCode) return translatedCode
      return fallback
    }

    const msg = apiErr?.error || apiErr?.message
    if (msg) return translateKnownErrorMessage(msg) || msg

    const translatedGeneric = translateKnownErrorMessage(err.message)
    if (translatedGeneric) return translatedGeneric

    return fallback
  }

  const axiosLikeData = getAxiosLikeResponseData(err)
  if (axiosLikeData !== null) {
    if (typeof axiosLikeData === 'string') {
      return translateKnownErrorMessage(axiosLikeData) || axiosLikeData || fallback
    }

    const apiErr = coerceApiErrorResponse(axiosLikeData)
    const code = apiErr?.code
    if (code) {
      const translatedCode = translateKnownErrorMessage(code)
      if (translatedCode) return translatedCode
      return fallback
    }

    const msg = apiErr?.error || apiErr?.message
    if (msg) return translateKnownErrorMessage(msg) || msg
  }

  if (err instanceof Error) {
    return translateKnownErrorMessage(err.message) || fallback
  }

  if (typeof err === 'string') {
    return translateKnownErrorMessage(err) || err || fallback
  }

  return fallback
}
