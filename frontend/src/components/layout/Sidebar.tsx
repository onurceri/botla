import { FC } from 'react'
import { Link, NavLink } from 'react-router-dom'

const Sidebar: FC = () => {
  return (
    <aside className="flex w-64 flex-col border-r bg-sidebar p-4 text-sidebar-foreground">
      <Link to="/" className="mb-6 text-xl font-bold">
        Botla
      </Link>
      <nav className="flex flex-1 flex-col gap-2">
        <NavLink to="/dashboard" className={({ isActive }) => `rounded px-3 py-2 ${isActive ? 'bg-sidebar-primary text-sidebar-primary-foreground' : 'hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'}`}>Dashboard</NavLink>
        <NavLink to="/chatbots" className={({ isActive }) => `rounded px-3 py-2 ${isActive ? 'bg-sidebar-primary text-sidebar-primary-foreground' : 'hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'}`}>Chatbotlar</NavLink>
        <NavLink to="/analytics" className={({ isActive }) => `rounded px-3 py-2 ${isActive ? 'bg-sidebar-primary text-sidebar-primary-foreground' : 'hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'}`}>Analitik</NavLink>
        <NavLink to="/settings" className={({ isActive }) => `rounded px-3 py-2 ${isActive ? 'bg-sidebar-primary text-sidebar-primary-foreground' : 'hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'}`}>Ayarlar</NavLink>
      </nav>
    </aside>
  )
}

export default Sidebar
