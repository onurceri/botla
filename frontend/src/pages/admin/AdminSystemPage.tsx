/**
 * AdminSystemPage - System health monitoring page
 * Shows detailed status of all dependencies (DB, Redis, Qdrant, OpenAI)
 */
export function AdminSystemPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Sistem Durumu</h1>
      <p className="text-muted-foreground">
        Sistem bileşenlerinin sağlık durumunu izle.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Sistem sağlık bilgileri burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
