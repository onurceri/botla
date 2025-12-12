import { api } from './client'

export interface Workspace {
    id: string
    organization_id: string
    name: string
    slug: string
    client_name?: string
    created_at: string
}

export const getWorkspaces = async (orgId: string): Promise<Workspace[]> => {
    const { data } = await api.get(`/api/v1/organizations/${orgId}/workspaces`)
    return data
}

export const createWorkspace = async (
    orgId: string,
    name: string,
    slug: string,
    clientName?: string,
): Promise<Workspace> => {
    const { data } = await api.post(`/api/v1/organizations/${orgId}/workspaces`, {
        name,
        slug,
        client_name: clientName,
    })
    return data
}

export const updateWorkspace = async (
    orgId: string,
    wsId: string,
    name: string,
    slug: string,
    clientName?: string,
): Promise<void> => {
    await api.patch(`/api/v1/organizations/${orgId}/workspaces/${wsId}`, {
        name,
        slug,
        client_name: clientName,
    })
}

export const deleteWorkspace = async (orgId: string, wsId: string): Promise<void> => {
    await api.delete(`/api/v1/organizations/${orgId}/workspaces/${wsId}`)
}
