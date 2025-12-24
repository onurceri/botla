/**
 * AdminDashboardPage - Main overview page for admin dashboard
 * Shows platform statistics and health status
 */
export function AdminDashboardPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Genel Bakış</h1>
      <p className="text-muted-foreground">
        Admin dashboard genel bakış sayfası - İstatistikler ve sistem durumu burada görüntülenecek.
      </p>

      {/* Placeholder for stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-card rounded-lg p-6 border border-border animate-pulse">
            <div className="h-4 bg-muted rounded w-1/2 mb-2" />
            <div className="h-8 bg-muted rounded w-3/4" />
          </div>
        ))}
      </div>

      {/* Placeholder for panels */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-card rounded-lg p-6 border border-border animate-pulse">
          <div className="h-6 bg-muted rounded w-1/3 mb-4" />
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-4 bg-muted rounded" />
            ))}
          </div>
        </div>
        <div className="bg-card rounded-lg p-6 border border-border animate-pulse">
          <div className="h-6 bg-muted rounded w-1/3 mb-4" />
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-4 bg-muted rounded" />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
