# Phase 3: Frontend Admin Dashboard

> **Estimated Time:** 5-7 days  
> **Priority:** High (Week 2-3)  
> **Depends On:** Phase 1, Phase 2  

This phase builds the React frontend for the admin dashboard.

---

## Step 3.1: Admin Route Protection

Set up admin-only routing.

### Tasks

- [x] **Update `frontend/src/types/user.ts`**
  
  Add admin flag to User type:
  ```typescript
  export interface User {
    // ... existing fields
    is_platform_admin: boolean;
  }
  ```

- [x] **Create `frontend/src/features/admin/AdminRoute.tsx`**
  
  ```typescript
  import { Navigate } from 'react-router-dom';
  import { useAuth } from '@/hooks/useAuth';
  
  interface AdminRouteProps {
    children: React.ReactNode;
  }
  
  export function AdminRoute({ children }: AdminRouteProps) {
    const { user, isLoading } = useAuth();
    
    if (isLoading) {
      return <div className="flex items-center justify-center h-screen">Loading...</div>;
    }
    
    if (!user?.is_platform_admin) {
      return <Navigate to="/dashboard" replace />;
    }
    
    return <>{children}</>;
  }
  ```

- [x] **Register admin routes** in `frontend/src/App.tsx`
  
  ```typescript
  import { AdminRoute } from '@/features/admin/AdminRoute';
  import { AdminLayout } from '@/pages/admin/AdminLayout';
  import { AdminDashboardPage } from '@/pages/admin/AdminDashboardPage';
  // ... other imports
  
  // Inside route config:
  <Route path="/admin" element={
    <AdminRoute>
      <AdminLayout />
    </AdminRoute>
  }>
    <Route index element={<AdminDashboardPage />} />
    <Route path="users" element={<AdminUsersPage />} />
    <Route path="users/:id" element={<AdminUserDetailPage />} />
    <Route path="organizations" element={<AdminOrganizationsPage />} />
    <Route path="chatbots" element={<AdminChatbotsPage />} />
    <Route path="sources" element={<AdminSourcesPage />} />
    <Route path="system" element={<AdminSystemPage />} />
    <Route path="queues" element={<AdminQueuesPage />} />
    <Route path="errors" element={<AdminErrorsPage />} />
    <Route path="audit" element={<AdminAuditPage />} />
  </Route>
  ```

---

## Step 3.2: Admin API Client

Create typed API functions for admin endpoints.

### Tasks

