/**
 * AdminErrorsPage - Error logs page
 * Shows error logs with filtering by severity
 */
export function AdminErrorsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Hatalar</h1>
      <p className="text-muted-foreground">
        Sistem hata günlüklerini görüntüle.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Hata günlükleri burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
