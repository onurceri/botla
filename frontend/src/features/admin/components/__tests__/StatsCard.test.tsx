import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatsCard } from '../StatsCard'
import { Users } from 'lucide-react'

describe('StatsCard', () => {
  it('renders title and value', () => {
    render(<StatsCard title="Total Users" value="1,234" icon={<Users />} />)
    
    expect(screen.getByText('Total Users')).toBeInTheDocument()
    expect(screen.getByText('1,234')).toBeInTheDocument()
  })

  it('renders subtitle when provided', () => {
    render(<StatsCard title="Total Users" value="1,234" subtitle="+10 today" icon={<Users />} />)
    
    expect(screen.getByText('+10 today')).toBeInTheDocument()
  })

  it('renders trend when provided', () => {
    render(
      <StatsCard 
        title="Total Users" 
        value="1,234" 
        trend={{ value: 12, isPositive: true }} 
        icon={<Users />} 
      />
    )
    
    expect(screen.getByText('+12%')).toBeInTheDocument()
    expect(screen.getByText('+12%')).toHaveClass('text-green-500')
  })

  it('renders negative trend correctly', () => {
    render(
      <StatsCard 
        title="Total Users" 
        value="1,234" 
        trend={{ value: 5, isPositive: false }} 
        icon={<Users />} 
      />
    )
    
    expect(screen.getByText('-5%')).toBeInTheDocument()
    expect(screen.getByText('-5%')).toHaveClass('text-red-500')
  })

  it('renders loading skeleton when isLoading is true', () => {
    render(
      <StatsCard title="Total Users" value="1,234" icon={<Users />} isLoading={true} />
    )
    
    expect(screen.getByTestId('stats-card-skeleton')).toBeInTheDocument()
    expect(screen.queryByText('Total Users')).not.toBeInTheDocument()
  })
})
