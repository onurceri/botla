import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Mail, Lock, ArrowRight, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'

const LoginPage = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !password) {
      toast('Lütfen tüm alanları doldurun.', 'error')
      return
    }

    setIsLoading(true)
    try {
      const { data } = await api.post('/api/v1/auth/login', { email, password })
      window.localStorage.setItem('botla_token', data.token)
      window.localStorage.setItem('botla_refresh_token', data.refresh_token)
      toast('Giriş başarılı! Yönlendiriliyorsunuz...', 'success')
      navigate('/dashboard')
    } catch {
      toast('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex bg-background relative overflow-hidden">
      {/* Animated Background Mesh */}
      <div className="absolute inset-0 gradient-mesh opacity-60" />

      {/* Floating Decorative Elements */}
      <div className="absolute top-20 left-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl animate-float" />
      <div
        className="absolute bottom-20 right-20 w-96 h-96 bg-accent/30 rounded-full blur-3xl animate-float"
        style={{ animationDelay: '-3s' }}
      />
      <div className="absolute top-1/2 left-1/3 w-64 h-64 bg-primary/5 blob-animated blur-2xl" />

      {/* Left Side - Premium Branding (Hidden on mobile) */}
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

        {/* Hero Content */}
        <div className="max-w-lg animate-fade-up" style={{ animationDelay: '100ms' }}>
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/10 text-primary text-sm font-medium mb-8">
            <Sparkles className="w-4 h-4" />
            <span>Yapay Zeka Destekli Çözümler</span>
          </div>
          <h1 className="heading-xl text-foreground mb-6 text-balance">
            Kendi verilerinizle eğitilmiş <span className="text-gradient">akıllı chatbotlar</span>{' '}
            oluşturun.
          </h1>
          <p className="body-lg mb-8">
            PDF, URL veya metin dosyalarınızı yükleyin, dakikalar içinde sitenize entegre edin. 7/24
            çalışan, yorulmayan ve sürekli öğrenen bir asistan ile işinizi büyütün.
          </p>

          {/* Trust Indicators */}
          <div className="flex items-center gap-6 text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-success animate-pulse-soft" />
              <span>Ücretsiz deneyin</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-primary" />
              <span>Dakikalar içinde kurulum</span>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div
          className="text-sm text-muted-foreground animate-fade-up"
          style={{ animationDelay: '200ms' }}
        >
          © 2024 botla.app. Tüm hakları saklıdır.
        </div>
      </div>

      {/* Right Side - Login Form */}
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
            data-testid="login-page"
          >
            <div className="text-center mb-8">
              <h2
                className="heading-md text-foreground mb-2"
                data-testid="login-page-title"
              >
                Hoş Geldiniz
              </h2>
              <p className="body-sm">Hesabınıza giriş yapın</p>
            </div>

            <form onSubmit={handleLogin} className="space-y-5 stagger-children">
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
                    data-testid="login-page-email-input"
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
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium text-foreground" htmlFor="password">
                    Şifre
                  </label>
                  <Link
                    to="/forgot-password"
                    className="text-sm font-medium text-primary hover:text-primary/80
                             transition-colors duration-200"
                    data-testid="login-page-forgot-password-link"
                  >
                    Şifremi unuttum?
                  </Link>
                </div>
                <div className="relative">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    data-testid="login-page-password-input"
                    className="pl-12 h-12 rounded-xl border-border/50 bg-white/50 backdrop-blur-sm
                             focus:bg-white focus:border-primary/50 focus:ring-2 focus:ring-primary/20
                             transition-all duration-200"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
              </div>

              {/* Submit Button */}
              <Button
                className="w-full h-12 rounded-xl text-base font-semibold
                         bg-primary hover:bg-primary/90
                         shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30
                         transition-all duration-300 group"
                type="submit"
                isLoading={isLoading}
                data-testid="login-page-submit-button"
              >
                <span>Giriş Yap</span>
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

            {/* Register Link */}
            <p className="text-center text-sm text-muted-foreground">
              Hesabınız yok mu?{' '}
              <Link
                to="/register"
                className="font-semibold text-primary hover:text-primary/80 
                         transition-colors duration-200"
              >
                Kayıt Olun
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default LoginPage
