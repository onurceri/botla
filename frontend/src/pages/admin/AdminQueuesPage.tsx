/**
 * AdminQueuesPage - Queue management page
 * Shows queue statistics and stuck jobs
 */
export function AdminQueuesPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Kuyruklar</h1>
      <p className="text-muted-foreground">
        İş kuyruklarını ve takılmış görevleri izle.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Kuyruk bilgileri burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
