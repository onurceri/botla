import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { CustomTooltip } from '../DashboardPage'

describe('Dashboard CustomTooltip', () => {
  it('renders formatted date and payload items when active', () => {
    const date = new Date(2024, 4, 10).toISOString()
    const payload = [
      { name: 'Konuşma', value: 3, color: '#8b5cf6' },
      { name: 'Mesaj', value: 5, color: '#f59e0b' },
    ]
    render(<CustomTooltip active={true} payload={payload} label={date} />)
    expect(screen.getByText(/10 Mayıs 2024/)).toBeInTheDocument()
    expect(screen.getByText('Konuşma:')).toBeInTheDocument()
    expect(screen.getByText('Mesaj:')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('returns null when not active', () => {
    const { container } = render(<CustomTooltip active={false} payload={[]} label={''} />)
    expect(container.firstChild).toBeNull()
  })
})

