import { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Edit2, Trash2, Zap, Globe, Bolt } from 'lucide-react'
import {
  Action,
  CreateActionRequest,
  getActions,
  createAction,
  updateAction,
  deleteAction,
} from '@/api/action'
import ActionForm from './ActionForm'
import { useToast } from '@/components/ui/toast'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import ActionLogs from './ActionLogs'

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

  return (
    <Tabs defaultValue="list" className="w-full">
      <TabsList className="mb-6">
        <TabsTrigger value="list">Aksiyonlar</TabsTrigger>
        <TabsTrigger value="logs">Geçmiş</TabsTrigger>
      </TabsList>

      <TabsContent value="list" className="mt-0">
        {isEditing ? (
          <>
            <div className="mb-4">
              <Button
                variant="ghost"
                onClick={() => setIsEditing(false)}
                className="pl-0 hover:pl-0 hover:bg-transparent text-muted-foreground hover:text-foreground"
              >
                &larr; Listeye Dön
              </Button>
            </div>
            <ActionForm
              action={editingAction}
              onSave={handleSave}
              onCancel={() => setIsEditing(false)}
              isSaving={isSaving}
            />
          </>
        ) : (
          <Card className="border-0 shadow-none bg-transparent">
            <div className="flex flex-row items-center justify-between mb-6">
              <div>
                <h3 className="text-lg font-semibold">Tanımlı Aksiyonlar</h3>
                <p className="text-sm text-muted-foreground">
                  Botunuzun kullanabileceği yetenekler.
                </p>
              </div>
              {actions.length > 0 && (
                <Button onClick={handleCreate} size="sm" className="gap-2">
                  <Plus className="w-4 h-4" /> Yeni Aksiyon
                </Button>
              )}
            </div>

            <div className="space-y-4">
              {loading ? (
                <div className="text-center py-12 text-muted-foreground">Yükleniyor...</div>
              ) : actions.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 px-4 border-2 border-dashed rounded-xl bg-muted/10 text-center">
                  <div className="bg-primary/10 p-4 rounded-full mb-4">
                    <Bolt className="w-8 h-8 text-primary" />
                  </div>
                  <h3 className="text-xl font-semibold mb-2">Henüz bir aksiyon yok</h3>
                  <p className="text-muted-foreground max-w-md mb-6">
                    Botunuza dış dünya ile konuşma yeteneği kazandırabilirsiniz. Örneğin bir API'dan
                    veri çekebilir veya Zapier ile otomasyon tetikleyebilirsiniz.
                  </p>
                  <Button onClick={handleCreate} className="gap-2">
                    <Plus className="w-4 h-4" /> İlk Aksiyonu Oluştur
                  </Button>

                  <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4 w-full max-w-2xl text-left">
                    <div className="p-4 rounded-lg bg-card border text-sm">
                      <div className="font-medium mb-1 flex items-center gap-2">
                        <Globe className="w-4 h-4 text-blue-500" /> HTTP Request
                      </div>
                      Herhangi bir REST API'a istek atarak veri çekin veya gönderin.
                    </div>
                    <div className="p-4 rounded-lg bg-card border text-sm">
                      <div className="font-medium mb-1 flex items-center gap-2">
                        <Zap className="w-4 h-4 text-orange-500" /> Zapier Webhook
                      </div>
                      5000+ uygulama ile entegre olmak için Zapier kullanın.
                    </div>
                  </div>
                </div>
              ) : (
                <div className="grid gap-4">
                  {actions.map((action) => (
                    <div
                      key={action.id}
                      className="group flex items-center justify-between p-4 border rounded-xl bg-card hover:border-primary/50 transition-all hover:shadow-sm"
                    >
                      <div className="flex items-start gap-4">
                        <div
                          className={`mt-1 p-2.5 rounded-lg ${action.enabled ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}
                        >
                          {action.action_type === 'http' ? (
                            <Globe className="w-5 h-5" />
                          ) : action.action_type === 'zapier' ? (
                            <Zap className="w-5 h-5" />
                          ) : (
                            <Bolt className="w-5 h-5" />
                          )}
                        </div>
                        <div>
                          <div className="font-medium flex items-center gap-2 text-base">
                            {action.name}
                            {!action.enabled && (
                              <span className="text-[10px] bg-muted px-1.5 py-0.5 rounded border uppercase tracking-wider text-muted-foreground">
                                Pasif
                              </span>
                            )}
                          </div>
                          <div className="text-sm text-muted-foreground line-clamp-1 mt-0.5">
                            {action.description || 'Açıklama yok'}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(action)}
                          className="h-8 w-8 p-0"
                        >
                          <Edit2 className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                          onClick={() => handleDelete(action.id)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </Card>
        )}
      </TabsContent>

      <TabsContent value="logs" className="mt-0">
        <ActionLogs chatbotId={chatbotId} />
      </TabsContent>
    </Tabs>
  )
}
