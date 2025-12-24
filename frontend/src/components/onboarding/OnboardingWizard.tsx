import { useState, useCallback, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Bot,
  Upload,
  Palette,
  Rocket,
  ArrowRight,
  ArrowLeft,
  CheckCircle2,
  Sparkles,
  FileText,
  Globe,
  Type,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import * as onboardingApi from '@/api/onboarding'

interface OnboardingStep {
  id: number
  title: string
  subtitle: string
  icon: React.ElementType
}

const steps: OnboardingStep[] = [
  {
    id: 1,
    title: 'Botunuzu Adlandırın',
    subtitle: 'Chatbotunuza benzersiz bir isim verin',
    icon: Bot,
  },
  {
    id: 2,
    title: 'Bilgi Kaynağı Ekleyin',
    subtitle: 'Botunuzun öğreneceği içeriği yükleyin',
    icon: Upload,
  },
  {
    id: 3,
    title: 'Kişiliğini Belirleyin',
    subtitle: 'Botunuzun nasıl konuşacağını ayarlayın',
    icon: Palette,
  },
  {
    id: 4,
    title: 'Hazır!',
    subtitle: 'Botunuz kullanıma hazır',
    icon: Rocket,
  },
]

type SourceType = 'text' | 'url' | 'file'

const MAX_FILE_SIZE_MB = 10

const OnboardingWizard = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [currentStep, setCurrentStep] = useState(1)
  const [isLoading, setIsLoading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Form States
  const [botName, setBotName] = useState('')
  const [sourceType, setSourceType] = useState<SourceType>('text')
  const [textContent, setTextContent] = useState('')
  const [urlContent, setUrlContent] = useState('')
  const [pdfFile, setPdfFile] = useState<File | null>(null)
  const [systemPrompt, setSystemPrompt] = useState(
    'Sen yardımsever ve samimi bir müşteri destek asistanısın. Kısa ve öz cevaplar ver.',
  )
  const [welcomeMessage, setWelcomeMessage] = useState('Merhaba! Size nasıl yardımcı olabilirim?')
  const [createdBotId, setCreatedBotId] = useState<string | null>(null)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.type !== 'application/pdf') {
      toast('Yalnızca PDF dosyaları desteklenir.', 'error')
      if (fileInputRef.current) fileInputRef.current.value = ''
      return
    }

    const maxSize = MAX_FILE_SIZE_MB * 1024 * 1024
    if (file.size > maxSize) {
      toast(`Dosya boyutu ${MAX_FILE_SIZE_MB}MB'den büyük olamaz.`, 'error')
      if (fileInputRef.current) fileInputRef.current.value = ''
      return
    }

    setPdfFile(file)
  }

  // Load saved onboarding state on mount
  useEffect(() => {
    const loadOnboardingState = async () => {
      try {
        const state = await onboardingApi.getOnboardingState()

        // If onboarding is already completed or skipped, redirect to dashboard
        if (state.completed || state.skipped) {
          navigate('/dashboard')
          return
        }

        // Restore state if user has started onboarding
        if (state.step > 0 && state.data) {
          setCurrentStep(state.step)
          if (state.data.bot_name) setBotName(state.data.bot_name)
          if (state.data.source_type) setSourceType(state.data.source_type as SourceType)
          if (state.data.text_content) setTextContent(state.data.text_content)
          if (state.data.url_content) setUrlContent(state.data.url_content)
          if (state.data.system_prompt) setSystemPrompt(state.data.system_prompt)
          if (state.data.welcome_message) setWelcomeMessage(state.data.welcome_message)
          if (state.data.created_bot_id) setCreatedBotId(state.data.created_bot_id)
        }
      } catch (error) {
        console.error('Failed to load onboarding state:', error)
      }
    }

    loadOnboardingState()
  }, [navigate])

  // Save state whenever it changes (debounced)
  useEffect(() => {
    const saveState = async () => {
      if (currentStep === 0 || currentStep === 4) return // Don't save on initial or final step

      const data: onboardingApi.OnboardingData = {
        bot_name: botName,
        source_type: sourceType,
        text_content: textContent,
        url_content: urlContent,
        system_prompt: systemPrompt,
        welcome_message: welcomeMessage,
        created_bot_id: createdBotId || undefined,
      }

      try {
        await onboardingApi.updateOnboardingState(currentStep, data)
      } catch (error) {
        console.error('Failed to save onboarding state:', error)
      }
    }

    const timer = setTimeout(saveState, 500) // Debounce state saving
    return () => clearTimeout(timer)
  }, [
    currentStep,
    botName,
    sourceType,
    textContent,
    urlContent,
    systemPrompt,
    welcomeMessage,
    createdBotId,
  ])

  const canProceed = useCallback(() => {
    switch (currentStep) {
      case 1:
        return botName.trim().length >= 2
      case 2:
        if (sourceType === 'text') return textContent.trim().length >= 50
        if (sourceType === 'url') return urlContent.trim().startsWith('http')
        if (sourceType === 'file') return pdfFile !== null
        return false
      case 3:
        return systemPrompt.trim().length >= 10
      case 4:
        return true
      default:
        return false
    }
  }, [currentStep, botName, sourceType, textContent, urlContent, pdfFile, systemPrompt])

  const handleNext = async () => {
    if (!canProceed()) {
      toast('Lütfen gerekli alanları doldurun.', 'error')
      return
    }

    // If on step 3, create the bot
    if (currentStep === 3) {
      setIsLoading(true)
      try {
        // Create the chatbot
        const { data: chatbot } = await api.post('/api/v1/chatbots', {
          name: botName,
          system_prompt: systemPrompt,
          welcome_message: welcomeMessage,
        })

        setCreatedBotId(chatbot.id)

        // Add source based on type using FormData (backend expects multipart/form-data)
        if (sourceType === 'text' && textContent.trim()) {
          const formData = new FormData()
          formData.append('source_type', 'text')
          formData.append('text', textContent)
          await api.post(`/api/v1/chatbots/${chatbot.id}/sources`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          })
        } else if (sourceType === 'url' && urlContent.trim()) {
          const formData = new FormData()
          formData.append('source_type', 'url')
          formData.append('source_url', urlContent)
          await api.post(`/api/v1/chatbots/${chatbot.id}/sources`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          })
        } else if (sourceType === 'file' && pdfFile) {
          const formData = new FormData()
          formData.append('source_type', 'pdf')
          formData.append('file', pdfFile)
          await api.post(`/api/v1/chatbots/${chatbot.id}/sources`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          })
        }

        // Mark onboarding as completed
        await onboardingApi.completeOnboarding(chatbot.id)

        toast('Botunuz başarıyla oluşturuldu!', 'success')
        setCurrentStep(4)
      } catch {
        toast('Bot oluşturulurken bir hata oluştu.', 'error')
      } finally {
        setIsLoading(false)
      }
      return
    }

    setCurrentStep((prev) => Math.min(prev + 1, 4))
  }

  const handleBack = () => {
    setCurrentStep((prev) => Math.max(prev - 1, 1))
  }

  const handleFinish = async () => {
    if (createdBotId) {
      navigate(`/chatbots/${createdBotId}`)
    } else {
      navigate('/dashboard')
    }
  }

  const handleSkip = async () => {
    try {
      await onboardingApi.skipOnboarding()
      navigate('/dashboard')
    } catch (error) {
      console.error('Failed to skip onboarding:', error)
      navigate('/dashboard') // Navigate anyway
    }
  }

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <div className="space-y-6 animate-fade-up">
            <div className="text-center mb-8">
              <div
                className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                            bg-primary/10 mb-4"
              >
                <Bot className="w-8 h-8 text-primary" />
              </div>
              <h2 className="heading-md text-foreground mb-2">Botunuza İsim Verin</h2>
              <p className="body-sm">Bu isim dashboard'da ve widget'ta görünecektir</p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground" htmlFor="botName">
                Bot Adı
              </label>
              <Input
                id="botName"
                placeholder="Örn: Müşteri Destek Botu"
                value={botName}
                onChange={(e) => setBotName(e.target.value)}
                className="h-12 rounded-xl border-border/50 bg-white/50 
                         focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
              />
              <p className="text-xs text-muted-foreground">Minimum 2 karakter</p>
            </div>
          </div>
        )

      case 2:
        return (
          <div className="space-y-6 animate-fade-up">
            <div className="text-center mb-8">
              <div
                className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                            bg-primary/10 mb-4"
              >
                <Upload className="w-8 h-8 text-primary" />
              </div>
              <h2 className="heading-md text-foreground mb-2">Bilgi Kaynağı Ekleyin</h2>
              <p className="body-sm">Botunuzun öğrenmesini istediğiniz içeriği seçin</p>
            </div>

            {/* Source Type Selector */}
            <div className="grid grid-cols-3 gap-3 mb-6">
              {[
                { type: 'text' as const, icon: Type, label: 'Metin' },
                { type: 'url' as const, icon: Globe, label: 'URL' },
                { type: 'file' as const, icon: FileText, label: 'PDF' },
              ].map(({ type, icon: Icon, label }) => (
                <button
                  key={type}
                  onClick={() => setSourceType(type)}
                  className={`p-4 rounded-xl border-2 transition-all duration-200 cursor-pointer
                    ${
                      sourceType === type
                        ? 'border-primary bg-primary/5 text-foreground'
                        : 'border-border/50 bg-white/50 text-muted-foreground hover:border-border'
                    }
                  `}
                >
                  <Icon className="w-6 h-6 mx-auto mb-2" />
                  <span className="text-sm font-medium">{label}</span>
                </button>
              ))}
            </div>

            {/* Content Input */}
            {sourceType === 'text' && (
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="textContent">
                  İçerik
                </label>
                <Textarea
                  id="textContent"
                  placeholder="Botunuzun öğrenmesini istediğiniz bilgileri buraya yapıştırın..."
                  value={textContent}
                  onChange={(e) => setTextContent(e.target.value)}
                  className="min-h-[160px] rounded-xl border-border/50 bg-white/50 
                           focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
                />
                <p className="text-xs text-muted-foreground">
                  Minimum 50 karakter ({textContent.length}/50)
                </p>
              </div>
            )}

            {sourceType === 'url' && (
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="urlContent">
                  Web Sitesi URL'si
                </label>
                <Input
                  id="urlContent"
                  type="url"
                  placeholder="https://example.com"
                  value={urlContent}
                  onChange={(e) => setUrlContent(e.target.value)}
                  className="h-12 rounded-xl border-border/50 bg-white/50 
                           focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
                />
                <p className="text-xs text-muted-foreground">
                  Site içeriği otomatik olarak analiz edilecek
                </p>
              </div>
            )}

            {sourceType === 'file' && (
              <div className="space-y-4">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,application/pdf"
                  onChange={handleFileSelect}
                  className="hidden"
                  id="pdf-upload"
                />

                {!pdfFile ? (
                  <label
                    htmlFor="pdf-upload"
                    className="flex flex-col items-center justify-center p-8 rounded-xl border-2 border-dashed 
                              border-border/50 bg-white/50 hover:border-primary/50 hover:bg-primary/5
                              cursor-pointer transition-all duration-200"
                  >
                    <Upload className="w-10 h-10 text-muted-foreground mb-3" />
                    <span className="text-sm font-medium text-foreground">PDF Dosyası Seçin</span>
                    <span className="text-xs text-muted-foreground mt-1">
                      Maksimum {MAX_FILE_SIZE_MB}MB
                    </span>
                  </label>
                ) : (
                  <div className="flex items-center gap-4 p-4 rounded-xl border border-border bg-white/80">
                    <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                      <FileText className="w-6 h-6 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-foreground truncate">{pdfFile.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {(pdfFile.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        setPdfFile(null)
                        if (fileInputRef.current) fileInputRef.current.value = ''
                      }}
                      className="text-muted-foreground hover:text-destructive"
                    >
                      Değiştir
                    </Button>
                  </div>
                )}

                <p className="text-xs text-muted-foreground">
                  PDF dosyası yükleyerek botunuzun bu içeriği öğrenmesini sağlayın
                </p>
              </div>
            )}
          </div>
        )

      case 3:
        return (
          <div className="space-y-6 animate-fade-up">
            <div className="text-center mb-8">
              <div
                className="inline-flex items-center justify-center w-16 h-16 rounded-2xl 
                            bg-primary/10 mb-4"
              >
                <Palette className="w-8 h-8 text-primary" />
              </div>
              <h2 className="heading-md text-foreground mb-2">Kişiliğini Belirleyin</h2>
              <p className="body-sm">Botunuzun nasıl davranacağını ve konuşacağını ayarlayın</p>
            </div>

            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="systemPrompt">
                  Sistem Talimatı
                </label>
                <Textarea
                  id="systemPrompt"
                  placeholder="Botunuzun nasıl davranacağını açıklayın..."
                  value={systemPrompt}
                  onChange={(e) => setSystemPrompt(e.target.value)}
                  className="min-h-[100px] rounded-xl border-border/50 bg-white/50 
                           focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="welcomeMessage">
                  Karşılama Mesajı
                </label>
                <Input
                  id="welcomeMessage"
                  placeholder="Merhaba! Size nasıl yardımcı olabilirim?"
                  value={welcomeMessage}
                  onChange={(e) => setWelcomeMessage(e.target.value)}
                  className="h-12 rounded-xl border-border/50 bg-white/50 
                           focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
                />
              </div>
            </div>
          </div>
        )

      case 4:
        return (
          <div className="text-center animate-fade-up">
            <div
              className="inline-flex items-center justify-center w-20 h-20 rounded-full 
                          bg-success/10 mb-6"
            >
              <CheckCircle2 className="w-10 h-10 text-success" />
            </div>
            <h2 className="heading-md text-foreground mb-3">Tebrikler! 🎉</h2>
            <p className="body-lg mb-8">
              <span className="font-semibold text-foreground">{botName}</span> başarıyla oluşturuldu
              ve kullanıma hazır.
            </p>

            <div className="glass-panel p-6 text-left mb-6">
              <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
                <Sparkles className="w-5 h-5 text-primary" />
                Şimdi Yapabilecekleriniz
              </h3>
              <ul className="space-y-3 text-sm text-muted-foreground">
                <li className="flex items-start gap-3">
                  <span className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 text-xs font-medium text-primary">
                    1
                  </span>
                  <span>Test Alanı'nda botunuzu test edin</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 text-xs font-medium text-primary">
                    2
                  </span>
                  <span>Daha fazla kaynak ekleyerek bilgi tabanını genişletin</span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 text-xs font-medium text-primary">
                    3
                  </span>
                  <span>Web sitenize embed kodu ile entegre edin</span>
                </li>
              </ul>
            </div>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className="min-h-screen bg-background relative overflow-hidden flex items-center justify-center p-6">
      {/* Animated Background */}
      <div className="absolute inset-0 gradient-mesh opacity-50" />
      <div className="absolute top-20 left-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl animate-float" />
      <div
        className="absolute bottom-20 right-20 w-96 h-96 bg-accent/30 rounded-full blur-3xl animate-float"
        style={{ animationDelay: '-3s' }}
      />

      <div className="relative z-10 w-full max-w-xl">
        {/* Progress Header */}
        <div className="mb-8">
          {/* Skip Button */}
          <div className="flex justify-end mb-4">
            <button
              onClick={handleSkip}
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              Atla
            </button>
          </div>

          {/* Progress Steps */}
          <div className="flex items-center justify-between mb-4">
            {steps.map((step, index) => (
              <div key={step.id} className="flex items-center">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300
                    ${
                      currentStep >= step.id
                        ? 'bg-primary text-white'
                        : 'bg-muted text-muted-foreground'
                    }
                    ${currentStep === step.id ? 'ring-4 ring-primary/20' : ''}
                  `}
                >
                  {currentStep > step.id ? (
                    <CheckCircle2 className="w-5 h-5" />
                  ) : (
                    <step.icon className="w-5 h-5" />
                  )}
                </div>
                {index < steps.length - 1 && (
                  <div
                    className={`w-16 sm:w-24 h-1 mx-2 rounded-full transition-colors duration-300
                    ${currentStep > step.id ? 'bg-primary' : 'bg-muted'}
                  `}
                  />
                )}
              </div>
            ))}
          </div>

          {/* Current Step Label */}
          <div className="text-center">
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              Adım {currentStep} / {steps.length}
            </p>
          </div>
        </div>

        {/* Form Card */}
        <div className="glass-card p-8 lg:p-10">
          {renderStepContent()}

          {/* Navigation Buttons */}
          <div className="flex items-center justify-between mt-8 pt-6 border-t border-border/50">
            {currentStep > 1 && currentStep < 4 ? (
              <Button variant="ghost" onClick={handleBack} className="gap-2">
                <ArrowLeft className="w-4 h-4" />
                Geri
              </Button>
            ) : (
              <div />
            )}

            {currentStep < 4 ? (
              <Button
                onClick={handleNext}
                isLoading={isLoading}
                disabled={!canProceed()}
                className="gap-2 bg-primary hover:bg-primary/90 shadow-lg shadow-primary/25"
              >
                {currentStep === 3 ? 'Botu Oluştur' : 'İleri'}
                {!isLoading && <ArrowRight className="w-4 h-4" />}
              </Button>
            ) : (
              <Button
                onClick={handleFinish}
                className="w-full gap-2 bg-primary hover:bg-primary/90 shadow-lg shadow-primary/25"
              >
                Botu Görüntüle
                <ArrowRight className="w-4 h-4" />
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default OnboardingWizard