- [x] **Create `frontend/src/api/admin.ts`**
  
  ```typescript
  import { api } from './client';
  
  // Types
  export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    per_page: number;
  }
  
  export interface OverviewStats {
    total_users: number;
    total_organizations: number;
    total_chatbots: number;
    total_conversations: number;
    users_today: number;
    conversations_today: number;
    active_plans: Record<string, number>;
  }
  
  export interface DependencyStatus {
    name: string;
    status: 'ok' | 'degraded' | 'down';
    latency_ms: number;
    message?: string;
  }
  
  export interface DetailedHealth {
    status: 'healthy' | 'degraded' | 'unhealthy';
    version: string;
    uptime: string;
    dependencies: DependencyStatus[];
    environment: string;
  }
  
  export interface QueueStats {
    queue_name: string;
    pending_count: number;
    processing_count: number;
    failed_count: number;
    oldest_pending?: string;
  }
  
  export interface StuckJob {
    id: string;
    queue_name: string;
    source_id?: string;
    status: string;
    started_at: string;
    stuck_duration: string;
    error_message?: string;
  }
  
  export interface ErrorLogEntry {
    id: string;
    error_type: string;
    message: string;
    severity: string;
    created_at: string;
    stack_trace?: string;
  }
  
  export interface AdminUser {
    id: string;
    email: string;
    full_name: string;
    plan_id: string;
    created_at: string;
    is_suspended: boolean;
  }
  
  // API Functions
  export const adminApi = {
    // Stats
    getOverviewStats: () => 
      api.get<OverviewStats>('/admin/stats/overview'),
    
    // Health
    getDetailedHealth: () => 
      api.get<DetailedHealth>('/admin/health/detailed'),
    
    // Users
    listUsers: (params: { page?: number; search?: string; status?: string }) =>
      api.get<PaginatedResponse<AdminUser>>('/admin/users', { params }),
    
    getUser: (id: string) =>
      api.get<AdminUser>(`/admin/users/${id}`),
    
    updateUser: (id: string, data: { status?: string }) =>
      api.patch(`/admin/users/${id}`, data),
    
    // Queues
    getQueues: () =>
      api.get<QueueStats[]>('/admin/queues'),
    
    getStuckJobs: (threshold?: string) =>
      api.get<StuckJob[]>('/admin/queues/stuck', { params: { threshold } }),
    
    retryJob: (id: string) =>
      api.post(`/admin/queues/${id}/retry`),
    
    deleteJob: (id: string) =>
      api.delete(`/admin/queues/${id}`),
    
    // Errors
    listErrors: (params: { page?: number; severity?: string; type?: string }) =>
      api.get<PaginatedResponse<ErrorLogEntry>>('/admin/errors', { params }),
    
    getError: (id: string) =>
      api.get<ErrorLogEntry>(`/admin/errors/${id}`),
    
    // Organizations
    listOrganizations: (params: { page?: number; search?: string }) =>
      api.get<PaginatedResponse<any>>('/admin/organizations', { params }),
    
    // Chatbots
    listChatbots: (params: { page?: number; search?: string; status?: string }) =>
      api.get<PaginatedResponse<any>>('/admin/chatbots', { params }),
    
    forceRefreshChatbot: (id: string) =>
      api.post(`/admin/chatbots/${id}/force-refresh`),
    
    // Audit
    listAuditLogs: (params: { page?: number }) =>
      api.get<PaginatedResponse<any>>('/admin/audit-logs', { params }),
  };
  ```

---

## Step 3.3: Admin Layout

