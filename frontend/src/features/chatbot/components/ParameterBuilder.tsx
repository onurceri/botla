import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Plus, Trash2 } from 'lucide-react'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'

interface Props {
  value: string // JSON Schema string
  onChange: (value: string) => void
}

interface Parameter {
  id: string // internal id for list management
  name: string
  type: string
  description: string
  required: boolean
}

export default function ParameterBuilder({ value, onChange }: Props) {
  const [params, setParams] = useState<Parameter[]>([])

  useEffect(() => {
    try {
      if (!value) return
      const schema = JSON.parse(value)
      if (schema.type !== 'object' || !schema.properties) return

      const loadedParams: Parameter[] = []
      const requiredList = (schema.required as string[]) || []

      Object.entries(schema.properties).forEach(([key, val]: [string, any]) => {
        loadedParams.push({
          id: Math.random().toString(36).substr(2, 9),
          name: key,
          type: val.type || 'string',
          description: val.description || '',
          required: requiredList.includes(key),
        })
      })
      setParams(loadedParams)
    } catch {
      // Invalid JSON, ignore
    }
  }, []) // On mount

  const updateParent = (currentParams: Parameter[]) => {
    const properties: Record<string, any> = {}
    const required: string[] = []

    currentParams.forEach((p) => {
      if (!p.name) return
      properties[p.name] = {
        type: p.type,
        description: p.description,
      }
      if (p.required) {
        required.push(p.name)
      }
    })

    const schema = {
      type: 'object',
      properties,
      required: required.length > 0 ? required : undefined,
    }

    onChange(JSON.stringify(schema, null, 2))
  }

  const addParam = () => {
    const newParams = [
      ...params,
      {
        id: Math.random().toString(36).substr(2, 9),
        name: '',
        type: 'string',
        description: '',
        required: true,
      },
    ]
    setParams(newParams)
  }

  const removeParam = (id: string) => {
    const newParams = params.filter((p) => p.id !== id)
    setParams(newParams)
    updateParent(newParams)
  }

  const handleChange = (id: string, field: keyof Parameter, val: any) => {
    const newParams = params.map((p) => {
      if (p.id === id) {
        return { ...p, [field]: val }
      }
      return p
    })
    setParams(newParams)
    updateParent(newParams)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Label>Parametreler</Label>
        <Button type="button" variant="outline" size="sm" onClick={addParam} className="gap-2">
          <Plus className="w-4 h-4" /> Parametre Ekle
        </Button>
      </div>

      {params.length === 0 && (
        <div className="text-sm text-muted-foreground border border-dashed rounded-lg p-8 text-center">
          Bu aksiyon için henüz bir parametre tanımlanmamış.
        </div>
      )}

      <div className="space-y-4">
        {params.map((param) => (
          <div
            key={param.id}
            className="grid grid-cols-1 md:grid-cols-12 gap-4 items-start border p-4 rounded-lg bg-muted/20"
          >
            <div className="md:col-span-3 space-y-2">
              <Label className="text-xs">Parametre Adı</Label>
              <Input
                placeholder="Örn: city"
                value={param.name}
                onChange={(e) => handleChange(param.id, 'name', e.target.value)}
              />
            </div>

            <div className="md:col-span-2 space-y-2">
              <Label className="text-xs">Tip</Label>
              <Select value={param.type} onValueChange={(v) => handleChange(param.id, 'type', v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="string">Metin (String)</SelectItem>
                  <SelectItem value="number">Sayı (Number)</SelectItem>
                  <SelectItem value="boolean">Mantıksal (Boolean)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="md:col-span-4 space-y-2">
              <Label className="text-xs">Açıklama (AI için önemli)</Label>
              <Input
                placeholder="Örn: Kullanıcının sorduğu şehir"
                value={param.description}
                onChange={(e) => handleChange(param.id, 'description', e.target.value)}
              />
            </div>

            <div className="md:col-span-2 space-y-2 flex flex-col justify-center h-full pt-6">
              <div className="flex items-center space-x-2">
                <Switch
                  id={`req-${param.id}`}
                  checked={param.required}
                  onCheckedChange={(c) => handleChange(param.id, 'required', c)}
                />
                <Label htmlFor={`req-${param.id}`} className="text-xs cursor-pointer">
                  Zorunlu
                </Label>
              </div>
            </div>

            <div className="md:col-span-1 pt-6 flex justify-end">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => removeParam(param.id)}
                className="text-destructive hover:text-destructive h-8 w-8"
              >
                <Trash2 className="w-4 h-4" />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
