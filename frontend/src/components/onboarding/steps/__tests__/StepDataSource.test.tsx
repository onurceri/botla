/**
 * Unit tests for StepDataSource component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { StepDataSource } from '../StepDataSource'

const defaultProps = {
  sourceType: 'text' as const,
  textContent: '',
  urlContent: '',
  pdfFile: null,
  onSourceTypeChange: vi.fn(),
  onTextContentChange: vi.fn(),
  onUrlContentChange: vi.fn(),
  onFileSelect: vi.fn(),
  onFileRemove: vi.fn(),
}

describe('StepDataSource', () => {
  it('renders the step title and description', () => {
    render(<StepDataSource {...defaultProps} />)

    expect(screen.getByText('Bilgi Kaynağı Ekleyin')).toBeInTheDocument()
    expect(screen.getByText('Botunuzun öğrenmesini istediğiniz içeriği seçin')).toBeInTheDocument()
  })

  it('renders all three source type buttons', () => {
    render(<StepDataSource {...defaultProps} />)

    expect(screen.getByText('Metin')).toBeInTheDocument()
    expect(screen.getByText('URL')).toBeInTheDocument()
    expect(screen.getByText('PDF')).toBeInTheDocument()
  })

  it('calls onSourceTypeChange when clicking source type buttons', () => {
    const onChange = vi.fn()
    render(<StepDataSource {...defaultProps} onSourceTypeChange={onChange} />)

    fireEvent.click(screen.getByText('URL'))
    expect(onChange).toHaveBeenCalledWith('url')

    fireEvent.click(screen.getByText('PDF'))
    expect(onChange).toHaveBeenCalledWith('file')
  })

  describe('text source type', () => {
    it('shows text area when sourceType is text', () => {
      render(<StepDataSource {...defaultProps} sourceType="text" />)

      expect(screen.getByLabelText('İçerik')).toBeInTheDocument()
    })

    it('calls onTextContentChange when typing', () => {
      const onChange = vi.fn()
      render(<StepDataSource {...defaultProps} sourceType="text" onTextContentChange={onChange} />)

      fireEvent.change(screen.getByLabelText('İçerik'), { target: { value: 'Some content' } })
      expect(onChange).toHaveBeenCalledWith('Some content')
    })

    it('shows character count', () => {
      render(<StepDataSource {...defaultProps} sourceType="text" textContent="Hello" />)

      expect(screen.getByText('Minimum 50 karakter (5/50)')).toBeInTheDocument()
    })
  })

  describe('url source type', () => {
    it('shows URL input when sourceType is url', () => {
      render(<StepDataSource {...defaultProps} sourceType="url" />)

      expect(screen.getByLabelText("Web Sitesi URL'si")).toBeInTheDocument()
    })

    it('calls onUrlContentChange when typing', () => {
      const onChange = vi.fn()
      render(<StepDataSource {...defaultProps} sourceType="url" onUrlContentChange={onChange} />)

      fireEvent.change(screen.getByLabelText("Web Sitesi URL'si"), {
        target: { value: 'https://example.com' },
      })
      expect(onChange).toHaveBeenCalledWith('https://example.com')
    })
  })

  describe('file source type', () => {
    it('shows file upload area when sourceType is file', () => {
      render(<StepDataSource {...defaultProps} sourceType="file" />)

      expect(screen.getByText('PDF Dosyası Seçin')).toBeInTheDocument()
    })

    it('shows uploaded file info when pdfFile is set', () => {
      const file = new File(['content'], 'test.pdf', { type: 'application/pdf' })
      Object.defineProperty(file, 'size', { value: 1024 * 1024 }) // 1MB

      render(<StepDataSource {...defaultProps} sourceType="file" pdfFile={file} />)

      expect(screen.getByText('test.pdf')).toBeInTheDocument()
      expect(screen.getByText('1.00 MB')).toBeInTheDocument()
    })

    it('calls onFileRemove when clicking change button', () => {
      const onRemove = vi.fn()
      const file = new File(['content'], 'test.pdf', { type: 'application/pdf' })

      render(
        <StepDataSource {...defaultProps} sourceType="file" pdfFile={file} onFileRemove={onRemove} />,
      )

      fireEvent.click(screen.getByText('Değiştir'))
      expect(onRemove).toHaveBeenCalled()
    })
  })
})
