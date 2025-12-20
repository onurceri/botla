import { useState } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Action, CreateActionRequest } from '@/api/action'

interface Props {
  action?: Action
  onSave: (action: CreateActionRequest) => Promise<void>
  onCancel: () => void
  isSaving: boolean
}

export default function ActionForm({ action, onSave, onCancel, isSaving }: Props) {
  const [name, setName] = useState(action?.name || '')
  const [description, setDescription] = useState(action?.description || '')
  const [actionType, setActionType] = useState<'builtin' | 'http' | 'zapier'>(action?.action_type || 'http')
  const [enabled, setEnabled] = useState(action?.enabled ?? true)
  
  // HTTP Config
  const [url, setUrl] = useState(action?.config?.url || '')
  const [method, setMethod] = useState(action?.config?.method || 'POST')
  const [headers, setHeaders] = useState<string>(action?.config?.headers ? JSON.stringify(action.config.headers, null, 2) : '{}')
  
  // Zapier Config
  const [webhookUrl, setWebhookUrl] = useState(action?.config?.webhook_url || '')

  // Parameters (JSON Schema)
  const [parameters, setParameters] = useState<string>(action?.parameters ? JSON.stringify(action.parameters, null, 2) : '{\n  "type": "object",\n  "properties": {}\n}')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    let config: any = {}
    if (actionType === 'http') {
      try {
        config = {
          url,
          method,
          headers: JSON.parse(headers)
        }
      } catch {
        alert('Headers JSON formatı hatalı')
        return
      }
    } else if (actionType === 'zapier') {
      config = {
        webhook_url: webhookUrl
      }
    }

    let parsedParams: any = {}
    try {
      parsedParams = JSON.parse(parameters)
    } catch {
      alert('Parameters JSON formatı hatalı')
      return
    }

    await onSave({
      name,
      description,
      action_type: actionType,
      config,
      parameters: parsedParams,
      enabled
    })
  }

  return (
    <Card>
      <form onSubmit={handleSubmit}>
        <CardHeader>
          <CardTitle>{action ? 'Aksiyon Düzenle' : 'Yeni Aksiyon'}</CardTitle>
          <CardDescription>
            Chatbotunuzun harici sistemlerle konuşmasını sağlayın.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <label className="text-sm font-medium">Aksiyon İsmi</label>
            <Input required value={name} onChange={e => setName(e.target.value)} placeholder="Örn: Sipariş Durumu Sorgula" />
          </div>
          
          <div className="grid gap-2">
            <label className="text-sm font-medium">Açıklama</label>
            <Input value={description} onChange={e => setDescription(e.target.value)} placeholder="Bu aksiyon ne işe yarar?" />
          </div>

          <div className="grid gap-2">
            <label className="text-sm font-medium">Tip</label>
            <select 
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              value={actionType}
              onChange={e => setActionType(e.target.value as any)}
            >
              <option value="http">HTTP Request</option>
              <option value="zapier">Zapier Webhook</option>
              {/* <option value="builtin">Built-in</option> */}
            </select>
          </div>

          {actionType === 'http' && (
            <div className="space-y-4 border p-4 rounded-md">
              <div className="grid gap-2">
                <label className="text-sm font-medium">URL</label>
                <Input required value={url} onChange={e => setUrl(e.target.value)} placeholder="https://api.example.com/v1/status" />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Method</label>
                <select 
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  value={method}
                  onChange={e => setMethod(e.target.value)}
                >
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                </select>
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Headers (JSON)</label>
                <textarea 
                  className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  value={headers}
                  onChange={e => setHeaders(e.target.value)}
                  placeholder='{"Authorization": "Bearer token"}'
                />
              </div>
            </div>
          )}

          {actionType === 'zapier' && (
             <div className="space-y-4 border p-4 rounded-md">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Zapier Webhook URL</label>
                <Input required value={webhookUrl} onChange={e => setWebhookUrl(e.target.value)} placeholder="https://hooks.zapier.com/..." />
              </div>
             </div>
          )}

          <div className="grid gap-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium">Parametreler (JSON Schema)</label>
              <Button 
                type="button" 
                variant="ghost" 
                size="sm" 
                className="h-6 text-xs"
                onClick={() => setParameters('{\n  "type": "object",\n  "properties": {\n    "city": {\n      "type": "string",\n      "description": "Kullanıcının sorduğu şehir"\n    }\n  },\n  "required": ["city"]\n}')}
              >
                Örnek Doldur
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              AI'ın parametreleri tanıması için <a href="https://json-schema.org/learn/getting-started-step-by-step" target="_blank" rel="noopener noreferrer" className="underline text-primary">JSON Schema</a> formatında giriniz.
              Örneğin bir şehir parametresi için "Örnek Doldur" butonunu kullanabilirsiniz.
            </p>
            <textarea 
              className="flex min-h-[150px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 font-mono"
              value={parameters}
              onChange={e => setParameters(e.target.value)}
              placeholder='{\n  "type": "object",\n  "properties": { ... }\n}'
            />
          </div>

          <div className="flex items-center space-x-2">
            <input 
              type="checkbox" 
              id="enabled" 
              checked={enabled} 
              onChange={e => setEnabled(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
            />
            <label htmlFor="enabled" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
              Aktif
            </label>
          </div>

        </CardContent>
        <CardFooter className="flex justify-between">
          <Button type="button" variant="outline" onClick={onCancel}>İptal</Button>
          <Button type="submit" disabled={isSaving}>{isSaving ? 'Kaydediliyor...' : 'Kaydet'}</Button>
        </CardFooter>
      </form>
    </Card>
  )
}
