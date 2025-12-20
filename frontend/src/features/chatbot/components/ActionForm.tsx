import { useState } from 'react'
import { Card, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Action, CreateActionRequest } from '@/api/action'
import HeadersEditor from './HeadersEditor'
import ParameterBuilder from './ParameterBuilder'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'

interface Props {
  action?: Action
  onSave: (action: CreateActionRequest) => Promise<void>
  onCancel: () => void
  isSaving: boolean
}

export default function ActionForm({ action, onSave, onCancel, isSaving }: Props) {
  // General
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

  // Parameters
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
        alert('Headers configuration is invalid')
        return
      }
    } else if (actionType === 'zapier') {
      config = {
        webhook_url: webhookUrl
      }
    }

    let parsedParams: any = {}
    try {
      if (parameters) {
        parsedParams = JSON.parse(parameters)
      }
    } catch {
      alert('Parameters configuration is invalid')
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
    <Card className="border-0 shadow-none">
      <form onSubmit={handleSubmit}>
        <div className="space-y-8 p-6">
          
          {/* Section 1: General Information */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <div className="bg-primary/10 text-primary w-6 h-6 rounded-full flex items-center justify-center text-xs">1</div>
              Genel Bilgiler
            </h3>
            <div className="grid gap-4 pl-8">
              <div className="grid gap-2">
                <Label>Aksiyon İsmi</Label>
                <Input required value={name} onChange={e => setName(e.target.value)} placeholder="Örn: Sipariş Durumu Sorgula" />
                <p className="text-xs text-muted-foreground">AI'ın bu aksiyonu ne zaman kullanacağını anlaması için açıklayıcı bir isim verin.</p>
              </div>
              
              <div className="grid gap-2">
                <Label>Açıklama</Label>
                <Input value={description} onChange={e => setDescription(e.target.value)} placeholder="Bu aksiyon ne işe yarar?" />
              </div>

              <div className="grid gap-2">
                <Label>Tip</Label>
                <select 
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                  value={actionType}
                  onChange={e => setActionType(e.target.value as any)}
                >
                  <option value="http">HTTP Request</option>
                  <option value="zapier">Zapier Webhook</option>
                </select>
              </div>

              <div className="flex items-center space-x-2 pt-2">
                <Switch 
                  id="enabled"
                  checked={enabled}
                  onCheckedChange={setEnabled}
                />
                <Label htmlFor="enabled">Aktif</Label>
              </div>
            </div>
          </div>

          <div className="h-px bg-border" />

          {/* Section 2: Configuration */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <div className="bg-primary/10 text-primary w-6 h-6 rounded-full flex items-center justify-center text-xs">2</div>
              Yapılandırma
            </h3>
            <div className="pl-8">
              {actionType === 'http' && (
                <div className="space-y-6">
                    <div className="grid gap-2">
                    <Label>Endpoint URL</Label>
                    <Input required value={url} onChange={e => setUrl(e.target.value)} placeholder="https://api.example.com/v1/status" />
                  </div>
                  <div className="grid gap-2">
                    <Label>Method</Label>
                    <select 
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                      value={method}
                      onChange={e => setMethod(e.target.value)}
                    >
                      <option value="GET">GET</option>
                      <option value="POST">POST</option>
                      <option value="PUT">PUT</option>
                      <option value="DELETE">DELETE</option>
                    </select>
                  </div>
                  
                  <HeadersEditor value={headers} onChange={setHeaders} />
                </div>
              )}

              {actionType === 'zapier' && (
                <div className="grid gap-2">
                  <Label>Zapier Webhook URL</Label>
                  <Input required value={webhookUrl} onChange={e => setWebhookUrl(e.target.value)} placeholder="https://hooks.zapier.com/..." />
                </div>
              )}
            </div>
          </div>

          <div className="h-px bg-border" />

          {/* Section 3: Parameters */}
          <div className="space-y-4">
             <h3 className="text-lg font-semibold flex items-center gap-2">
              <div className="bg-primary/10 text-primary w-6 h-6 rounded-full flex items-center justify-center text-xs">3</div>
              Parametreler
            </h3>
             <div className="pl-8">
                <div className="mb-4 text-sm text-muted-foreground">
                  AI'ın bu aksiyonu çalıştırırken kullanıcıdan hangi bilgileri alması gerektiğini tanımlayın.
                </div>
                <ParameterBuilder value={parameters} onChange={setParameters} />
             </div>
          </div>

        </div>

        <CardFooter className="flex justify-between py-4 border-t bg-muted/10">
          <Button type="button" variant="outline" onClick={onCancel}>İptal</Button>
          <Button type="submit" disabled={isSaving}>
            {isSaving ? 'Kaydediliyor...' : 'Kaydet ve Tamamla'}
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}
