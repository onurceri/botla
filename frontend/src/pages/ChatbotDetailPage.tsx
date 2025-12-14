import { useState, useRef, useEffect } from 'react'
import { useParams, useNavigate, Outlet, useLocation } from 'react-router-dom'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import HeaderActions from '@/features/chatbot/components/HeaderActions'
import { ChatbotSidebar } from '@/features/chatbot/components/ChatbotSidebar'
import NewChatbotForm from '@/features/chatbot/components/NewChatbotForm'
import { useOrganization } from '@/features/organization/context/OrganizationContext'
import { ChatbotProvider, useChatbotContext } from '@/features/chatbot/context/ChatbotContext'
import { useToastErrors } from '@/features/chatbot/hooks/useToastErrors'

function ChatbotDetailContent() {
  const { id = '' } = useParams()
  const isNew = id === 'new'
  const navigate = useNavigate()
  const location = useLocation()
  const { toast } = useToast()
  const toasts = useToastErrors()
  const { currentWorkspace, isLoading: isOrgLoading } = useOrganization()
  
  const {
    name, setName,
    description, setDescription,
    validate,
    buildPayload
  } = useChatbotContext()

  const [isCreating, setIsCreating] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  // Test Chat State
  const [chatHistory, setChatHistory] = useState<{role: 'user' | 'assistant', content: string}[]>([])

  // Redirect for legacy query params
  useEffect(() => {
    try {
      const u = new URL(window.location.href)
      const tab = u.searchParams.get('tab')
      if (tab && !location.pathname.includes(tab)) {
         navigate(tab, { replace: true })
      }
    } catch {}
  }, [])

  const handleCreate = async () => {
    if (!validate()) {
      toasts.error('Lütfen bir bot ismi girin.')
      return
    }

    if (!currentWorkspace) {
      toasts.error('Lütfen önce bir çalışma alanı seçin.')
      return
    }

    setIsCreating(true)
    const payload = buildPayload()

    try {
      const { data } = await api.post('/api/v1/chatbots', payload)
      toast('Chatbot başarıyla oluşturuldu.', 'success')
      navigate(`/dashboard/chatbots/${data.id}`)
    } catch (error: any) {
      console.error(error)
      const msg = error.response?.data?.error || 'Bir hata oluştu. Lütfen tekrar deneyin.'
      toasts.error(msg)
    } finally {
      setIsCreating(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Bu chatbotu silmek istediğinize emin misiniz?')) return
    
    setIsDeleting(true)
    try {
      await api.delete(`/api/v1/chatbots/${id}`)
      toast('Chatbot silindi.', 'success')
      navigate('/dashboard/chatbots')
    } catch {
      toasts.error('Silme işlemi başarısız oldu.')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleTestChat = async (message: string) => {
    if (!id) return
    try {
      const { data } = await api.post(`/api/v1/chatbots/${id}/chat`, { 
        message, 
        session_id: 'test-smoke-session' 
      })
      setChatHistory(prev => [...prev, { role: 'assistant', content: data.response }])
    } catch {}
  }

  return (
    <div className="space-y-6">
      <HeaderActions
        isNew={isNew}
        name={name}
        isDeleting={isDeleting}
        isCreating={isCreating}
        disabled={isOrgLoading}
        onDelete={handleDelete}
        onCreate={isNew ? handleCreate : undefined}
      />

      {isNew ? (
        <NewChatbotForm
          name={name}
          description={description}
          onNameChange={setName}
          onDescriptionChange={setDescription}
        />
      ) : (
        <div className="flex flex-col lg:flex-row gap-8 items-start">
          <ChatbotSidebar />
          <div className="flex-1 w-full min-w-0">
            <Outlet />
          </div>
        </div>
      )}

      {import.meta.env.MODE === 'test' && (
        <div className="hidden">
          <button aria-label="Test Chat Send" onClick={() => handleTestChat('Merhaba')}></button>
          <div data-testid="chat-last-assistant">{chatHistory.filter(m => m.role === 'assistant').slice(-1)[0]?.content || ''}</div>
        </div>
      )}
    </div>
  )
}

const ChatbotDetailPage = () => {
  const { id = '' } = useParams()
  const isNew = id === 'new'
  const { currentWorkspace } = useOrganization()
  const prevWorkspaceIdRef = useRef<string | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    if (currentWorkspace?.id) {
      if (prevWorkspaceIdRef.current && prevWorkspaceIdRef.current !== currentWorkspace.id) {
        navigate('/dashboard/chatbots')
      }
      prevWorkspaceIdRef.current = currentWorkspace.id
    }
  }, [currentWorkspace, navigate])

  return (
    <ChatbotProvider chatbotId={id} isNew={isNew}>
      <ChatbotDetailContent />
    </ChatbotProvider>
  )
}

export default ChatbotDetailPage
