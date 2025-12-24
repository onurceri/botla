/**
 * AdminChatbotsPage - Chatbot management page
 * Lists all chatbots across the platform
 */
export function AdminChatbotsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Chatbotlar</h1>
      <p className="text-muted-foreground">
        Platform genelindeki tüm chatbotları görüntüle ve yönet.
      </p>

      {/* Placeholder content */}
      <div className="bg-card rounded-lg p-6 border border-border">
        <p className="text-center text-muted-foreground py-8">
          Chatbot listesi burada görüntülenecek.
        </p>
      </div>
    </div>
  )
}
