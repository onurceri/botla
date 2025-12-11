import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useChatbotContext } from '../../context/ChatbotContext'
import { usePreview } from '../../hooks/usePreview'
import IdentitySection from '../../components/IdentitySection'
import AppearanceSection from '../../components/AppearanceSection'
import ColorsSection from '../../components/ColorsSection'
import BrandingSettings from '../../components/BrandingSettings'
import PlaygroundPreview from '../../components/PlaygroundPreview'
import { Smartphone, Palette } from 'lucide-react'

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
    suggestedQuestions
  } = useChatbotContext()

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">Görünüm ve Test</h2>
        <p className="text-muted-foreground">
          Chatbotunuzun görünümünü özelleştirin ve anlık olarak test edin.
        </p>
      </div>

      <div className="grid lg:grid-cols-12 gap-8 h-[calc(100vh-220px)] min-h-[600px]">
        {/* Settings Sidebar */}
        <div className="lg:col-span-4 overflow-y-auto pr-2 space-y-6">
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
        <div className="lg:col-span-8 bg-muted/30 rounded-3xl border border-border px-8 pb-8 pt-16 flex flex-col relative overflow-hidden shadow-inner">
            <div className="absolute top-6 left-6 flex items-center gap-2 text-muted-foreground bg-background/80 backdrop-blur px-4 py-1.5 rounded-full text-xs font-medium border border-border/50 shadow-sm">
                <Smartphone className="w-3 h-3" />
                Canlı Önizleme
            </div>
            
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
  )
}
