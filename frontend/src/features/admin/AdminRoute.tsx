import { Navigate } from 'react-router-dom'
import { useProfile } from '@/hooks/queries/useProfile'

interface AdminRouteProps {
  children: React.ReactNode
}

/**
 * AdminRoute - Protected route wrapper for admin-only pages
 * 
 * Checks if the current user has is_platform_admin flag.
 * Redirects to /dashboard if user is not an admin.
 */
export function AdminRoute({ children }: AdminRouteProps) {
  const { data: user, isLoading } = useProfile()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="flex flex-col items-center gap-4">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <span className="text-muted-foreground">Yükleniyor...</span>
        </div>
      </div>
    )
  }

  if (!user?.is_platform_admin) {
    return <Navigate to="/dashboard" replace />
  }

  return <>{children}</>
}
