import React, { createContext, useContext, useState, useCallback } from 'react'
import { X, CheckCircle2, AlertCircle, Info } from 'lucide-react'
import { cn } from '@/lib/utils'

type ToastType = 'success' | 'error' | 'info'

interface Toast {
  id: string
  message: string
  type: ToastType
  duration?: number
}

interface ToastContextType {
  toast: (message: string, type?: ToastType, duration?: number) => void
}

const ToastContext = createContext<ToastContextType | undefined>(undefined)

export const useToast = () => {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context
}

export const ToastProvider = ({ children }: { children: React.ReactNode }) => {
  const [toasts, setToasts] = useState<Toast[]>([])

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const toast = useCallback((message: string, type: ToastType = 'info', duration = 3000) => {
    const id = Math.random().toString(36).substring(2, 9)
    setToasts((prev) => [...prev, { id, message, type, duration }])

    setTimeout(() => {
      removeToast(id)
    }, duration)
  }, [removeToast])

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={cn(
              "pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-xl shadow-xl border transition-all duration-300 animate-in slide-in-from-right-full fade-in",
              t.type === 'success' && "bg-emerald-50 border-emerald-200 text-emerald-800 dark:bg-emerald-900/90 dark:border-emerald-800 dark:text-emerald-100",
              t.type === 'error' && "bg-red-50 border-red-200 text-red-800 dark:bg-red-900/90 dark:border-red-800 dark:text-red-100",
              t.type === 'info' && "bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/90 dark:border-blue-800 dark:text-blue-100",
              "bg-white/95 dark:bg-gray-900/95 border-border text-foreground" // Default fallback
            )}
          >
            {t.type === 'success' && <CheckCircle2 className="w-5 h-5" />}
            {t.type === 'error' && <AlertCircle className="w-5 h-5" />}
            {t.type === 'info' && <Info className="w-5 h-5" />}
            <p className="text-sm font-medium">{t.message}</p>
            <button 
              onClick={() => removeToast(t.id)}
              className="ml-2 text-muted-foreground hover:text-foreground"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
