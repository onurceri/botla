import { useState, useEffect, useRef } from 'react'
import {
  Terminal,
  ChevronUp,
  ChevronDown,
  Trash2,
  Info,
  AlertCircle,
  MessageSquare,
  Bot,
  Cpu,
} from 'lucide-react'

type LogEntry = {
  id: string
  type: 'info' | 'error' | 'message_sent' | 'response_received' | 'config_loaded' | 'handoff'
  message: string
  payload?: any
  timestamp: Date
}

export default function PlaygroundConsole() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isExpanded, setIsExpanded] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const { type, payload } = event.data
      if (typeof type !== 'string' || !type.startsWith('WIDGET_EVENT_')) return

      const eventType = type.replace('WIDGET_EVENT_', '').toLowerCase()

      const newEntry: LogEntry = {
        id: Math.random().toString(36).substring(2, 11),
        type: eventType as any,
        message: getMessageForType(eventType, payload),
        payload,
        timestamp: new Date(),
      }

      setLogs((prev) => [...prev, newEntry].slice(-50)) // Keep last 50 logs
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [])

  useEffect(() => {
    if (scrollRef.current && isExpanded) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs, isExpanded])

  const getMessageForType = (type: string, payload: any): string => {
    switch (type) {
      case 'config_loaded':
        return 'Widget yapılandırması başarıyla yüklendi.'
      case 'message_sent':
        return `Kullanıcı mesajı: "${payload.content}"`
      case 'response_received':
        return `Bot yanıtı: "${payload.content}"`
      case 'error':
        return `Hata oluştu: ${payload.message}`
      case 'handoff':
        return `Canlı desteğe yönlendirme tetiklendi.`
      default:
        return `${type} olayı tetiklendi.`
    }
  }

  const getIconForType = (type: string) => {
    switch (type) {
      case 'config_loaded':
        return <Cpu className="w-3.5 h-3.5 text-blue-400" />
      case 'message_sent':
        return <MessageSquare className="w-3.5 h-3.5 text-slate-400" />
      case 'response_received':
        return <Bot className="w-3.5 h-3.5 text-emerald-400" />
      case 'error':
        return <AlertCircle className="w-3.5 h-3.5 text-red-400" />
      case 'handoff':
        return <Info className="w-3.5 h-3.5 text-amber-400" />
      default:
        return <Terminal className="w-3.5 h-3.5 text-slate-400" />
    }
  }

  const clearLogs = () => setLogs([])

  return (
    <div
      className={`shrink-0 bg-slate-950/95 lg:bg-slate-950/90 backdrop-blur-xl text-slate-300 transition-all duration-500 cubic-bezier(0.4, 0, 0.2, 1) z-30 border-t border-slate-800/50 rounded-b-2xl absolute bottom-0 left-0 right-0 ${isExpanded ? 'h-64 sm:h-80 shadow-[0_-20px_50px_-12px_rgba(0,0,0,0.5)]' : 'h-10 lg:h-11'}`}
      data-testid="playground-console-container"
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-3 lg:px-6 h-10 lg:h-11 cursor-pointer hover:bg-slate-800/30 transition-colors group"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center gap-2 lg:gap-3">
          <div
            className={`p-1 rounded bg-slate-800 transition-colors ${isExpanded ? 'text-emerald-400' : 'text-slate-400 group-hover:text-emerald-400'}`}
          >
            <Terminal className="w-3 h-3 lg:w-3.5 lg:h-3.5" />
          </div>
          <span className="text-[9px] lg:text-[11px] font-mono font-bold tracking-[0.1em] uppercase text-slate-400 group-hover:text-slate-200 transition-colors truncate">
            Hata Ayıklama Konsolu
          </span>
          {logs.length > 0 && !isExpanded && (
            <div className="flex items-center gap-1 ml-1 lg:ml-2 bg-emerald-500/10 px-1.5 lg:px-2 py-0.5 rounded-full border border-emerald-500/20">
              <span className="flex h-1 w-1 lg:h-1.5 lg:w-1.5 rounded-full bg-emerald-500 animate-pulse" />
              <span className="text-[9px] lg:text-[10px] text-emerald-400 font-bold">
                {logs.length}
              </span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2 lg:gap-4">
          {isExpanded && (
            <button
              onClick={(e) => {
                e.stopPropagation()
                clearLogs()
              }}
              className="p-1.5 hover:bg-slate-800 hover:text-red-400 rounded-md transition-all text-slate-500"
              title="Konsolu Temizle"
            >
              <Trash2 className="w-3 h-3 lg:w-3.5 lg:h-3.5" />
            </button>
          )}
          <div className="text-slate-500 group-hover:text-slate-300 transition-colors">
            {isExpanded ? (
              <ChevronDown className="w-3.5 h-3.5 lg:w-4 lg:h-4" />
            ) : (
              <ChevronUp className="w-3.5 h-3.5 lg:w-4 lg:h-4" />
            )}
          </div>
        </div>
      </div>

      {/* Logs Content */}
      {isExpanded && (
        <div
          ref={scrollRef}
          className="p-4 lg:p-6 h-[calc(100%-40px)] lg:h-[calc(100%-44px)] overflow-y-auto font-mono text-[10px] lg:text-[11px] leading-relaxed space-y-3 selection:bg-emerald-500/30 scrollbar-thin scrollbar-thumb-slate-800 scrollbar-track-transparent"
        >
          {logs.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-slate-600 gap-3">
              <Cpu className="w-8 h-8 opacity-20" />
              <p className="italic text-[12px]">
                Henüz olay kaydedilmedi. Widget ile etkileşime geçin.
              </p>
            </div>
          ) : (
            <div className="space-y-2.5">
              {logs.map((log) => (
                <div
                  key={log.id}
                  className="flex gap-4 group/log animate-in fade-in slide-in-from-left-2 duration-300"
                >
                  <span className="text-slate-600 shrink-0 select-none font-medium opacity-70">
                    [
                    {log.timestamp.toLocaleTimeString([], {
                      hour12: false,
                      hour: '2-digit',
                      minute: '2-digit',
                      second: '2-digit',
                    })}
                    ]
                  </span>
                  <div className="flex items-start gap-2.5 min-w-0">
                    <span className="mt-0.5 shrink-0 opacity-80 group-hover/log:opacity-100 transition-opacity">
                      {getIconForType(log.type)}
                    </span>
                    <div className="flex flex-col gap-1 min-w-0">
                      <span
                        className={`break-words font-medium ${
                          log.type === 'error'
                            ? 'text-red-400'
                            : log.type === 'response_received'
                              ? 'text-emerald-400'
                              : log.type === 'message_sent'
                                ? 'text-slate-200'
                                : 'text-slate-300'
                        }`}
                      >
                        {log.message}
                      </span>
                      {log.payload && Object.keys(log.payload).length > 0 && (
                        <div className="hidden group-hover/log:block mt-1 p-2 bg-slate-900/50 rounded border border-slate-800/50 text-[10px] text-slate-500 overflow-x-auto max-w-xl">
                          <pre>{JSON.stringify(log.payload, null, 2)}</pre>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
