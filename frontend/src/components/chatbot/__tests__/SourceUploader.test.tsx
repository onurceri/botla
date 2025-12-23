import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, screen, fireEvent, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SourceUploader from '../SourceUploader'

describe('SourceUploader', () => {
  it('toggles modes and calls upload handlers', async () => {
    const user = userEvent.setup()
    const onUploadPDF = vi.fn().mockResolvedValue(undefined)
    const onUploadURL = vi.fn().mockResolvedValue(undefined)
    const onUploadText = vi.fn().mockResolvedValue(undefined)

    render(
      <QueryWrapper>
        <SourceUploader
          onUploadPDF={onUploadPDF}
          onUploadURL={onUploadURL}
          onUploadText={onUploadText}
        />
      </QueryWrapper>,
    )

    await user.click(screen.getByText('PDF Yükle'))
    const file = new File(['dummy'], 'doc.pdf', { type: 'application/pdf' })
    const hiddenInput = document.getElementById('pdf-upload') as HTMLInputElement
    fireEvent.change(hiddenInput, { target: { files: [file] } })
    expect(onUploadPDF).toHaveBeenCalled()

    const urlButtons = screen.getAllByText('Web Sitesi')
    await user.click(urlButtons[urlButtons.length - 1])
    const urlInputs = screen.getAllByPlaceholderText('https://example.com')
    const urlInput = urlInputs[urlInputs.length - 1]
    await user.type(urlInput, 'https://example.com')
    const ekleButtons = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(ekleButtons[ekleButtons.length - 1])
    expect(onUploadURL).toHaveBeenCalledWith('https://example.com')

    const textButtons = screen.getAllByText('Metin')
    await user.click(textButtons[textButtons.length - 1])
    const textarea = screen.getByPlaceholderText('Metin içeriğini buraya yapıştırın...')
    await user.type(textarea, 'hello')
    const buttons = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(buttons[buttons.length - 1])
    expect(onUploadText).toHaveBeenCalledWith('hello')
  })

  it('does not call URL/Text handlers on empty input', async () => {
    const user = userEvent.setup()
    const onUploadPDF = vi.fn()
    const onUploadURL = vi.fn()
    const onUploadText = vi.fn()
    render(
      <QueryWrapper>
        <SourceUploader
          onUploadPDF={onUploadPDF}
          onUploadURL={onUploadURL}
          onUploadText={onUploadText}
        />
      </QueryWrapper>,
    )
    const urlBtns = screen.getAllByText('Web Sitesi')
    await user.click(urlBtns[urlBtns.length - 1])
    const ekleBtns = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(ekleBtns[ekleBtns.length - 1])
    expect(onUploadURL).not.toHaveBeenCalled()

    const textBtns = screen.getAllByText('Metin')
    await user.click(textBtns[textBtns.length - 1])
    const buttons = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(buttons[buttons.length - 1])
    expect(onUploadText).not.toHaveBeenCalled()
  })

  it('validates URL format client-side and shows error toast', async () => {
    const user = userEvent.setup()
    const onUploadPDF = vi.fn()
    const onUploadURL = vi.fn()
    const onUploadText = vi.fn()
    render(
      <QueryWrapper>
        <SourceUploader
          onUploadPDF={onUploadPDF}
          onUploadURL={onUploadURL}
          onUploadText={onUploadText}
        />
      </QueryWrapper>,
    )
    const urlBtns = screen.getAllByText('Web Sitesi')
    await user.click(urlBtns[urlBtns.length - 1])
    const urlInputs2 = screen.getAllByPlaceholderText('https://example.com')
    const urlInput = urlInputs2[urlInputs2.length - 1]
    await user.type(urlInput, 'not-a-url')
    const ekleBtns2 = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(ekleBtns2[ekleBtns2.length - 1])
    expect(onUploadURL).not.toHaveBeenCalled()
    expect(await screen.findByText('Lütfen geçerli bir URL girin.')).toBeInTheDocument()
  })

  it('shows error toast when URL upload fails (unreachable)', async () => {
    const user = userEvent.setup()
    const onUploadPDF = vi.fn()
    const onUploadURL = vi
      .fn()
      .mockRejectedValue({ response: { data: { message: 'URL erişilemedi' } } })
    const onUploadText = vi.fn()
    render(
      <QueryWrapper>
        <SourceUploader
          onUploadPDF={onUploadPDF}
          onUploadURL={onUploadURL}
          onUploadText={onUploadText}
        />
      </QueryWrapper>,
    )
    const urlBtns = screen.getAllByText('Web Sitesi')
    await user.click(urlBtns[urlBtns.length - 1])
    const urlInputs = screen.getAllByPlaceholderText('https://example.com')
    const urlInput = urlInputs[urlInputs.length - 1]
    await user.type(urlInput, 'https://bad.example.com')
    const ekleBtns = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(ekleBtns[ekleBtns.length - 1])
    expect(onUploadURL).toHaveBeenCalledWith('https://bad.example.com')
    expect(await screen.findByText('URL erişilemedi')).toBeInTheDocument()
  })

  it('shows server error toast when PDF upload fails', async () => {
    const user = userEvent.setup()
    const serverError = { response: { data: { message: 'Sunucu hatası' } } }
    render(
      <QueryWrapper>
        <SourceUploader
          onUploadPDF={vi.fn().mockRejectedValueOnce(serverError)}
          onUploadURL={vi.fn()}
          onUploadText={vi.fn()}
        />
      </QueryWrapper>,
    )
    const pdfBtns = screen.getAllByText('PDF Yükle')
    await user.click(pdfBtns[pdfBtns.length - 1])
    const fileInput = document.querySelector('#pdf-upload') as HTMLInputElement
    const file = new File(['%PDF-1.4'], 'doc.pdf', { type: 'application/pdf' })
    fireEvent.change(fileInput, { target: { files: [file] } })
    expect(await screen.findByText('Sunucu hatası')).toBeInTheDocument()
  })

  it('rejects non-PDF file type with error toast', async () => {
    const onUploadPDF = vi.fn()
    render(
      <QueryWrapper>
        <SourceUploader onUploadPDF={onUploadPDF} onUploadURL={vi.fn()} onUploadText={vi.fn()} />
      </QueryWrapper>,
    )
    const pdfBtns1 = screen.getAllByText('PDF Yükle')
    await userEvent.click(pdfBtns1[pdfBtns1.length - 1])
    const hiddenInput = document.getElementById('pdf-upload') as HTMLInputElement
    const badFile = new File(['x'], 'doc.txt', { type: 'text/plain' })
    fireEvent.change(hiddenInput, { target: { files: [badFile] } })
    expect(onUploadPDF).not.toHaveBeenCalled()
    expect(await screen.findByText('Yalnızca PDF dosyaları desteklenir.')).toBeInTheDocument()
  })

  it('rejects PDF larger than 50MB with error toast', async () => {
    const onUploadPDF = vi.fn()
    render(
      <QueryWrapper>
        <SourceUploader onUploadPDF={onUploadPDF} onUploadURL={vi.fn()} onUploadText={vi.fn()} />
      </QueryWrapper>,
    )
    const pdfBtns2 = screen.getAllByText('PDF Yükle')
    await userEvent.click(pdfBtns2[pdfBtns2.length - 1])
    const hiddenInput = document.getElementById('pdf-upload') as HTMLInputElement
    const bigFile = new File(['a'], 'big.pdf', { type: 'application/pdf' })
    Object.defineProperty(bigFile, 'size', { value: 51 * 1024 * 1024 })
    fireEvent.change(hiddenInput, { target: { files: [bigFile] } })
    expect(onUploadPDF).not.toHaveBeenCalled()
    expect(await screen.findByText("Dosya boyutu 50MB'den büyük olamaz.")).toBeInTheDocument()
  })

  it('closes input area when X is clicked', async () => {
    const user = userEvent.setup()
    const utils = render(
      <QueryWrapper>
        <SourceUploader onUploadPDF={vi.fn()} onUploadURL={vi.fn()} onUploadText={vi.fn()} />
      </QueryWrapper>,
    )
    const view = within(utils.container)
    const urlBtns = view.getAllByText('Web Sitesi')
    await user.click(urlBtns[urlBtns.length - 1])
    expect(view.getAllByPlaceholderText('https://example.com').length).toBeGreaterThan(0)
    const closeBtn = utils.container.querySelector('button .lucide-x')
      ?.parentElement as HTMLButtonElement
    await user.click(closeBtn)
    expect(view.queryAllByPlaceholderText('https://example.com').length).toBe(0)
  })
})
