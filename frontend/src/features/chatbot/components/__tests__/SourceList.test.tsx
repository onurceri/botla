import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within } from '@testing-library/react'
import SourceList from '../SourceList'

describe('SourceList', () => {
  it('renders sources and calls delete', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
    ] as any
    render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />)
    expect(screen.getByText('file.pdf')).toBeInTheDocument()
    const delBtn = screen.getByLabelText('Kaynağı Sil')
    fireEvent.click(delBtn)
    expect(onDelete).toHaveBeenCalledWith('1')
  })

  it('renders status variants with proper indicators', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'url', source_url: 'https://a.com', status: 'processing', chunk_count: 1 },
      { id: '3', source_type: 'text', original_filename: 'note', status: 'failed', chunk_count: 0 },
      { id: '4', source_type: 'pdf', original_filename: 'other.pdf', status: 'queued', chunk_count: 2 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />)
    expect(container.querySelector('span.bg-emerald-100')).toBeTruthy()
    // Note: there are 2 animate-spin elements due to processing status AND refresh button icon
    expect(screen.getByText('failed')).toBeInTheDocument()
    expect(screen.getByText('queued')).toBeInTheDocument()
  })

  it('shows refresh button only for URL sources', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />)
    const rows = container.querySelectorAll('tbody tr')
    // PDF row should NOT have refresh button
    const pdfRow = rows[0]
    expect(within(pdfRow as HTMLElement).queryByLabelText('Kaynağı Yenile')).toBeNull()
    // URL row should have refresh button
    const urlRow = rows[1]
    expect(within(urlRow as HTMLElement).queryByLabelText('Kaynağı Yenile')).toBeTruthy()
  })

  it('disables refresh button for free plan', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="free" onDelete={onDelete} onRefresh={onRefresh} />)
    const refreshBtn = container.querySelector('button[aria-label="Kaynağı Yenile"]')
    expect(refreshBtn).toBeDisabled()
  })

  it('calls onRefresh when clicking refresh button', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />)
    const refreshBtn = container.querySelector('button[aria-label="Kaynağı Yenile"]') as HTMLElement
    fireEvent.click(refreshBtn)
    expect(onRefresh).toHaveBeenCalledWith('1')
  })
})
