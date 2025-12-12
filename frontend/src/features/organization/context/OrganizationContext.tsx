import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from 'react'
import {
  	getOrganizations,
	Organization,
} from '@/api/organization'
import { getWorkspaces, Workspace } from '@/api/workspace'
import { useToast } from '@/components/ui/toast'

interface OrganizationContextType {
  organizations: Organization[]
  currentOrganization: Organization | null
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  isLoading: boolean
  selectOrganization: (orgId: string) => Promise<void>
  selectWorkspace: (workspaceId: string) => void
  refreshOrganizations: () => Promise<void>
  refreshWorkspaces: () => Promise<void>
}

const OrganizationContext = createContext<OrganizationContextType | undefined>(
  undefined,
)

export const OrganizationProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const { toast } = useToast()
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [currentOrganization, setCurrentOrganization] =
    useState<Organization | null>(null)
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [currentWorkspace, setCurrentWorkspace] = useState<Workspace | null>(
    null,
  )
  const [isLoading, setIsLoading] = useState(true)

  const loadOrganizations = useCallback(async () => {
    try {
      const orgs = (await getOrganizations()) || []
      setOrganizations(orgs)
      
      // Auto-select first org if none selected
      if (!currentOrganization && orgs.length > 0) {
        // Try to recover from local storage
        const savedOrgId = localStorage.getItem('botla_last_org_id')
        const targetOrg = orgs.find(o => o.id === savedOrgId) || orgs[0]
        await selectOrganization(targetOrg.id, orgs)
      } else if (currentOrganization) {
        // Refresh current org data
        const updatedCurrent = orgs.find(o => o.id === currentOrganization.id)
        if (updatedCurrent) setCurrentOrganization(updatedCurrent)
      }
    } catch (error) {
      console.error('Failed to load organizations', error)
      toast('Organizasyonlar yüklenemedi', 'error')
    } finally {
      setIsLoading(false)
    }
  }, [currentOrganization])

  const selectOrganization = async (orgId: string, orgsList = organizations) => {
    const org = orgsList.find((o) => o.id === orgId)
    if (!org) return

    setCurrentOrganization(org)
    localStorage.setItem('botla_last_org_id', orgId)
    
    // Reset workspace when switching orgs
    setCurrentWorkspace(null)
    setWorkspaces([])
    
    // Load workspaces for this org
    try {
      const ws = (await getWorkspaces(orgId)) || []
      setWorkspaces(ws)
      
      // Auto-select first workspace or recover from local storage
      const savedWsId = localStorage.getItem(`botla_last_ws_id_${orgId}`)
      const targetWs = ws.find(w => w.id === savedWsId) || ws[0]
      if (targetWs) {
        selectWorkspace(targetWs.id, ws, orgId)
      }
    } catch (error) {
      console.error('Failed to load workspaces', error)
      toast('Çalışma alanları yüklenemedi', 'error')
    }
  }

  const selectWorkspace = (workspaceId: string, wsList = workspaces, orgId?: string) => {
    const ws = wsList.find((w) => w.id === workspaceId)
    if (ws) {
      setCurrentWorkspace(ws)
      // Use orgId parameter if provided (during org switch), otherwise use currentOrganization
      const targetOrgId = orgId || currentOrganization?.id
      if (targetOrgId) {
        localStorage.setItem(`botla_last_ws_id_${targetOrgId}`, workspaceId)
      }
    }
  }

  const refreshWorkspaces = async () => {
    if (currentOrganization) {
      try {
        const ws = await getWorkspaces(currentOrganization.id)
        setWorkspaces(ws)
      } catch (error) {
        console.error(error)
      }
    }
  }

  useEffect(() => {
    loadOrganizations()
  }, [])

  return (
    <OrganizationContext.Provider
      value={{
        organizations,
        currentOrganization,
        workspaces,
        currentWorkspace,
        isLoading,
        selectOrganization: (id) => selectOrganization(id),
        selectWorkspace: (id) => selectWorkspace(id),
        refreshOrganizations: loadOrganizations,
        refreshWorkspaces,
      }}
    >
      {children}
    </OrganizationContext.Provider>
  )
}

export const useOrganization = () => {
  const context = useContext(OrganizationContext)
  if (context === undefined) {
    throw new Error(
      'useOrganization must be used within an OrganizationProvider',
    )
  }
  return context
}
