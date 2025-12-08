import React, { useState } from 'react'
import { Plus, Layers, Settings } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
} from '@/components/ui/select'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useOrganization } from '../context/OrganizationContext'
import { CreateOrganizationDialog } from './CreateOrganizationDialog'
import { CreateWorkspaceDialog } from './CreateWorkspaceDialog'
import { cn } from '@/lib/utils'

function getInitials(name: string) {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

export const OrganizationSwitcher: React.FC = () => {
  const {
    organizations,
    currentOrganization,
    workspaces,
    currentWorkspace,
    selectOrganization,
    selectWorkspace,
    isLoading,
  } = useOrganization()

  const navigate = useNavigate()
  const [createOrgOpen, setCreateOrgOpen] = useState(false)
  const [createWsOpen, setCreateWsOpen] = useState(false)

  if (isLoading) {
    return (
      <div className="flex gap-2">
        <div className="h-10 w-10 md:w-[200px] animate-pulse rounded-md bg-muted" />
        <div className="h-10 w-10 md:w-[200px] animate-pulse rounded-md bg-muted" />
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      {/* Organization Select */}
      <Select
        value={currentOrganization?.id || ''}
        onValueChange={(val) => {
          if (val === 'create-new') {
            setCreateOrgOpen(true)
            return
          }
          if (val === 'settings') {
            navigate('/settings/organization')
            return
          }
          selectOrganization(val)
        }}
      >
        <SelectTrigger 
          className={cn(
            "w-[50px] md:w-[260px] px-2 md:px-3 transition-all duration-200",
            !currentOrganization && "text-muted-foreground"
          )}
        >
          {currentOrganization ? (
            <div className="flex items-center gap-2 overflow-hidden">
              <Avatar className="h-6 w-6 border border-border shrink-0">
                <AvatarImage src={`https://avatar.vercel.sh/${currentOrganization.id}.png?text=${getInitials(currentOrganization.name)}`} />
                <AvatarFallback>{getInitials(currentOrganization.name)}</AvatarFallback>
              </Avatar>
              <span className="hidden md:block truncate text-sm font-medium">
                {currentOrganization.name}
              </span>
            </div>
          ) : (
            <span className="hidden md:block">Organizasyon Seç</span>
          )}
        </SelectTrigger>
        <SelectContent align="end" className="w-[280px]">
          <SelectGroup>
            <SelectLabel className="text-xs text-muted-foreground px-2 py-1.5 font-normal">
              Organizasyonlar
            </SelectLabel>
            {organizations?.map((org) => (
              <SelectItem key={org.id} value={org.id} className="cursor-pointer">
                <div className="flex items-center gap-2">
                  <Avatar className="h-5 w-5 border border-border">
                    <AvatarFallback className="text-[10px]">{getInitials(org.name)}</AvatarFallback>
                  </Avatar>
                  <span className="truncate">{org.name}</span>
                </div>
              </SelectItem>
            ))}
            <SelectSeparator />
            <SelectItem value="create-new" className="cursor-pointer text-muted-foreground focus:text-primary">
              <div className="flex items-center gap-2">
                <Plus className="h-4 w-4" />
                <span>Yeni Organizasyon</span>
              </div>
            </SelectItem>
            <SelectItem value="settings" className="cursor-pointer text-muted-foreground focus:text-primary">
              <div className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                <span>Ayarlar</span>
              </div>
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>

      {/* Workspace Select */}
      {currentOrganization && (
        <Select
          value={currentWorkspace?.id || 'none'}
          onValueChange={(val) => {
            if (val === 'create-new') {
              setCreateWsOpen(true)
            } else if (val === 'settings') {
              navigate('/settings/workspace')
            } else if (val !== 'none') {
              selectWorkspace(val)
            }
          }}
        >
          <SelectTrigger 
            className={cn(
              "w-[50px] md:w-[260px] px-2 md:px-3 transition-all duration-200",
              (!currentWorkspace || currentWorkspace.id === 'none') && "text-muted-foreground"
            )}
          >
            {currentWorkspace ? (
              <div className="flex items-center gap-2 overflow-hidden">
                <div className="h-6 w-6 rounded-md bg-primary/10 flex items-center justify-center border border-primary/20 shrink-0">
                  <Layers className="h-3.5 w-3.5 text-primary" />
                </div>
                <span className="hidden md:block truncate text-sm">
                  {currentWorkspace.name}
                </span>
              </div>
            ) : (
              <span className="hidden md:block">Çalışma Alanı</span>
            )}
          </SelectTrigger>
          <SelectContent align="end" className="w-[280px]">
            <SelectGroup>
              <SelectLabel className="text-xs text-muted-foreground px-2 py-1.5 font-normal">
                {currentOrganization.name} / Çalışma Alanları
              </SelectLabel>
              {workspaces?.map((ws) => (
                <SelectItem key={ws.id} value={ws.id} className="cursor-pointer">
                  <div className="flex items-center gap-2">
                    <div className="h-5 w-5 rounded bg-muted flex items-center justify-center shrink-0">
                       <span className="text-[10px] font-bold text-muted-foreground">
                         {ws.name.substring(0,1).toUpperCase()}
                       </span>
                    </div>
                    <span className="truncate">{ws.name}</span>
                  </div>
                </SelectItem>
              ))}
              {workspaces?.length === 0 && (
                <div className="px-2 py-2 text-xs text-muted-foreground text-center">
                  Henüz çalışma alanı yok
                </div>
              )}
              <SelectSeparator />
              <SelectItem value="create-new" className="cursor-pointer text-muted-foreground focus:text-primary">
                <div className="flex items-center gap-2">
                  <Plus className="h-4 w-4" />
                  <span>Yeni Çalışma Alanı</span>
                </div>
              </SelectItem>
              {currentWorkspace && (
                <SelectItem value="settings" className="cursor-pointer text-muted-foreground focus:text-primary">
                  <div className="flex items-center gap-2">
                    <Settings className="h-4 w-4" />
                    <span>Ayarlar</span>
                  </div>
                </SelectItem>
              )}
            </SelectGroup>
          </SelectContent>
        </Select>
      )}

      <CreateOrganizationDialog
        open={createOrgOpen}
        onOpenChange={setCreateOrgOpen}
      />
      
      <CreateWorkspaceDialog
        open={createWsOpen}
        onOpenChange={setCreateWsOpen}
      />
    </div>
  )
}
