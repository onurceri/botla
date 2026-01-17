import { useState, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useChatbotContext } from '../../context/ChatbotContext'
import { usePreview } from '../../hooks/usePreview'
import IdentitySection from '../../components/IdentitySection'
import AppearanceSection from '../../components/AppearanceSection'
import ColorsSection from '../../components/ColorsSection'
import BrandingSettings from '../../components/BrandingSettings'
import SuggestionsPanel from '../../components/SuggestionsPanel'
import PlaygroundPreview from '../../components/PlaygroundPreview'
import { Settings2, MessageSquare, RefreshCw, ChevronDown, Play } from 'lucide-react'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import {
  useUpdateAppearance,
  useRegenerateSuggestions,
} from '@/hooks/mutations/useChatbotMutations'
import { useSuggestionPolling } from '@/features/chatbot/hooks/useSuggestionPolling'
import { Button } from '@/components/ui/button'

export default function PlaygroundTab() {
  const { id = '' } = useParams()
  const { previewOpen, sessionId } = usePreview()
  const [expandedSection, setExpandedSection] = useState<string | null>('identity')
  const [previewRefreshKey] = useState(0)

  const {
    botDisplayName,
    setBotDisplayName,
    botIcon,
    setBotIcon,
    welcomeMessage,
    setWelcomeMessage,
    position,
    setPosition,
    chatFontFamily,
    setChatFontFamily,
    themeColor,
    setThemeColor,
    chatBackgroundColor,
    setChatBackgroundColor,
    chatHeaderColor,
    setChatHeaderColor,
    chatHeaderTextColor,
    setChatHeaderTextColor,
    botMessageColor,
    setBotMessageColor,
    botMessageTextColor,
    setBotMessageTextColor,
    userMessageColor,
    setUserMessageColor,
    userMessageTextColor,
    setUserMessageTextColor,
    inputBackgroundColor,
    setInputBackgroundColor,
    inputTextColor,
    setInputTextColor,
    sendButtonColor,
    setSendButtonColor,
    bubbleRadius,
    setBubbleRadius,
    hideBranding,
    setHideBranding,
    customBranding,
    setCustomBranding,
    planConfig,
    suggestionsEnabled,
    setSuggestionsEnabled,
    suggestedQuestions,
    manualQuestions,
    setManualQuestions,
    refetchChatbot,
  } = useChatbotContext()

  // Poll for suggestions if processing is active
  useSuggestionPolling(id, refetchChatbot)

  const appearancePayload = useMemo(
    () => ({
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
      input_background_color: inputBackgroundColor,
      input_text_color: inputTextColor,
      send_button_color: sendButtonColor,
      bubble_radius: bubbleRadius,
      hide_branding: hideBranding,
      custom_branding: hideBranding ? customBranding : null,
      suggestions_enabled: suggestionsEnabled,
      manual_questions: manualQuestions,
    }),
    [
      botDisplayName,
      botIcon,
      welcomeMessage,
      position,
      chatFontFamily,
      themeColor,
      chatBackgroundColor,
      chatHeaderColor,
      chatHeaderTextColor,
      botMessageColor,
      botMessageTextColor,
      userMessageColor,
      userMessageTextColor,
      inputBackgroundColor,
      inputTextColor,
      sendButtonColor,
      bubbleRadius,
      hideBranding,
      customBranding,
      suggestionsEnabled,
      manualQuestions,
    ],
  )

  const { mutateAsync: updateAppearance } = useUpdateAppearance(id)
  const regenerateSuggestions = useRegenerateSuggestions(id)

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: appearancePayload,
    saveFn: (_, payload) => updateAppearance(payload),
  })

  return (
    <div className="flex flex-col animate-in fade-in duration-500 pb-4 lg:pb-0 space-y-4">
      {/* Header */}
      <div className="flex flex-col gap-2 px-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-slate-900/5 text-slate-900">
              <Play className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-xl font-bold tracking-tight text-slate-900">Görünüm ve Test</h2>
              <p className="text-sm text-slate-500 font-medium">
                Chatbotunuzun görünümünü özelleştirin ve canlı olarak test edin
              </p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
          </div>
        </div>
      </div>

      {/* Main Content - Two columns side by side */}
      <div className="flex flex-col lg:flex-row gap-4 xl:gap-6 lg:items-start">
        {/* Left: Settings Sidebar */}
        <div className="w-full lg:w-[380px] xl:w-[420px] shrink-0 flex flex-col bg-slate-50/40 backdrop-blur-xl rounded-[24px] lg:rounded-[32px] border border-slate-200/60 shadow-[0_8px_32px_rgba(0,0,0,0.04)] overflow-hidden">
          <div className="p-4 lg:p-5 border-b border-slate-200/60 bg-white/60 backdrop-blur-md flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className="p-1.5 rounded-lg bg-primary/10 text-primary">
                <Settings2 className="w-4 h-4" />
              </div>
              <span className="text-xs font-bold uppercase tracking-[0.1em] text-slate-500">
                Özelleştirme
              </span>
            </div>
            <div className="sm:hidden">
              <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
            </div>
          </div>

          <div className="overflow-y-auto p-4 lg:p-5 space-y-3 lg:space-y-4 scrollbar-thin scrollbar-thumb-slate-200 scrollbar-track-transparent">
            <IdentitySection
              isExpanded={expandedSection === 'identity'}
              onToggle={() =>
                setExpandedSection(expandedSection === 'identity' ? null : 'identity')
              }
              botDisplayName={botDisplayName}
              setBotDisplayName={setBotDisplayName}
              botIcon={botIcon}
              setBotIcon={setBotIcon}
              welcomeMessage={welcomeMessage}
              setWelcomeMessage={setWelcomeMessage}
            />

            <AppearanceSection
              isExpanded={expandedSection === 'appearance'}
              onToggle={() =>
                setExpandedSection(expandedSection === 'appearance' ? null : 'appearance')
              }
              position={position}
              setPosition={setPosition}
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
              inputBackgroundColor={inputBackgroundColor}
              setInputBackgroundColor={setInputBackgroundColor}
              inputTextColor={inputTextColor}
              setInputTextColor={setInputTextColor}
              sendButtonColor={sendButtonColor}
              setSendButtonColor={setSendButtonColor}
              chatFontFamily={chatFontFamily}
              setChatFontFamily={setChatFontFamily}
              themeColor={themeColor}
              setThemeColor={setThemeColor}
              bubbleRadius={bubbleRadius}
              setBubbleRadius={setBubbleRadius}
            />

            {/* Suggestions Section */}
            <div
              className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${expandedSection === 'suggestions' ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}
            >
              <button
                onClick={() =>
                  setExpandedSection(expandedSection === 'suggestions' ? null : 'suggestions')
                }
                className="w-full flex items-center justify-between p-4 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div
                    className={`p-2 rounded-xl transition-colors ${expandedSection === 'suggestions' ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}
                  >
                    <MessageSquare className="w-4 h-4" />
                  </div>
                  <span
                    className={`text-[13px] font-bold tracking-tight ${expandedSection === 'suggestions' ? 'text-slate-900' : 'text-slate-600'}`}
                  >
                    Örnek Sorular
                  </span>
                </div>
                <div
                  className={`transition-transform duration-300 ${expandedSection === 'suggestions' ? 'rotate-180' : ''}`}
                >
                  <ChevronDown className="w-4 h-4 text-slate-300" />
                </div>
              </button>

              {expandedSection === 'suggestions' && (
                <div className="px-4 pb-5 space-y-4 animate-in fade-in slide-in-from-top-2 duration-300">
                  <div className="h-px bg-slate-100 -mx-4 mb-4" />

                  <SuggestionsPanel
                    suggestionsEnabled={suggestionsEnabled}
                    setSuggestionsEnabled={setSuggestionsEnabled}
                    suggestedQuestions={suggestedQuestions}
                    manualQuestions={manualQuestions}
                    setManualQuestions={setManualQuestions}
                    maxManualQuestions={planConfig.chat?.max_manual_questions ?? 3}
                  />

                  {suggestionsEnabled && (
                    <div className="pt-2 border-t border-slate-100">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => regenerateSuggestions.mutate()}
                        disabled={regenerateSuggestions.isPending}
                        className="w-full h-9 text-[11px] font-bold text-primary hover:text-primary hover:bg-primary/5 gap-2 rounded-xl transition-all border border-primary/10 hover:border-primary/20"
                      >
                        <RefreshCw
                          className={`w-3.5 h-3.5 ${regenerateSuggestions.isPending ? 'animate-spin' : ''}`}
                        />
                        {regenerateSuggestions.isPending
                          ? 'Soruları Yeniden Üret'
                          : 'AI ile Soruları Yenile'}
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </div>

            <BrandingSettings
              isExpanded={expandedSection === 'branding'}
              onToggle={() =>
                setExpandedSection(expandedSection === 'branding' ? null : 'branding')
              }
              hideBranding={hideBranding}
              setHideBranding={setHideBranding}
              customBranding={customBranding}
              setCustomBranding={setCustomBranding}
              canHideBranding={planConfig.branding?.can_hide_branding ?? false}
              canCustomBranding={planConfig.branding?.can_custom_branding ?? false}
            />
          </div>
        </div>

        {/* Right: Widget Preview - Fixed height, sticky on desktop */}
        <div className="w-full lg:w-[420px] xl:w-[450px] shrink-0 h-[500px] lg:h-[600px] lg:sticky lg:top-4 rounded-[24px] lg:rounded-[32px] border border-slate-200/60 shadow-[0_8px_32px_rgba(0,0,0,0.04)] overflow-hidden bg-white">
          <PlaygroundPreview
            id={id || 'preview'}
            themeColor={themeColor}
            chatHeaderColor={chatHeaderColor}
            chatHeaderTextColor={chatHeaderTextColor}
            botMessageColor={botMessageColor}
            botMessageTextColor={botMessageTextColor}
            userMessageColor={userMessageColor}
            userMessageTextColor={userMessageTextColor}
            inputBackgroundColor={inputBackgroundColor}
            inputTextColor={inputTextColor}
            sendButtonColor={sendButtonColor}
            chatFontFamily={chatFontFamily}
            position={position}
            botDisplayName={botDisplayName}
            botIcon={botIcon}
            chatBackgroundColor={chatBackgroundColor}
            bubbleRadius={bubbleRadius}
            welcomeMessage={welcomeMessage}
            previewOpen={previewOpen}
            sessionId={sessionId}
            suggestionsEnabled={suggestionsEnabled}
            suggestedQuestions={[...manualQuestions, ...suggestedQuestions]}
            refreshKey={previewRefreshKey}
            hideBranding={hideBranding}
            customBranding={customBranding}
            autoOpen={true}
            panelHeight="100%"
            panelWidth="100%"
          />
        </div>
      </div>
    </div>
  )
}
