import { useEffect, useRef, useState } from 'react'
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
  const containerRef = useRef<HTMLDivElement>(null)
  const [dynamicPanelHeight, setDynamicPanelHeight] = useState('600px')

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const updateHeight = () => {
      const containerHeight = container.clientHeight
      // Panel yüksekliği: container yüksekliği - bubble alanı (80px) - margin (32px)
      // Minimum 400px, maximum containerHeight - 100px
      const calculatedHeight = Math.max(400, Math.min(containerHeight - 100, 700))
      setDynamicPanelHeight(`${calculatedHeight}px`)
    }

    // İlk yükleme
    updateHeight()

    // ResizeObserver ile container boyutunu dinle
    const resizeObserver = new ResizeObserver(updateHeight)
    resizeObserver.observe(container)

    return () => resizeObserver.disconnect()
  }, [])

  return (
    <div ref={containerRef} className="flex-1 relative h-full min-h-[400px]">
      <style>{styles}</style>
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
        panelHeight={dynamicPanelHeight}
      />
    </div>
  )
}
