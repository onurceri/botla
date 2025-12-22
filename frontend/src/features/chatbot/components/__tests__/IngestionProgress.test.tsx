import { describe, it, expect, afterEach } from 'vitest'
import { render, cleanup, within } from '@testing-library/react'
import IngestionProgress, { ProcessingSource } from '../IngestionProgress'

// Cleanup after each test to prevent element accumulation
afterEach(() => {
  cleanup()
})

describe('IngestionProgress', () => {
  const mockSources = (overrides?: Partial<ProcessingSource>[]): ProcessingSource[] => [
    { id: '1', source_type: 'pdf', name: 'document.pdf', status: 'processing', progress: 50, ...overrides?.[0] },
    { id: '2', source_type: 'url', name: 'https://example.com', status: 'queued', ...overrides?.[1] },
  ]

  it('renders nothing when sources array is empty', () => {
    const { container } = render(<IngestionProgress sources={[]} />)
    expect(container.querySelector('[data-testid="ingestion-progress"]')).toBeNull()
  })

  it('displays processing count header', () => {
    const { container } = render(<IngestionProgress sources={mockSources()} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('2')).toBeInTheDocument()
    expect(within(wrapper as HTMLElement).getByText(/kaynak işleniyor/)).toBeInTheDocument()
  })

  it('renders progress items for each source', () => {
    const { container } = render(<IngestionProgress sources={mockSources()} />)
    
    const items = container.querySelectorAll('[data-testid="progress-item"]')
    expect(items.length).toBe(2)
  })

  it('displays source name correctly', () => {
    const { container } = render(<IngestionProgress sources={mockSources()} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('document.pdf')).toBeInTheDocument()
    expect(within(wrapper as HTMLElement).getByText('https://example.com')).toBeInTheDocument()
  })

  it('shows correct status text for processing', () => {
    const { container } = render(<IngestionProgress sources={[{ id: '1', source_type: 'pdf', name: 'test.pdf', status: 'processing' }]} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('İşleniyor')).toBeInTheDocument()
  })

  it('shows correct status text for queued', () => {
    const { container } = render(<IngestionProgress sources={[{ id: '1', source_type: 'pdf', name: 'test.pdf', status: 'queued' }]} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('Beklemede')).toBeInTheDocument()
  })

  it('shows correct status text for completed', () => {
    const { container } = render(<IngestionProgress sources={[{ id: '1', source_type: 'pdf', name: 'test.pdf', status: 'completed' }]} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('Tamamlandı')).toBeInTheDocument()
  })

  it('shows correct status text for failed', () => {
    const { container } = render(<IngestionProgress sources={[{ id: '1', source_type: 'pdf', name: 'test.pdf', status: 'failed' }]} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('Başarısız')).toBeInTheDocument()
  })

  it('displays error message for failed sources', () => {
    const { container } = render(<IngestionProgress sources={[{ 
      id: '1', 
      source_type: 'url', 
      name: 'https://bad.com', 
      status: 'failed',
      error_message: 'Connection timeout'
    }]} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).getByText('Connection timeout')).toBeInTheDocument()
  })

  it('shows progress bar for processing sources', () => {
    const { container } = render(<IngestionProgress sources={[{ 
      id: '1', 
      source_type: 'pdf', 
      name: 'test.pdf', 
      status: 'processing',
      progress: 75
    }]} />)
    
    const progressBar = container.querySelector('[data-testid="progress-bar"]')
    expect(progressBar).toBeInTheDocument()
  })

  it('shows progress bar for queued sources', () => {
    const { container } = render(<IngestionProgress sources={[{ 
      id: '1', 
      source_type: 'pdf', 
      name: 'test.pdf', 
      status: 'queued'
    }]} />)
    
    const progressBar = container.querySelector('[data-testid="progress-bar"]')
    expect(progressBar).toBeInTheDocument()
  })

  it('does not show progress bar for completed sources', () => {
    const { container } = render(<IngestionProgress sources={[{ 
      id: '1', 
      source_type: 'pdf', 
      name: 'test.pdf', 
      status: 'completed'
    }]} />)
    
    const progressBar = container.querySelector('[data-testid="progress-bar"]')
    expect(progressBar).toBeNull()
  })

  it('does not show processing header when all sources are completed', () => {
    const { container } = render(<IngestionProgress sources={[
      { id: '1', source_type: 'pdf', name: 'test.pdf', status: 'completed' },
      { id: '2', source_type: 'url', name: 'https://example.com', status: 'completed' }
    ]} />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')!
    expect(within(wrapper as HTMLElement).queryByText(/kaynak işleniyor/)).toBeNull()
  })

  it('applies custom className', () => {
    const { container } = render(<IngestionProgress sources={mockSources()} className="test-class" />)
    
    const wrapper = container.querySelector('[data-testid="ingestion-progress"]')
    expect(wrapper).toHaveClass('test-class')
  })
})
