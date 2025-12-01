import { useState, useEffect } from 'react'
import { login, logout, protectedPing } from '@/api/auth'

export const useAuth = () => {
  const [user, setUser] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('botla_token')
    if (!token) {
      setLoading(false)
      return
    }
    
    // We rely on the axios interceptor to handle token refresh if the token is expired.
    // So we just try to ping.
    protectedPing()
      .then((d) => setUser({ id: d.user_id }))
      .catch(() => {
        // If ping fails (and refresh also failed in interceptor), user is null
        setUser(null)
      })
      .finally(() => setLoading(false))
  }, [])

  const signIn = async (email: string, password: string) => {
    const data = await login(email, password)
    localStorage.setItem('botla_token', data.token)
    localStorage.setItem('botla_refresh_token', data.refresh_token)
    const ping = await protectedPing().catch(() => null)
    setUser(ping ? { id: ping.user_id } : { id: 'me' })
  }

  const signOut = () => {
    const rt = localStorage.getItem('botla_refresh_token')
    if (rt) logout(rt).catch(() => {})
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    setUser(null)
    window.location.replace('/login')
  }

  const isAuthenticated = !!user
  return { user, loading, isAuthenticated, signIn, signOut }
}
