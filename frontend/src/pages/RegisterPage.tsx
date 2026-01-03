import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Mail, Lock, User, ArrowRight, Zap, Shield, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import { getTurkishErrorMessage } from '@/lib/errorMessages'

const RegisterPage = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !email || !password) {
      toast('Lütfen tüm alanları doldurun.', 'error')
      return
    }

    setErrorMsg(null)
    setIsLoading(true)
    try {
      // Register the user
      await api.post('/api/v1/auth/register', { full_name: name, email, password })

      // Auto-login after registration
      const { data } = await api.post('/api/v1/auth/login', { email, password })
      window.localStorage.setItem('botla_token', data.token)
      window.localStorage.setItem('botla_refresh_token', data.refresh_token)
      window.localStorage.removeItem('botla_last_org_id')

      // Fetch onboarding status to determine where to redirect
      const { data: onboardingState } = await api.get('/api/v1/me/onboarding')

      toast('Hesabınız oluşturuldu! Hadi başlayalım.', 'success')

      // Redirect based on onboarding status
      if (onboardingState.completed || onboardingState.skipped) {
        navigate('/dashboard')
      } else {
        navigate('/onboarding')
      }
    } catch (err: any) {
      const errorMessage = getTurkishErrorMessage(err, 'Kayıt başarısız. Lütfen tekrar deneyin.')
      toast(errorMessage, 'error')
      setErrorMsg(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  const features = [
    {
      icon: Zap,
      title: 'Dakikalar İçinde Kurulum',
      description: 'Bot oluşturma sürecinizi sadece birkaç dakikada tamamlayın.',
    },
    {
      icon: Shield,
      title: 'Güvenli & Gizli',
      description: 'Verileriniz şifrelenir ve güvenle saklanır.',
    },
    {
      icon: Clock,
      title: '7/24 Aktif',
      description: 'Botunuz asla uyumaz, her zaman müşterilerinize hizmet eder.',
    },
  ]

  return (
    <div className="min-h-screen flex bg-background relative overflow-hidden">
      {/* Animated Background Mesh */}
      <div className="absolute inset-0 gradient-mesh opacity-60" />

      {/* Floating Decorative Elements */}
      <div className="absolute top-40 right-20 w-80 h-80 bg-primary/10 rounded-full blur-3xl animate-float" />
      <div
        className="absolute bottom-10 left-10 w-96 h-96 bg-accent/30 rounded-full blur-3xl animate-float"
        style={{ animationDelay: '-2s' }}
      />
      <div className="absolute top-1/3 right-1/4 w-52 h-52 bg-primary/5 blob-animated blur-2xl" />

      {/* Left Side - Registration Form */}
      <div className="flex-1 flex items-center justify-center p-6 lg:p-12 relative z-10">
        <div className="w-full max-w-md">
          {/* Mobile Logo */}
          <div className="lg:hidden flex items-center justify-center gap-3 mb-12 animate-fade-up">
            <img src="/logo-128.png" alt="Botla Logo" className="w-10 h-10 rounded-xl shadow-lg" />
            <span className="text-xl font-bold text-foreground">botla.app</span>
          </div>

          {/* Form Card */}
          <div
            className="glass-card p-8 lg:p-10 animate-scale-in"
            data-testid="register-page"
          >
            <div className="text-center mb-8">
              <h2
                className="heading-md text-foreground mb-2"
                data-testid="register-page-title"
              >
                Hesap Oluştur
              </h2>
              <p className="body-sm">Hemen başlayın, ücretsiz deneyin</p>
            </div>

            {/* Error Message */}
            {errorMsg && (
              <div
                className="mb-6 p-4 rounded-xl bg-destructive/10 border border-destructive/20
                         text-destructive text-sm animate-fade-up"
                role="alert"
                data-testid="register-page-error-message"
              >
                {errorMsg}
              </div>
            )}

            <form onSubmit={handleRegister} className="space-y-5 stagger-children">
              {/* Name Field */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="name">
                  Ad Soyad
                </label>
                <div className="relative">
                  <User className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input
                    id="name"
                    placeholder="Adınız Soyadınız"
                    data-testid="register-page-name-input"
                    className="pl-12 h-12 rounded-xl border-border/50 bg-white/50 backdrop-blur-sm
                             focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                             transition-all duration-200"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                </div>
              </div>

              {/* Email Field */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="email">
                  Email
                </label>
                <div className="relative">
                  <Mail className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input
                    id="email"
                    placeholder="ornek@sirket.com"
                    type="email"
                    data-testid="register-page-email-input"
                    className="pl-12 h-12 rounded-xl border-border/50 bg-white/50 backdrop-blur-sm
                             focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                             transition-all duration-200"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </div>
              </div>

              {/* Password Field */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground" htmlFor="password">
                  Şifre
                </label>
                <div className="relative">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input
                    id="password"
                    type="password"
                    placeholder="Güçlü şifre oluşturun"
                    data-testid="register-page-password-input"
                    className="pl-12 h-12 rounded-xl border-border/50 bg-white/50 backdrop-blur-sm
                             focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                             transition-all duration-200"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
                <p className="text-xs text-muted-foreground ml-1">
                  En az 8 karakter, büyük harf, küçük harf, rakam ve özel karakter (@$!%*?&)
                </p>
              </div>

              {/* Submit Button */}
              <Button
                className="w-full h-12 rounded-xl text-base font-semibold
                         bg-primary hover:bg-primary/90
                         shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30
                         transition-all duration-300 group"
                type="submit"
                isLoading={isLoading}
                data-testid="register-page-submit-button"
              >
                <span>Kayıt Ol</span>
                {!isLoading && (
                  <ArrowRight className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform duration-200" />
                )}
              </Button>
            </form>

            {/* Divider */}
            <div className="relative my-8">
              <div className="divider" />
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="px-4 text-xs font-medium uppercase text-muted-foreground bg-white/80 backdrop-blur-sm rounded-full">
                  veya
                </span>
              </div>
            </div>

            {/* Login Link */}
            <p className="text-center text-sm text-muted-foreground">
              Zaten hesabınız var mı?{' '}
              <Link
                to="/login"
                className="font-semibold text-primary hover:text-primary/80 
                         transition-colors duration-200"
              >
                Giriş Yapın
              </Link>
            </p>
          </div>
        </div>
      </div>

      {/* Right Side - Features (Hidden on mobile) */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-between p-12 relative z-10">
        {/* Logo */}
        <div className="flex items-center gap-3 animate-fade-up">
          <div className="relative">
            <div className="absolute inset-0 bg-primary/20 rounded-2xl blur-xl" />
            <img
              src="/logo-128.png"
              alt="Botla Logo"
              className="relative w-12 h-12 rounded-2xl shadow-lg"
            />
          </div>
          <span className="text-2xl font-bold text-foreground">botla.app</span>
        </div>

        {/* Features Content */}
        <div className="max-w-lg">
          <h1
            className="heading-lg text-foreground mb-6 animate-fade-up text-balance"
            style={{ animationDelay: '100ms' }}
          >
            Müşteri desteğinizi <span className="text-gradient">otomatize edin</span>
          </h1>
          <p className="body-lg mb-10 animate-fade-up" style={{ animationDelay: '150ms' }}>
            Yapay zeka destekli chatbot ile işletmenizi bir üst seviyeye taşıyın.
          </p>

          {/* Feature Cards */}
          <div className="space-y-4">
            {features.map((feature, index) => (
              <div
                key={feature.title}
                className="glass-panel p-5 animate-fade-up"
                style={{ animationDelay: `${200 + index * 50}ms` }}
              >
                <div className="flex items-start gap-4">
                  <div
                    className="flex-shrink-0 w-10 h-10 rounded-xl bg-primary/10 
                                flex items-center justify-center"
                  >
                    <feature.icon className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-foreground mb-1">{feature.title}</h3>
                    <p className="text-sm text-muted-foreground">{feature.description}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div
          className="text-sm text-muted-foreground animate-fade-up"
          style={{ animationDelay: '400ms' }}
        >
          © 2024 botla.app. Tüm hakları saklıdır.
        </div>
      </div>
    </div>
  )
}

export default RegisterPage
