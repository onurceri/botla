/**
 * AdminSourcesPage - Data sources management page
 * Lists all data sources and their processing status
 */
export function AdminSourcesPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Kaynaklar</h1>
      <p className="text-muted-foreground">
        Veri kaynaklarını ve işleme durumlarını görüntüle.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Kaynak listesi burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
