import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import SourceList from '../SourceList'

describe('SourceList', () => {
  it('renders sources and calls delete', () => {
    const onDelete = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
    ] as any
    render(<SourceList sources={sources} onDelete={onDelete} />)
    expect(screen.getByText('file.pdf')).toBeInTheDocument()
    const delBtn = screen.getByLabelText('Kaynağı Sil')
    fireEvent.click(delBtn)
    expect(onDelete).toHaveBeenCalledWith('1')
  })

  it('renders status variants with proper indicators', () => {
    const onDelete = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'url', source_url: 'https://a.com', status: 'processing', chunk_count: 1 },
      { id: '3', source_type: 'text', original_filename: 'note', status: 'failed', chunk_count: 0 },
      { id: '4', source_type: 'pdf', original_filename: 'other.pdf', status: 'queued', chunk_count: 2 },
    ] as any
    const { container } = render(<SourceList sources={sources} onDelete={onDelete} />)
    expect(container.querySelector('span.bg-emerald-100')).toBeTruthy()
    expect(container.querySelector('.animate-spin')).toBeTruthy()
    expect(screen.getByText('failed')).toBeInTheDocument()
    expect(screen.getByText('queued')).toBeInTheDocument()
  })
})
