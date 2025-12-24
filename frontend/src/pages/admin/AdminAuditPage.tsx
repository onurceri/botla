/**
 * AdminAuditPage - Audit log page
 * Shows admin actions and changes
 */
export function AdminAuditPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Denetim Günlüğü</h1>
      <p className="text-muted-foreground">
        Admin işlemlerini ve değişiklikleri izle.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Denetim günlüğü burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
