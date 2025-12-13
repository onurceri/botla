import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Mail, Lock, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
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
      localStorage.setItem('botla_token', data.token)
      localStorage.setItem('botla_refresh_token', data.refresh_token)
      toast('Giriş başarılı! Yönlendiriliyorsunuz...', 'success')
      navigate('/dashboard')
    } catch {
      toast('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen grid lg:grid-cols-2 bg-background text-foreground">
      {/* Left Side - Branding */}
      <div className="hidden lg:flex flex-col justify-between p-12 bg-primary/5 relative overflow-hidden border-r border-border">
        {/* Abstract Background Shapes */}
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-accent/20 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2" />
        
        <div className="relative z-10 flex items-center gap-2 font-bold text-2xl">
          <img src="/logo-128.png" alt="Botla Logo" className="w-10 h-10 rounded-xl shadow-lg shadow-primary/20" />
          <span className="text-foreground">Botla.co</span>
        </div>

        <div className="relative z-10 max-w-lg">
          <h1 className="text-4xl font-bold mb-6 leading-tight text-foreground">
            Kendi verilerinizle eğitilmiş <span className="text-primary">akıllı chatbotlar</span> oluşturun.
          </h1>
          <p className="text-lg text-muted-foreground">
            PDF, URL veya metin dosyalarınızı yükleyin, dakikalar içinde sitenize entegre edin.
          </p>
        </div>

        <div className="relative z-10 text-sm text-muted-foreground">
          © 2024 Botla.co. Tüm hakları saklıdır.
        </div>
      </div>

      {/* Right Side - Form */}
      <div className="flex items-center justify-center p-6 lg:p-12 bg-background">
        <Card className="w-full max-w-md border-0 shadow-none bg-transparent">
          <CardHeader className="space-y-1 px-0">
            <CardTitle className="text-2xl font-bold">Hoş Geldiniz</CardTitle>
            <CardDescription>Hesabınıza giriş yapın</CardDescription>
          </CardHeader>
          <CardContent className="px-0">
            <form onSubmit={handleLogin} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="email">
                  Email
                </label>
                <div className="relative">
                  <Mail className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input 
                    id="email" 
                    placeholder="ornek@sirket.com" 
                    type="email" 
                    className="pl-9 h-11"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="password">
                    Şifre
                  </label>
                  <Link to="/forgot-password" className="text-sm font-medium text-primary hover:underline">
                    Şifremi unuttum?
                  </Link>
                </div>
                <div className="relative">
                  <Lock className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input 
                    id="password" 
                    type="password" 
                    placeholder="********"
                    className="pl-9 h-11"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
              </div>
              <Button className="w-full group h-11" type="submit" isLoading={isLoading}>
                Giriş Yap
                {!isLoading && <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />}
              </Button>
            </form>
          </CardContent>
          <CardFooter className="px-0 flex flex-col gap-4">
            <div className="relative w-full">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  veya
                </span>
              </div>
            </div>
            <div className="text-center text-sm">
              Hesabınız yok mu?{' '}
              <Link to="/register" className="font-medium text-primary hover:underline">
                Kayıt Olun
              </Link>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}

export default LoginPage
