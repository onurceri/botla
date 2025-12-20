import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Plus, Trash2 } from 'lucide-react'

interface Props {
  value: string // JSON string
  onChange: (value: string) => void
}

interface HeaderPair {
  key: string
  value: string
}

export default function HeadersEditor({ value, onChange }: Props) {
  const [pairs, setPairs] = useState<HeaderPair[]>([])

  useEffect(() => {
    try {
      if (!value || value === '{}') {
        setPairs([])
        return
      }
      const parsed = JSON.parse(value)
      const newPairs = Object.entries(parsed).map(([k, v]) => ({
        key: k,
        value: String(v)
      }))
      setPairs(newPairs)
    } catch {
      // If invalid JSON, ignore or maybe handle error? 
      // For now we assume value comes from valid source or empty
    }
  }, []) // On mount typically, or we could watch value if we want two-way sync but be careful of loops

  const updateParent = (currentPairs: HeaderPair[]) => {
    const obj = currentPairs.reduce((acc, pair) => {
      if (pair.key) {
        acc[pair.key] = pair.value
      }
      return acc
    }, {} as Record<string, string>)
    onChange(JSON.stringify(obj, null, 2))
  }

  const addRow = () => {
    const newPairs = [...pairs, { key: '', value: '' }]
    setPairs(newPairs)
    // Don't update parent yet, wait for input? Or update immediately?
    // Updating immediately with empty keys might represent empty object, which is fine.
  }

  const removeRow = (index: number) => {
    const newPairs = pairs.filter((_, i) => i !== index)
    setPairs(newPairs)
    updateParent(newPairs)
  }

  const handleChange = (index: number, field: 'key' | 'value', text: string) => {
    const newPairs = [...pairs]
    newPairs[index][field] = text
    setPairs(newPairs)
    updateParent(newPairs)
  }

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">Headers</div>
      <div className="space-y-2">
        {pairs.map((pair, index) => (
          <div key={index} className="flex gap-2">
            <Input 
              placeholder="Key (e.g. Authorization)" 
              value={pair.key} 
              onChange={(e) => handleChange(index, 'key', e.target.value)}
              className="flex-1"
            />
            <Input 
              placeholder="Value (e.g. Bearer token)" 
              value={pair.value} 
              onChange={(e) => handleChange(index, 'value', e.target.value)}
              className="flex-1"
            />
            <Button variant="ghost" size="icon" onClick={() => removeRow(index)} className="text-destructive hover:text-destructive">
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        ))}
      </div>
      <Button type="button" variant="outline" size="sm" onClick={addRow} className="gap-2">
        <Plus className="w-4 h-4" /> Header Ekle
      </Button>
    </div>
  )
}