Create the admin shell with sidebar navigation.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminLayout.tsx`**
  
  ```typescript
  import { Outlet, NavLink } from 'react-router-dom';
  import { 
    LayoutDashboard, Users, Building2, Bot, Database, 
    AlertTriangle, Activity, Clock, FileText 
  } from 'lucide-react';
  
  const navItems = [
    { to: '/admin', icon: LayoutDashboard, label: 'Overview', end: true },
    { to: '/admin/users', icon: Users, label: 'Users' },
    { to: '/admin/organizations', icon: Building2, label: 'Organizations' },
    { to: '/admin/chatbots', icon: Bot, label: 'Chatbots' },
    { to: '/admin/sources', icon: Database, label: 'Sources' },
    { to: '/admin/system', icon: Activity, label: 'System Health' },
    { to: '/admin/queues', icon: Clock, label: 'Queues' },
    { to: '/admin/errors', icon: AlertTriangle, label: 'Errors' },
    { to: '/admin/audit', icon: FileText, label: 'Audit Log' },
  ];
  
  export function AdminLayout() {
    return (
      <div className="flex h-screen bg-gray-100 dark:bg-gray-900">
        {/* Sidebar */}
        <aside className="w-64 bg-white dark:bg-gray-800 border-r">
          <div className="p-4 border-b">
            <h1 className="text-xl font-bold text-red-600">Admin Dashboard</h1>
          </div>
          <nav className="p-4 space-y-1">
            {navItems.map(item => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                    isActive 
                      ? 'bg-primary text-primary-foreground' 
                      : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`
                }
              >
                <item.icon className="w-5 h-5" />
                {item.label}
              </NavLink>
            ))}
          </nav>
        </aside>
        
        {/* Main content */}
        <main className="flex-1 overflow-auto">
          <div className="p-6">
            <Outlet />
          </div>
        </main>
      </div>
    );
  }
  ```

---

## Step 3.4: Admin Dashboard Overview Page

Main dashboard with stats and status panels.

### Tasks

- [x] **Create `frontend/src/features/admin/components/StatsCard.tsx`**
  
  ```typescript
  interface StatsCardProps {
    title: string;
    value: number | string;
    subtitle?: string;
    icon: React.ReactNode;
    trend?: { value: number; isPositive: boolean };
  }
  
  export function StatsCard({ title, value, subtitle, icon, trend }: StatsCardProps) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500">{title}</p>
            <p className="text-3xl font-bold mt-1">{value}</p>
            {subtitle && <p className="text-sm text-gray-400 mt-1">{subtitle}</p>}
          </div>
          <div className="p-3 bg-primary/10 rounded-full">
            {icon}
          </div>
        </div>
      </div>
    );
  }
  ```

- [x] **Create `frontend/src/features/admin/components/HealthPanel.tsx`**
  
  ```typescript
  import { useQuery } from '@tanstack/react-query';
  import { adminApi, DependencyStatus } from '@/api/admin';
  
  function StatusBadge({ status }: { status: string }) {
    const colors = {
      ok: 'bg-green-100 text-green-800',
      degraded: 'bg-yellow-100 text-yellow-800',
      down: 'bg-red-100 text-red-800',
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs font-medium ${colors[status] || colors.down}`}>
        {status.toUpperCase()}
      </span>
    );
  }
  
  export function HealthPanel() {
    const { data, isLoading, refetch } = useQuery({
      queryKey: ['admin', 'health'],
      queryFn: () => adminApi.getDetailedHealth(),
      refetchInterval: 30000, // Refresh every 30s
    });
    
    if (isLoading) return <div>Loading...</div>;
    
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold">System Health</h3>
          <StatusBadge status={data?.data.status || 'unknown'} />
        </div>
        <div className="space-y-3">
          {data?.data.dependencies.map(dep => (
            <div key={dep.name} className="flex items-center justify-between">
              <span className="capitalize">{dep.name}</span>
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-500">{dep.latency_ms}ms</span>
                <StatusBadge status={dep.status} />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }
  ```

- [x] **Create `frontend/src/pages/admin/AdminDashboardPage.tsx`**
  
  ```typescript
  import { useQuery } from '@tanstack/react-query';
  import { Users, Building2, Bot, MessageSquare } from 'lucide-react';
  import { adminApi } from '@/api/admin';
  import { StatsCard } from '@/features/admin/components/StatsCard';
  import { HealthPanel } from '@/features/admin/components/HealthPanel';
  
  export function AdminDashboardPage() {
    const { data: stats } = useQuery({
      queryKey: ['admin', 'stats'],
      queryFn: () => adminApi.getOverviewStats(),
    });
    
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Dashboard Overview</h1>
        
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatsCard
            title="Total Users"
            value={stats?.data.total_users || 0}
            subtitle={`+${stats?.data.users_today || 0} today`}
            icon={<Users className="w-6 h-6 text-primary" />}
          />
          <StatsCard
            title="Organizations"
            value={stats?.data.total_organizations || 0}
            icon={<Building2 className="w-6 h-6 text-primary" />}
          />
          <StatsCard
            title="Chatbots"
            value={stats?.data.total_chatbots || 0}
            icon={<Bot className="w-6 h-6 text-primary" />}
          />
          <StatsCard
            title="Conversations"
            value={stats?.data.total_conversations || 0}
            subtitle={`+${stats?.data.conversations_today || 0} today`}
            icon={<MessageSquare className="w-6 h-6 text-primary" />}
          />
        </div>
        
        {/* Panels Row */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <HealthPanel />
          {/* Add more panels: RecentErrors, QueueStatus */}
        </div>
      </div>
    );
  }
  ```

---

## Step 3.5: System Health Page

Dedicated page for monitoring all dependencies.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminSystemPage.tsx`**
  
  ```typescript
  import { useQuery } from '@tanstack/react-query';
  import { adminApi } from '@/api/admin';
  import { RefreshCw } from 'lucide-react';
  import { Button } from '@/components/ui/Button';
  
  export function AdminSystemPage() {
    const { data, isLoading, refetch, isFetching } = useQuery({
      queryKey: ['admin', 'health'],
      queryFn: () => adminApi.getDetailedHealth(),
    });
    
    const health = data?.data;
    
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">System Health</h1>
          <Button onClick={() => refetch()} disabled={isFetching}>
            <RefreshCw className={`w-4 h-4 mr-2 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
        
        {/* Overall Status */}
        <div className="bg-white dark:bg-gray-800 rounded-lg p-6">
          <div className="flex items-center gap-4">
            <div className={`w-4 h-4 rounded-full ${
              health?.status === 'healthy' ? 'bg-green-500' :
              health?.status === 'degraded' ? 'bg-yellow-500' : 'bg-red-500'
            }`} />
            <span className="text-lg font-semibold capitalize">
              {health?.status || 'Unknown'}
            </span>
          </div>
          <div className="mt-4 grid grid-cols-3 gap-4 text-sm text-gray-500">
            <div>Version: {health?.version}</div>
            <div>Uptime: {health?.uptime}</div>
            <div>Environment: {health?.environment}</div>
          </div>
        </div>
        
        {/* Dependencies Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {health?.dependencies.map(dep => (
            <div key={dep.name} className="bg-white dark:bg-gray-800 rounded-lg p-6">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold capitalize">{dep.name}</h3>
                <span className={`px-2 py-1 rounded text-xs ${
                  dep.status === 'ok' ? 'bg-green-100 text-green-800' :
                  dep.status === 'degraded' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-red-100 text-red-800'
                }`}>
                  {dep.status.toUpperCase()}
                </span>
              </div>
              <p className="text-sm text-gray-500">Latency: {dep.latency_ms}ms</p>
              {dep.message && (
                <p className="text-sm text-red-500 mt-2">{dep.message}</p>
              )}
            </div>
          ))}
        </div>
      </div>
    );
  }
  ```

---

## Step 3.6: Queues Page

Monitor and manage processing queues.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminQueuesPage.tsx`**
  
  ```typescript
  import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
  import { adminApi } from '@/api/admin';
  import { Button } from '@/components/ui/Button';
  import { RefreshCw, Trash2 } from 'lucide-react';
  
  export function AdminQueuesPage() {
    const queryClient = useQueryClient();
    
    const { data: queues } = useQuery({
      queryKey: ['admin', 'queues'],
      queryFn: () => adminApi.getQueues(),
      refetchInterval: 10000,
    });
    
    const { data: stuckJobs } = useQuery({
      queryKey: ['admin', 'queues', 'stuck'],
      queryFn: () => adminApi.getStuckJobs(),
      refetchInterval: 10000,
    });
    
    const retryMutation = useMutation({
      mutationFn: (id: string) => adminApi.retryJob(id),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'queues'] }),
    });
    
    const deleteMutation = useMutation({
      mutationFn: (id: string) => adminApi.deleteJob(id),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'queues'] }),
    });
    
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Queue Management</h1>
        
        {/* Queue Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {queues?.data.map(q => (
            <div key={q.queue_name} className="bg-white dark:bg-gray-800 rounded-lg p-6">
              <h3 className="font-semibold capitalize">{q.queue_name.replace('_', ' ')}</h3>
              <div className="mt-4 space-y-2 text-sm">
                <div className="flex justify-between">
                  <span>Pending</span>
                  <span className="font-medium">{q.pending_count}</span>
                </div>
                <div className="flex justify-between">
                  <span>Processing</span>
                  <span className="font-medium">{q.processing_count}</span>
                </div>
                <div className="flex justify-between">
                  <span>Failed</span>
                  <span className="font-medium text-red-500">{q.failed_count}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
        
        {/* Stuck Jobs */}
        <div className="bg-white dark:bg-gray-800 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">
            Stuck Jobs ({stuckJobs?.data.length || 0})
          </h2>
          {stuckJobs?.data.length === 0 ? (
            <p className="text-gray-500">No stuck jobs</p>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="text-left text-sm text-gray-500">
                  <th className="pb-2">Queue</th>
                  <th className="pb-2">Status</th>
                  <th className="pb-2">Stuck Duration</th>
                  <th className="pb-2">Error</th>
                  <th className="pb-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {stuckJobs?.data.map(job => (
                  <tr key={job.id} className="border-t">
                    <td className="py-3">{job.queue_name}</td>
                    <td className="py-3">{job.status}</td>
                    <td className="py-3">{job.stuck_duration}</td>
                    <td className="py-3 text-sm text-red-500 max-w-xs truncate">
                      {job.error_message}
                    </td>
                    <td className="py-3">
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => retryMutation.mutate(job.id)}
                          disabled={retryMutation.isPending}
                        >
                          <RefreshCw className="w-4 h-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => deleteMutation.mutate(job.id)}
                          disabled={deleteMutation.isPending}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    );
  }
  ```

---

## Step 3.7: Errors Page

View and filter error logs.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminErrorsPage.tsx`**
  
  ```typescript
  import { useState } from 'react';
  import { useQuery } from '@tanstack/react-query';
  import { adminApi } from '@/api/admin';
  import { formatDistanceToNow } from 'date-fns';
  
  export function AdminErrorsPage() {
    const [severity, setSeverity] = useState<string>('');
    const [page, setPage] = useState(1);
    
    const { data, isLoading } = useQuery({
      queryKey: ['admin', 'errors', { severity, page }],
      queryFn: () => adminApi.listErrors({ severity, page }),
    });
    
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Error Logs</h1>
          <select
            value={severity}
            onChange={e => setSeverity(e.target.value)}
            className="px-3 py-2 border rounded-lg"
          >
            <option value="">All Severities</option>
            <option value="critical">Critical</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
          </select>
        </div>
        
        <div className="bg-white dark:bg-gray-800 rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr className="text-left text-sm">
                <th className="px-4 py-3">Severity</th>
                <th className="px-4 py-3">Type</th>
                <th className="px-4 py-3">Message</th>
                <th className="px-4 py-3">Time</th>
              </tr>
            </thead>
            <tbody>
              {data?.data.data.map(error => (
                <tr key={error.id} className="border-t hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer">
                  <td className="px-4 py-3">
                    <span className={`px-2 py-1 rounded text-xs ${
                      error.severity === 'critical' ? 'bg-red-100 text-red-800' :
                      error.severity === 'error' ? 'bg-orange-100 text-orange-800' :
                      'bg-yellow-100 text-yellow-800'
                    }`}>
                      {error.severity}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm">{error.error_type}</td>
                  <td className="px-4 py-3 text-sm max-w-md truncate">{error.message}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {formatDistanceToNow(new Date(error.created_at), { addSuffix: true })}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        
        {/* Pagination */}
        <div className="flex justify-center gap-2">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-4 py-2 border rounded disabled:opacity-50"
          >
            Previous
          </button>
          <span className="px-4 py-2">Page {page}</span>
          <button
            onClick={() => setPage(p => p + 1)}
            disabled={!data?.data.data.length}
            className="px-4 py-2 border rounded disabled:opacity-50"
          >
            Next
          </button>
        </div>
      </div>
    );
  }
  ```

---

## Step 3.8: Users Page

List and manage users.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminUsersPage.tsx`**
  
  ```typescript
  import { useState } from 'react';
  import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
  import { adminApi } from '@/api/admin';
  import { Link } from 'react-router-dom';
  import { MoreHorizontal, UserX, UserCheck, Download, LogIn } from 'lucide-react';
  import { DropdownMenu } from '@/components/ui/DropdownMenu';
  
  export function AdminUsersPage() {
    const [search, setSearch] = useState('');
    const [status, setStatus] = useState('');
    const [page, setPage] = useState(1);
    const queryClient = useQueryClient();
    
    const { data, isLoading } = useQuery({
      queryKey: ['admin', 'users', { search, status, page }],
      queryFn: () => adminApi.listUsers({ search, status, page }),
    });
    
    const suspendMutation = useMutation({
      mutationFn: (id: string) => adminApi.updateUser(id, { status: 'suspended' }),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'users'] }),
    });
    
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Users</h1>
        
        {/* Filters */}
        <div className="flex gap-4">
          <input
            type="text"
            placeholder="Search by email..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="flex-1 px-4 py-2 border rounded-lg"
          />
          <select
            value={status}
            onChange={e => setStatus(e.target.value)}
            className="px-4 py-2 border rounded-lg"
          >
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>
        
        {/* Users Table */}
        <div className="bg-white dark:bg-gray-800 rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr className="text-left text-sm">
                <th className="px-4 py-3">Email</th>
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3">Plan</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Created</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {data?.data.data.map(user => (
                <tr key={user.id} className="border-t">
                  <td className="px-4 py-3">
                    <Link to={`/admin/users/${user.id}`} className="text-primary hover:underline">
                      {user.email}
                    </Link>
                  </td>
                  <td className="px-4 py-3">{user.full_name}</td>
                  <td className="px-4 py-3">{user.plan_id}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-1 rounded text-xs ${
                      user.is_suspended 
                        ? 'bg-red-100 text-red-800' 
                        : 'bg-green-100 text-green-800'
                    }`}>
                      {user.is_suspended ? 'Suspended' : 'Active'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {new Date(user.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    {/* Actions dropdown menu */}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }
  ```

---

## Step 3.9: Additional Pages

Create remaining pages with basic structure.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminOrganizationsPage.tsx`**
  - List organizations with member count, chatbot count
  - Filter by plan, search by name
  - Link to detail page

- [x] **Create `frontend/src/pages/admin/AdminChatbotsPage.tsx`**
  - List all chatbots across platform
  - Filter by status, organization
  - Force refresh action

- [x] **Create `frontend/src/pages/admin/AdminSourcesPage.tsx`**
  - List all data sources
  - Filter by status (processing, ready, failed)
  - Reprocess failed sources

- [x] **Create `frontend/src/pages/admin/AdminAuditPage.tsx`**
  - List admin actions
  - Filter by action type, admin user

---

## Verification

### Manual Testing

1. **Login as admin user**
2. **Navigate to `/admin`**
   - Should show dashboard with stats
   - Should show health panel
3. **Test navigation**
   - All sidebar links work
   - Routes load correctly
4. **Test system health page**
   - Refresh button works
   - Dependencies show status
5. **Test queues page**
   - Queue stats display
   - Stuck jobs table (if any)
   - Retry/delete actions work
6. **Test errors page**
   - Errors list with pagination
   - Severity filter works

### Non-admin Access Test

1. Login as regular user (non-admin)
2. Navigate directly to `/admin`
3. Should redirect to `/dashboard`

---

## Files to Create

| File | Description |
|------|-------------|
| `frontend/src/features/admin/AdminRoute.tsx` | Route protection |
| `frontend/src/api/admin.ts` | Admin API client |
| `frontend/src/pages/admin/AdminLayout.tsx` | Admin shell layout |
| `frontend/src/pages/admin/AdminDashboardPage.tsx` | Overview dashboard |
| `frontend/src/pages/admin/AdminSystemPage.tsx` | Health monitoring |
| `frontend/src/pages/admin/AdminQueuesPage.tsx` | Queue management |
| `frontend/src/pages/admin/AdminErrorsPage.tsx` | Error logs |
| `frontend/src/pages/admin/AdminUsersPage.tsx` | User management |
| `frontend/src/pages/admin/AdminOrganizationsPage.tsx` | Org management |
| `frontend/src/pages/admin/AdminChatbotsPage.tsx` | Chatbot management |
| `frontend/src/pages/admin/AdminSourcesPage.tsx` | Source management |
| `frontend/src/pages/admin/AdminAuditPage.tsx` | Audit log |
| `frontend/src/features/admin/components/StatsCard.tsx` | Stats display |
| `frontend/src/features/admin/components/HealthPanel.tsx` | Health panel |
