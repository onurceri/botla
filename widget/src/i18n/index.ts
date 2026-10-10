/**
 * Widget UI strings. The language follows the bot's language_code from the
 * public config (or an explicit `lang` embed param), falling back to Turkish.
 */

const tr = {
  openChat: 'Sohbeti aç',
  closeChat: 'Sohbeti kapat',
  unread: (n: number) => `Okunmamış ${n}`,
  newConversation: 'Yeni Konuşma',
  startNewConversation: 'Yeni Konuşma Başlat',
  close: 'Kapat',
  longConversationHint: 'Daha iyi sonuçlar için yeni konuşma başlatabilirsiniz.',
  newShort: 'Yeni',
  inputPlaceholder: 'Mesaj yazın...',
  send: 'Gönder',
  sending: 'Gönderiliyor...',
  resetConfirmMessage: 'Mevcut konuşmanız silinecek.',
  confirmStart: 'Başlat',
  cancel: 'Vazgeç',
  emailSendFailed: 'E-posta gönderilemedi, lütfen tekrar deneyin.',
  handoffReceived: 'Talebiniz alındı!',
  handoffWillContact: 'En kısa sürede sizinle iletişime geçeceğiz.',
  handoffTitle: 'Destek Talebi',
  handoffDescription: 'Size dönüş yapabilmemiz için lütfen e-postanızı paylaşın.',
  emailPlaceholder: 'e-posta@adresiniz.com',
  helpful: 'Yararlı',
  notHelpful: 'Yararlı değil',
  suggestedQuestions: 'Önerilen sorular',
  suggestedQuestionsTitle: 'ÖNERİLEN SORULAR',
  previousQuestion: 'Önceki soru',
  nextQuestion: 'Sonraki soru',
  errorMessage: 'Şu an bir hata oluştu, lütfen tekrar deneyin.',
  emptyResponse: 'Merhaba!',
}

export type Messages = typeof tr

const en: Messages = {
  openChat: 'Open chat',
  closeChat: 'Close chat',
  unread: (n: number) => `${n} unread`,
  newConversation: 'New conversation',
  startNewConversation: 'Start a new conversation',
  close: 'Close',
  longConversationHint: 'Start a new conversation for better results.',
  newShort: 'New',
  inputPlaceholder: 'Type a message...',
  send: 'Send',
  sending: 'Sending...',
  resetConfirmMessage: 'Your current conversation will be deleted.',
  confirmStart: 'Start',
  cancel: 'Cancel',
  emailSendFailed: 'Could not send your email, please try again.',
  handoffReceived: 'Request received!',
  handoffWillContact: "We'll get back to you as soon as possible.",
  handoffTitle: 'Support Request',
  handoffDescription: 'Please share your email so we can get back to you.',
  emailPlaceholder: 'you@example.com',
  helpful: 'Helpful',
  notHelpful: 'Not helpful',
  suggestedQuestions: 'Suggested questions',
  suggestedQuestionsTitle: 'SUGGESTED QUESTIONS',
  previousQuestion: 'Previous question',
  nextQuestion: 'Next question',
  errorMessage: 'Something went wrong, please try again.',
  emptyResponse: 'Hello!',
}

const dictionaries = { tr, en }

export type Lang = keyof typeof dictionaries
export const DEFAULT_LANG: Lang = 'tr'

/** Maps a language code like "en-US", "EN" or "tr_TR" to a supported language, or undefined. */
export function normalizeLang(code?: string | null): Lang | undefined {
  const base = (code || '').trim().toLowerCase().split(/[-_]/)[0]
  return base in dictionaries ? (base as Lang) : undefined
}

/** First supported candidate wins (e.g. embed override, then bot language); defaults to Turkish. */
export function resolveLang(...candidates: Array<string | null | undefined>): Lang {
  for (const c of candidates) {
    const lang = normalizeLang(c)
    if (lang) return lang
  }
  return DEFAULT_LANG
}

export function getMessages(lang: Lang = DEFAULT_LANG): Messages {
  return dictionaries[lang] ?? dictionaries[DEFAULT_LANG]
}

type StringKey = { [K in keyof Messages]: Messages[K] extends string ? K : never }[keyof Messages]

/** Typed lookup for a plain string. */
export function t(key: StringKey, lang: Lang = DEFAULT_LANG): string {
  return getMessages(lang)[key]
}
