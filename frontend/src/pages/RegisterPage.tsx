import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Bot, Mail, Lock, User, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'

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
      await api.post('/api/v1/auth/register', { full_name: name, email, password })
      toast('Kayıt başarılı! Giriş yapabilirsiniz.', 'success')
      navigate('/login')
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Kayıt başarısız. Lütfen tekrar deneyin.'
      toast(errorMessage, 'error')
      setErrorMsg(errorMessage)
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
          <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-primary-foreground shadow-lg shadow-primary/20">
            <Bot className="w-6 h-6" />
          </div>
          <span className="text-foreground">Botla.co</span>
        </div>

        <div className="relative z-10 max-w-lg">
          <h1 className="text-4xl font-bold mb-6 leading-tight text-foreground">
            Müşteri desteğinizi <span className="text-primary">otomatize edin</span>.
          </h1>
          <p className="text-lg text-muted-foreground">
            7/24 çalışan, yorulmayan ve sürekli öğrenen bir asistan ile işinizi büyütün.
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
            <CardTitle className="text-2xl font-bold">Hesap Oluştur</CardTitle>
            <CardDescription>Hemen başlayın</CardDescription>
          </CardHeader>
          <CardContent className="px-0">
            {errorMsg && (
              <div className="mb-4 text-sm text-red-600" role="alert">{errorMsg}</div>
            )}
            <form onSubmit={handleRegister} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="name">
                  Ad Soyad
                </label>
                <div className="relative">
                  <User className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input 
                    id="name" 
                    placeholder="Adınız Soyadınız" 
                    className="pl-9"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                </div>
              </div>
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
                    className="pl-9"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="password">
                  Şifre
                </label>
                <div className="relative">
                  <Lock className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input 
                    id="password" 
                    type="password" 
                    className="pl-9"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
              </div>
              <Button className="w-full group" type="submit" isLoading={isLoading}>
                Kayıt Ol
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
              Zaten hesabınız var mı?{' '}
              <Link to="/login" className="font-medium text-primary hover:underline">
                Giriş Yapın
              </Link>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}

export default RegisterPage
