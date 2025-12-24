/**
 * User type representing the authenticated user's profile data
 * from the /api/v1/me endpoint
 */
export interface User {
  id: string
  email: string
  created_at: string
  full_name?: string
  avatar_url?: string
  is_platform_admin: boolean
  organizations?: UserOrganization[]
}

/**
 * Organization membership as returned in the user profile
 */
export interface UserOrganization {
  id: string
  name: string
  role: 'owner' | 'admin' | 'member'
}
