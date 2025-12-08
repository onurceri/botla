import { useState, useEffect } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Edit2, Trash2, Zap, Globe, Bolt } from 'lucide-react'
import { Action, CreateActionRequest, getActions, createAction, updateAction, deleteAction } from '@/api/action'
import ActionForm from './ActionForm'
import { useToast } from '@/components/ui/toast'

interface Props {
  chatbotId: string
}

export default function ActionList({ chatbotId }: Props) {
  const [actions, setActions] = useState<Action[]>([])
  const [loading, setLoading] = useState(true)
  const [isEditing, setIsEditing] = useState(false)
  const [editingAction, setEditingAction] = useState<Action | undefined>(undefined)
  const [isSaving, setIsSaving] = useState(false)
  const { toast } = useToast()

  const fetchActions = async () => {
    setLoading(true)
    try {
      const data = await getActions(chatbotId)
      setActions(data)
    } catch (error) {
      console.error(error)
      toast('Aksiyonlar yüklenemedi.', 'error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (chatbotId) {
      fetchActions()
    }
  }, [chatbotId])

  const handleCreate = () => {
    setEditingAction(undefined)
    setIsEditing(true)
  }

  const handleEdit = (action: Action) => {
    setEditingAction(action)
    setIsEditing(true)
  }

  const handleDelete = async (actionId: string) => {
    if (!confirm('Bu aksiyonu silmek istediğinize emin misiniz?')) return
    try {
      await deleteAction(chatbotId, actionId)
      toast('Aksiyon silindi.', 'success')
      fetchActions()
    } catch {
      toast('Silme işlemi başarısız oldu.', 'error')
    }
  }

  const handleSave = async (req: CreateActionRequest) => {
    setIsSaving(true)
    try {
      if (editingAction) {
        await updateAction(chatbotId, editingAction.id, req)
        toast('Aksiyon güncellendi.', 'success')
      } else {
        await createAction(chatbotId, req)
        toast('Aksiyon oluşturuldu.', 'success')
      }
      setIsEditing(false)
      fetchActions()
    } catch (error) {
      console.error(error)
      toast('Kaydetme işlemi başarısız oldu.', 'error')
    } finally {
      setIsSaving(false)
    }
  }

  if (isEditing) {
    return (
      <ActionForm 
        action={editingAction} 
        onSave={handleSave} 
        onCancel={() => setIsEditing(false)} 
        isSaving={isSaving}
      />
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Aksiyonlar</CardTitle>
          <CardDescription>Botunuzun yeteneklerini genişletin.</CardDescription>
        </div>
        <Button onClick={handleCreate} size="sm" className="gap-2">
          <Plus className="w-4 h-4" /> Yeni Aksiyon
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? (
          <div className="text-center py-8 text-muted-foreground">Yükleniyor...</div>
        ) : actions.length === 0 ? (
          <div className="text-center py-8 border rounded-lg bg-muted/20">
            <Zap className="w-10 h-10 text-muted-foreground mx-auto mb-3" />
            <h3 className="text-lg font-medium">Henüz aksiyon yok</h3>
            <p className="text-sm text-muted-foreground mb-4">Botunuzun harici sistemlerle konuşması için bir aksiyon ekleyin.</p>
            <Button variant="outline" onClick={handleCreate}>Aksiyon Ekle</Button>
          </div>
        ) : (
          <div className="grid gap-4">
            {actions.map(action => (
              <div key={action.id} className="flex items-center justify-between p-4 border rounded-lg bg-card hover:bg-accent/5 transition-colors">
                <div className="flex items-start gap-3">
                  <div className={`mt-1 p-2 rounded-md ${action.enabled ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}>
                    {action.action_type === 'http' ? <Globe className="w-4 h-4" /> : 
                     action.action_type === 'zapier' ? <Zap className="w-4 h-4" /> : 
                     <Bolt className="w-4 h-4" />}
                  </div>
                  <div>
                    <div className="font-medium flex items-center gap-2">
                      {action.name}
                      {!action.enabled && <span className="text-xs bg-muted px-2 py-0.5 rounded text-muted-foreground">Pasif</span>}
                    </div>
                    <div className="text-sm text-muted-foreground line-clamp-1">{action.description}</div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" size="icon" onClick={() => handleEdit(action)}>
                    <Edit2 className="w-4 h-4" />
                  </Button>
                  <Button variant="ghost" size="icon" className="text-destructive hover:text-destructive" onClick={() => handleDelete(action.id)}>
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
