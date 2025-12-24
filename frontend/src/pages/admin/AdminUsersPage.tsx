/**
 * AdminUsersPage - User management page for platform admins
 * Lists all users with search, filter, and management capabilities
 */
export function AdminUsersPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Kullanıcılar</h1>
      <p className="text-muted-foreground">
        Platform kullanıcılarını görüntüle ve yönet.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Kullanıcı listesi burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
