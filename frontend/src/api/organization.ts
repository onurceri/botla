import { api } from './client'

export interface Organization {
    id: string
    name: string
    slug: string
    owner_id: string
    plan_id: string
    branding?: {
        primary_color?: string
        logo_url?: string
        remove_watermark?: boolean
    }
    created_at: string
    updated_at: string
}


export interface Member {
    id: string
    organization_id: string
    user_id: string
    role: 'owner' | 'admin' | 'member'
    created_at: string
    user: {
        id: string
        email: string
        full_name: string
        avatar_url: string
    }
}

export const getOrganizations = async (): Promise<Organization[]> => {
    const { data } = await api.get('/api/v1/organizations')
    return data
}

export const createOrganization = async (
    name: string,
    slug: string,
): Promise<Organization> => {
    const { data } = await api.post('/api/v1/organizations', { name, slug })
    return data
}

export const updateOrganization = async (id: string, name: string, slug: string): Promise<void> => {
    await api.patch(`/api/v1/organizations/${id}`, { name, slug })
}

export const deleteOrganization = async (id: string): Promise<void> => {
    await api.delete(`/api/v1/organizations/${id}`)
}

export interface MembersResponse {
    members: Member[]
    caller_role: 'owner' | 'admin' | 'member'
}

export const getMembers = async (orgId: string): Promise<MembersResponse> => {
    const { data } = await api.get(`/api/v1/organizations/${orgId}/members`)
    return data
}

export const addMember = async (orgId: string, email: string, role: string): Promise<void> => {
    await api.post(`/api/v1/organizations/${orgId}/members`, { email, role })
}

export const removeMember = async (orgId: string, userId: string): Promise<void> => {
    await api.delete(`/api/v1/organizations/${orgId}/members/${userId}`)
}

export const updateMemberRole = async (orgId: string, userId: string, role: string): Promise<void> => {
    await api.patch(`/api/v1/organizations/${orgId}/members/${userId}`, { role })
}
