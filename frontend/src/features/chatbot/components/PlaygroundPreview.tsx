import { WidgetApp } from '@widget/widgetApp'
import styles from '@widget/styles.css?raw'

type Props = {
  id: string
  themeColor: string
  chatHeaderColor: string
  chatHeaderTextColor: string
  botMessageColor: string
  botMessageTextColor: string
  userMessageColor: string
  userMessageTextColor: string
  chatFontFamily: string
  position: string
  botDisplayName: string
  botIcon: string
  chatBackgroundColor: string
  welcomeMessage: string
  previewOpen: boolean
  sessionId: string
  suggestionsEnabled: boolean
  suggestedQuestions: string[]
  refreshKey?: number
  hideBranding?: boolean
  customBranding?: { logo_url?: string; text?: string; link?: string } | null
}

export default function PlaygroundPreview({
  id,
  themeColor,
  chatHeaderColor,
  chatHeaderTextColor,
  botMessageColor,
  botMessageTextColor,
  userMessageColor,
  userMessageTextColor,
  chatFontFamily,
  position,
  botDisplayName,
  botIcon,
  chatBackgroundColor,
  welcomeMessage,
  previewOpen,
  sessionId,
  suggestionsEnabled,
  suggestedQuestions,
  refreshKey,
  hideBranding,
  customBranding,
}: Props) {
  return (
    <div className="flex-1 flex flex-col bg-background border border-border rounded-xl shadow-2xl overflow-hidden min-h-[500px]">
      <style>{styles}</style>
      <div className="h-10 bg-white/60 backdrop-blur border-b border-border flex items-center px-4 gap-4">
        <div className="flex gap-2">
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
        </div>
        <div className="flex-1 flex justify-center">
          <div className="bg-background border border-border rounded-md px-3 py-1 text-xs text-muted-foreground w-64 text-center flex items-center justify-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-500" />
            example.com
          </div>
        </div>
        <div className="w-16" />
      </div>
      <div className="flex-1 relative bg-slate-50" style={{ backgroundImage: 'radial-gradient(#cbd5e1 1px, transparent 1px)', backgroundSize: '20px 20px' }}>
        <div className="p-12 max-w-3xl mx-auto space-y-8 opacity-20 pointer-events-none select-none">
          <div className="h-12 w-48 bg-slate-300 dark:bg-slate-700 rounded-lg" />
          <div className="space-y-4">
            <div className="h-64 w-full bg-slate-200 dark:bg-slate-800 rounded-xl" />
            <div className="space-y-2">
              <div className="h-4 w-full bg-slate-300 dark:bg-slate-700 rounded" />
              <div className="h-4 w-5/6 bg-slate-300 dark:bg-slate-700 rounded" />
              <div className="h-4 w-4/6 bg-slate-300 dark:bg-slate-700 rounded" />
            </div>
          </div>
        </div>
        <WidgetApp
          key={`${id}:${refreshKey ?? 0}`}
          chatbotId={id || 'preview'}
          apiBase={import.meta.env.VITE_API_BASE_URL || ''}
          themeColor={themeColor}
          headerColor={chatHeaderColor}
          headerTextColor={chatHeaderTextColor}
          botMessageColor={botMessageColor}
          botMessageTextColor={botMessageTextColor}
          userMessageColor={userMessageColor}
          userMessageTextColor={userMessageTextColor}
          fontFamily={chatFontFamily}
          position={position as any}
          botNameOverride={botDisplayName}
          botIconOverride={botIcon}
          chatBg={chatBackgroundColor}
          welcome={welcomeMessage}
          autoOpen={previewOpen}
          useOverrides={true}
          resetSession={true}
          sessionIdOverride={sessionId}
          suggestions={suggestionsEnabled ? suggestedQuestions : []}
          hideBrandingOverride={hideBranding}
          customBrandingOverride={customBranding || undefined}
          positionStrategy="absolute"
        />
      </div>
    </div>
  )
}
