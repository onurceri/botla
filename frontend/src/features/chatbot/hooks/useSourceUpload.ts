import { useState, useRef } from 'react'
import { useToast } from '@/components/ui/toast'

export function useSourceUpload({ onUploadPDF, onUploadURL, onUploadText, maxFileSizeMB = 50 }: {
  onUploadPDF: (file: File) => Promise<void>
  onUploadURL: (url: string) => Promise<void>
  onUploadText: (text: string) => Promise<void>
  maxFileSizeMB?: number
}) {
  const { toast } = useToast()
  const [activeMode, setActiveMode] = useState<'pdf' | 'url' | 'text' | null>(null)
  const [loading, setLoading] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const isValidHttpUrl = (u: string) => {
    try {
      const parsed = new URL(u)
      return parsed.protocol === 'http:' || parsed.protocol === 'https:'
    } catch { return false }
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0]
      if (file.type !== 'application/pdf') {
        toast('Yalnızca PDF dosyaları desteklenir.', 'error')
        if (fileInputRef.current) fileInputRef.current.value = ''
        return
      }
      const max = maxFileSizeMB * 1024 * 1024
      if (file.size > max) {
        toast(`Dosya boyutu ${maxFileSizeMB}MB'den büyük olamaz.`, 'error')
        if (fileInputRef.current) fileInputRef.current.value = ''
        return
      }
      setLoading(true)
      try {
        await onUploadPDF(file)
        toast('PDF başarıyla yüklendi. İşleniyor...', 'success')
        setActiveMode(null)
      } catch (error: any) {
        const msg = error?.response?.data?.message || 'PDF yüklenirken bir hata oluştu.'
        toast(msg, 'error')
      } finally {
        setLoading(false)
        if (fileInputRef.current) fileInputRef.current.value = ''
      }
    }
  }

  const handleSubmit = async () => {
    if (!inputValue.trim()) return
    setLoading(true)
    try {
      if (activeMode === 'url') {
        if (!isValidHttpUrl(inputValue.trim())) {
          toast('Lütfen geçerli bir URL girin.', 'error')
          return
        }
        await onUploadURL(inputValue)
        toast('Web sitesi kaynağı eklendi. Taranıyor...', 'success')
      }
      if (activeMode === 'text') {
        await onUploadText(inputValue)
        toast('Metin kaynağı eklendi.', 'success')
      }
      setActiveMode(null)
      setInputValue('')
    } catch (error: any) {
      const msg = error?.response?.data?.message || 'Kaynak eklenirken bir hata oluştu.'
      toast(msg, 'error')
    } finally {
      setLoading(false)
    }
  }

  return {
    activeMode,
    setActiveMode,
    loading,
    inputValue,
    setInputValue,
    fileInputRef,
    handleFileChange,
    handleSubmit,
  }
}

