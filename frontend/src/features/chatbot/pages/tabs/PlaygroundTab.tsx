import { useState, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useChatbotContext } from '../../context/ChatbotContext'
import { usePreview } from '../../hooks/usePreview'
import IdentitySection from '../../components/IdentitySection'
import AppearanceSection from '../../components/AppearanceSection'
import ColorsSection from '../../components/ColorsSection'
import BrandingSettings from '../../components/BrandingSettings'
import PlaygroundPreview from '../../components/PlaygroundPreview'
import { Smartphone, Palette } from 'lucide-react'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateAppearance } from '@/hooks/mutations/useChatbotMutations'

export default function PlaygroundTab() {
  const { id = '' } = useParams()
  const { previewOpen, sessionId } = usePreview()
  const [expandedSection, setExpandedSection] = useState<string | null>('identity')
  const [previewRefreshKey] = useState(0)

  const {
    botDisplayName, setBotDisplayName,
    botIcon, setBotIcon,
    welcomeMessage, setWelcomeMessage,
    position, setPosition,
    chatFontFamily, setChatFontFamily,
    themeColor, setThemeColor,
    chatBackgroundColor, setChatBackgroundColor,
    chatHeaderColor, setChatHeaderColor,
    chatHeaderTextColor, setChatHeaderTextColor,
    botMessageColor, setBotMessageColor,
    botMessageTextColor, setBotMessageTextColor,
    userMessageColor, setUserMessageColor,
    userMessageTextColor, setUserMessageTextColor,
    hideBranding, setHideBranding,
    customBranding, setCustomBranding,
    planConfig,
    suggestionsEnabled,
    suggestedQuestions,
  } = useChatbotContext()

  // Memoize the payload to prevent unnecessary API calls on every keystroke
  // The debounce in useAutoSave still works, but this prevents extra re-renders
  const appearancePayload = useMemo(() => ({
    bot_display_name: botDisplayName,
    bot_icon: botIcon,
    welcome_message: welcomeMessage,
    position,
    chat_font_family: chatFontFamily,
    theme_color: themeColor,
    chat_background_color: chatBackgroundColor,
    chat_header_color: chatHeaderColor,
    chat_header_text_color: chatHeaderTextColor,
    bot_message_color: botMessageColor,
    bot_message_text_color: botMessageTextColor,
    user_message_color: userMessageColor,
    user_message_text_color: userMessageTextColor,
    hide_branding: hideBranding,
    custom_branding: hideBranding ? customBranding : null,
  }), [
    botDisplayName, botIcon, welcomeMessage, position, chatFontFamily,
    themeColor, chatBackgroundColor, chatHeaderColor, chatHeaderTextColor,
    botMessageColor, botMessageTextColor, userMessageColor, userMessageTextColor,
    hideBranding, customBranding,
  ])

  const { mutateAsync: updateAppearance } = useUpdateAppearance(id)

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: appearancePayload,
    saveFn: (id, payload) => updateAppearance(payload),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Görünüm ve Test</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Chatbotunuzun görünümünü özelleştirin ve anlık olarak test edin.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 h-[calc(100vh-250px)] min-h-[500px] lg:h-[calc(100vh-280px)] lg:min-h-[550px]">
        {/* Settings Sidebar */}
        <div className="lg:col-span-4 lg:overflow-y-auto pr-2 space-y-6">
             <div className="flex items-center gap-2 pb-2 border-b border-border">
                <Palette className="w-5 h-5 text-primary" />
                <h3 className="font-semibold text-foreground">Özelleştirme</h3>
             </div>
             
             <div className="space-y-4">
                <IdentitySection 
                  isExpanded={expandedSection === 'identity'}
                  onToggle={() => setExpandedSection(expandedSection === 'identity' ? null : 'identity')}
                  botDisplayName={botDisplayName}
                  setBotDisplayName={setBotDisplayName}
                  botIcon={botIcon}
                  setBotIcon={setBotIcon}
                  welcomeMessage={welcomeMessage}
                  setWelcomeMessage={setWelcomeMessage}
                />

                <AppearanceSection 
                  isExpanded={expandedSection === 'appearance'}
                  onToggle={() => setExpandedSection(expandedSection === 'appearance' ? null : 'appearance')}
                  position={position}
                  setPosition={setPosition}
                  chatFontFamily={chatFontFamily}
                  setChatFontFamily={setChatFontFamily}
                  themeColor={themeColor}
                  setThemeColor={setThemeColor}
                />

                <ColorsSection 
                  isExpanded={expandedSection === 'colors'}
                  onToggle={() => setExpandedSection(expandedSection === 'colors' ? null : 'colors')}
                  chatBackgroundColor={chatBackgroundColor}
                  setChatBackgroundColor={setChatBackgroundColor}
                  chatHeaderColor={chatHeaderColor}
                  setChatHeaderColor={setChatHeaderColor}
                  chatHeaderTextColor={chatHeaderTextColor}
                  setChatHeaderTextColor={setChatHeaderTextColor}
                  botMessageColor={botMessageColor}
                  setBotMessageColor={setBotMessageColor}
                  botMessageTextColor={botMessageTextColor}
                  setBotMessageTextColor={setBotMessageTextColor}
                  userMessageColor={userMessageColor}
                  setUserMessageColor={setUserMessageColor}
                  userMessageTextColor={userMessageTextColor}
                  setUserMessageTextColor={setUserMessageTextColor}
                />

                <BrandingSettings
                  isExpanded={expandedSection === 'branding'}
                  onToggle={() => setExpandedSection(expandedSection === 'branding' ? null : 'branding')}
                  hideBranding={hideBranding}
                  setHideBranding={setHideBranding}
                  customBranding={customBranding}
                  setCustomBranding={setCustomBranding}
                  canHideBranding={planConfig.branding?.can_hide_branding ?? false}
                  canCustomBranding={planConfig.branding?.can_custom_branding ?? false}
                />
             </div>
        </div>

        {/* Preview Area */}
        <div className="lg:col-span-8 relative h-full min-h-[450px] md:min-h-[500px] bg-slate-50/40 rounded-3xl border-2 border-dashed border-slate-200 overflow-hidden">
            {/* Background Pattern */}
            <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'radial-gradient(#000 1px, transparent 1px)', backgroundSize: '24px 24px' }} />

            {/* Instructional Overlay */}
            <div className="absolute inset-0 flex flex-col items-center justify-center text-center p-6 text-muted-foreground select-none pointer-events-none">
                <div className="bg-white/80 backdrop-blur-sm p-4 rounded-full shadow-sm mb-4 ring-1 ring-black/5 animate-in fade-in zoom-in duration-500">
                    <Smartphone className="w-8 h-8 text-primary/80" />
                </div>
                <h3 className="text-xl font-semibold text-foreground/80 mb-2">Canlı Önizleme Ortamı</h3>
                <p className="max-w-[420px] mb-6 text-balance leading-relaxed">
                    Chatbotunuz web sitenizde ziyaretçilerinize görüneceği şekilde burada aktiftir. 
                    Şu an <strong>{position === 'bottom-left' ? 'Sol Alt' : 'Sağ Alt'}</strong> köşede konumlanmıştır.
                </p>
                <div className="flex items-center gap-2 text-sm bg-white/60 backdrop-blur-sm px-4 py-2 rounded-full border border-border/50 shadow-sm text-foreground/70">
                    <Palette className="w-4 h-4" />
                    <span>Görünümü soldaki panelden özelleştirebilirsiniz</span>
                </div>
            </div>

            {/* Live Chatbot Layer */}
            <div className="absolute inset-0 z-10">
                <PlaygroundPreview
                  id={id || 'preview'}
                  themeColor={themeColor}
                  chatHeaderColor={chatHeaderColor}
                  chatHeaderTextColor={chatHeaderTextColor}
                  botMessageColor={botMessageColor}
                  botMessageTextColor={botMessageTextColor}
                  userMessageColor={userMessageColor}
                  userMessageTextColor={userMessageTextColor}
                  chatFontFamily={chatFontFamily}
                  position={position}
                  botDisplayName={botDisplayName}
                  botIcon={botIcon}
                  chatBackgroundColor={chatBackgroundColor}
                  welcomeMessage={welcomeMessage}
                  previewOpen={previewOpen}
                  sessionId={sessionId}
                  suggestionsEnabled={suggestionsEnabled}
                  suggestedQuestions={suggestedQuestions}
                  refreshKey={previewRefreshKey}
                  hideBranding={hideBranding}
                  customBranding={customBranding}
                />
            </div>
        </div>
      </div>
    </div>
  )
}
