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
import PlaygroundConsole from '../../components/PlaygroundConsole'
import { Smartphone, Palette, Monitor, Laptop, Zap, Settings2, MessageSquare, RefreshCw, ChevronDown } from 'lucide-react'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'
import { useUpdateAppearance, useRegenerateSuggestions } from '@/hooks/mutations/useChatbotMutations'
import { Button } from '@/components/ui/button'

export default function PlaygroundTab() {
  const { id = '' } = useParams()
  const { previewOpen, sessionId } = usePreview()
  const [expandedSection, setExpandedSection] = useState<string | null>('identity')
  const [previewRefreshKey] = useState(0)
  const [previewDevice, setPreviewDevice] = useState<'desktop' | 'tablet' | 'mobile'>('desktop')

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
    inputBackgroundColor, setInputBackgroundColor,
    inputTextColor, setInputTextColor,
    sendButtonColor, setSendButtonColor,
    bubbleRadius, setBubbleRadius,
    hideBranding, setHideBranding,
    customBranding, setCustomBranding,
    planConfig,
    suggestionsEnabled, setSuggestionsEnabled,
    suggestedQuestions, setSuggestedQuestions,
    allSuggestedQuestions,
  } = useChatbotContext()

  // Memoize the payload to prevent unnecessary API calls on every keystroke
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
    input_background_color: inputBackgroundColor,
    input_text_color: inputTextColor,
    send_button_color: sendButtonColor,
    bubble_radius: bubbleRadius,
    hide_branding: hideBranding,
    custom_branding: hideBranding ? customBranding : null,
  }), [
    botDisplayName, botIcon, welcomeMessage, position, chatFontFamily,
    themeColor, chatBackgroundColor, chatHeaderColor, chatHeaderTextColor,
    botMessageColor, botMessageTextColor, userMessageColor, userMessageTextColor,
    inputBackgroundColor, inputTextColor, sendButtonColor, bubbleRadius,
    hideBranding, customBranding,
  ])

  const { mutateAsync: updateAppearance } = useUpdateAppearance(id)
  const regenerateSuggestions = useRegenerateSuggestions(id)

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: appearancePayload,
    saveFn: (_, payload) => updateAppearance(payload),
  })

  return (
    <div className="flex flex-col h-[calc(100vh-140px)] -mt-2 animate-in fade-in duration-500">
      <div className="flex items-center justify-between mb-4 shrink-0 px-1">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-primary/10 text-primary shadow-inner">
            <Zap className="w-5 h-5 fill-primary/20" />
          </div>
          <div>
            <h2 className="text-xl font-bold tracking-tight text-slate-900">Playground</h2>
            <p className="text-xs text-muted-foreground font-medium">Test & Geliştirici Ortamı</p>
          </div>
        </div>
        <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
      </div>

      <div className="flex-1 flex gap-6 min-h-0">
        {/* Settings Sidebar (Property Inspector Style) */}
        <div className="w-[380px] shrink-0 flex flex-col bg-slate-50/40 backdrop-blur-xl rounded-[32px] border border-slate-200/60 shadow-[0_8px_32px_rgba(0,0,0,0.04)] overflow-hidden">
             <div className="p-5 border-b border-slate-200/60 bg-white/60 backdrop-blur-md flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                    <div className="p-1.5 rounded-lg bg-primary/10 text-primary">
                        <Settings2 className="w-4 h-4" />
                    </div>
                    <span className="text-xs font-bold uppercase tracking-[0.1em] text-slate-500">Özelleştirme</span>
                </div>
             </div>
             
             <div className="flex-1 overflow-y-auto p-5 space-y-4 scrollbar-thin scrollbar-thumb-slate-200 scrollbar-track-transparent">
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

                {/* Suggestions Section - Integrated into Sidebar */}
                <div className={`transition-all duration-300 border border-slate-200/60 rounded-2xl overflow-hidden ${expandedSection === 'suggestions' ? 'bg-white shadow-[0_4px_20px_rgba(0,0,0,0.03)]' : 'bg-white/40 hover:bg-white/60'}`}>
                  <button 
                    onClick={() => setExpandedSection(expandedSection === 'suggestions' ? null : 'suggestions')}
                    className="w-full flex items-center justify-between p-4 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-xl transition-colors ${expandedSection === 'suggestions' ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}>
                        <MessageSquare className="w-4 h-4" />
                      </div>
                      <span className={`text-[13px] font-bold tracking-tight ${expandedSection === 'suggestions' ? 'text-slate-900' : 'text-slate-600'}`}>Örnek Sorular</span>
                    </div>
                    <div className={`transition-transform duration-300 ${expandedSection === 'suggestions' ? 'rotate-180' : ''}`}>
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
                        setSuggestedQuestions={setSuggestedQuestions}
                        allSuggestedQuestions={allSuggestedQuestions}
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
                            <RefreshCw className={`w-3.5 h-3.5 ${regenerateSuggestions.isPending ? 'animate-spin' : ''}`} />
                            {regenerateSuggestions.isPending ? 'Soruları Yeniden Üret' : 'AI ile Soruları Yenile'}
                          </Button>
                        </div>
                      )}
                    </div>
                  )}
                </div>

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
        <div className="flex-1 flex flex-col bg-white rounded-[32px] border border-slate-200/60 overflow-hidden shadow-[0_8px_32px_rgba(0,0,0,0.02)] relative">
            {/* Toolbar */}
            <div className="h-14 bg-white/80 backdrop-blur-md border-b border-slate-100 px-6 flex items-center justify-between shrink-0 z-20">
                <div className="flex items-center gap-6">
                    <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-red-400/20 border border-red-400/30" />
                        <div className="w-2.5 h-2.5 rounded-full bg-amber-400/20 border border-amber-400/30" />
                        <div className="w-2.5 h-2.5 rounded-full bg-emerald-400/20 border border-emerald-400/30" />
                    </div>
                    <div className="h-4 w-px bg-slate-200" />
                    <div className="flex bg-slate-100/80 p-1 rounded-xl border border-slate-200/50">
                        <button 
                            onClick={() => setPreviewDevice('desktop')}
                            className={`p-1.5 rounded-lg transition-all duration-300 ${previewDevice === 'desktop' ? 'bg-white shadow-sm text-primary scale-105' : 'text-slate-400 hover:text-slate-600 hover:bg-white/50'}`}
                            title="Masaüstü"
                        >
                            <Monitor className="w-4 h-4" />
                        </button>
                        <button 
                            onClick={() => setPreviewDevice('tablet')}
                            className={`p-1.5 rounded-lg transition-all duration-300 ${previewDevice === 'tablet' ? 'bg-white shadow-sm text-primary scale-105' : 'text-slate-400 hover:text-slate-600 hover:bg-white/50'}`}
                            title="Tablet"
                        >
                            <Laptop className="w-4 h-4" />
                        </button>
                        <button 
                            onClick={() => setPreviewDevice('mobile')}
                            className={`p-1.5 rounded-lg transition-all duration-300 ${previewDevice === 'mobile' ? 'bg-white shadow-sm text-primary scale-105' : 'text-slate-400 hover:text-slate-600 hover:bg-white/50'}`}
                            title="Mobil"
                        >
                            <Smartphone className="w-4 h-4" />
                        </button>
                    </div>
                </div>
                <div className="flex-1 max-w-sm mx-8">
                    <div className="bg-slate-50/80 px-4 py-1.5 rounded-xl text-[11px] text-slate-500 font-medium truncate text-center border border-slate-100/80 shadow-inner flex items-center justify-center gap-2 group cursor-default">
                        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]" />
                        https://botla.app/preview/{id}
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
                </div>
            </div>

            <div className="flex-1 relative bg-slate-50/50 overflow-hidden flex items-center justify-center p-12 min-h-0">
                {/* Background Pattern */}
                <div className="absolute inset-0 opacity-[0.03] pointer-events-none" style={{ backgroundImage: 'radial-gradient(#000 1px, transparent 1px)', backgroundSize: '32px 32px' }} />

                {/* Device Container */}
                <div className={`relative h-full transition-all duration-700 cubic-bezier(0.4, 0, 0.2, 1) bg-white shadow-[0_32px_64px_-12px_rgba(0,0,0,0.1)] border border-slate-200 overflow-hidden ${
                    previewDevice === 'desktop' ? 'w-full rounded-2xl' : 
                    previewDevice === 'tablet' ? 'w-[768px] rounded-[40px] border-4 border-slate-800' : 
                    'w-[375px] rounded-[56px] border-[12px] border-slate-900 shadow-2xl'
                }`}>
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
                          suggestedQuestions={suggestedQuestions}
                          refreshKey={previewRefreshKey}
                          hideBranding={hideBranding}
                          customBranding={customBranding}
                        />
                    </div>
                </div>

                <PlaygroundConsole />
            </div>
        </div>
      </div>
    </div>
  )
}
