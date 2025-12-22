import { useState, useRef, useEffect } from 'react'
import { useParams, useNavigate, Outlet, useLocation } from 'react-router-dom'
import { api } from '@/api/client'
import { useToast } from '@/components/ui/toast'
import HeaderActions from '@/features/chatbot/components/HeaderActions'
import { ChatbotTabBar } from '@/features/chatbot/components/ChatbotTabBar'
import NewChatbotForm from '@/features/chatbot/components/NewChatbotForm'
import { useOrganization } from '@/features/organization/context/OrganizationContext'
import { ChatbotProvider, useChatbotContext } from '@/features/chatbot/context/ChatbotContext'
import { useToastErrors } from '@/features/chatbot/hooks/useToastErrors'
import { getTurkishErrorMessage } from '@/lib/errorMessages'

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
    buildPayload,
    isLoading: isChatbotLoading
  } = useChatbotContext()

  const [isCreating, setIsCreating] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  // Test Chat State
  const [chatHistory, setChatHistory] = useState<{role: 'user' | 'assistant', content: string}[]>([])

  // Redirect for legacy query params and old tab routes
  useEffect(() => {
    try {
      const u = new URL(window.location.href)
      const tab = u.searchParams.get('tab')
      if (tab && !location.pathname.includes(tab)) {
         navigate(tab, { replace: true })
      }
      
      // Redirect old routes to new 8-tab structure
      const pathTab = location.pathname.split('/').pop()
      const redirectMap: Record<string, string> = {
        'overview': 'settings',
        'guardrails': 'security',
        'handoff': 'security',
        'intelligence': 'sources',
        'suggestions': 'design',
        'connect': 'deploy',
        'analytics': 'insights',
        'requests': 'insights',
      }
      if (pathTab && redirectMap[pathTab]) {
        navigate(redirectMap[pathTab], { replace: true })
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
      toasts.error(getTurkishErrorMessage(error, 'Bir hata oluştu. Lütfen tekrar deneyin.'))
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
    } catch (error: any) {
      toasts.error(getTurkishErrorMessage(error, 'Silme işlemi başarısız oldu.'))
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
    } catch {
      setChatHistory(prev => [...prev, { role: 'assistant', content: 'Bir hata oluştu.' }])
    }
  }

  if (isChatbotLoading && !isNew) {
    return (
      <div className="space-y-6 pb-20 lg:pb-0">
        <div className="flex justify-between items-center h-16">
           <div className="space-y-2">
             <div className="h-8 w-48 bg-muted animate-pulse rounded-md" />
             <div className="h-4 w-64 bg-muted animate-pulse rounded-md" />
           </div>
           <div className="flex gap-2">
             <div className="h-10 w-24 bg-muted animate-pulse rounded-md" />
             <div className="h-10 w-24 bg-muted animate-pulse rounded-md" />
           </div>
        </div>
        <div className="h-12 w-full bg-muted animate-pulse rounded-xl" />
        <div className="h-96 w-full bg-muted animate-pulse rounded-xl" />
      </div>
    )
  }

  return (
    <div className="space-y-4 lg:space-y-6 pb-20 lg:pb-0">
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
        <div className="space-y-4 lg:space-y-6">
          <ChatbotTabBar />
          <div className="w-full min-w-0">
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
