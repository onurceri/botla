/**
 * AdminOrganizationsPage - Organization management page
 * Lists all organizations on the platform
 */
export function AdminOrganizationsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Organizasyonlar</h1>
      <p className="text-muted-foreground">
        Platform organizasyonlarını görüntüle ve yönet.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Organizasyon listesi burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
