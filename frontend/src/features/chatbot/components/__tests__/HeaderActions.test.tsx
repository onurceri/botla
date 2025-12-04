import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import HeaderActions from '../HeaderActions'

describe('HeaderActions', () => {
  it('renders title and save button for new chatbot', () => {
    const onSave = vi.fn()
    render(
      <HeaderActions
        isNew={true}
        name=""
        isDeleting={false}
        isSaving={false}
        onDelete={() => {}}
        onSave={onSave}
      />
    )
    expect(screen.getByText('Yeni Chatbot')).toBeInTheDocument()
    const createButton = screen.getByRole('button', { name: /Oluştur/i })
    fireEvent.click(createButton)
    expect(onSave).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('button', { name: /Değişiklikleri Kaydet/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Trash/i })).not.toBeInTheDocument()
  })

  it('renders name, delete and save for existing chatbot', () => {
    const onSave = vi.fn()
    const onDelete = vi.fn()
    render(
      <HeaderActions
        isNew={false}
        name="Destek Botu"
        isDeleting={false}
        isSaving={false}
        onDelete={onDelete}
        onSave={onSave}
      />
    )
    expect(screen.getByText('Destek Botu')).toBeInTheDocument()
    const saveButton = screen.getByRole('button', { name: /Değişiklikleri Kaydet/i })
    fireEvent.click(saveButton)
    expect(onSave).toHaveBeenCalledTimes(1)
    const deleteButton = screen.getByLabelText('Sil')
    fireEvent.click(deleteButton)
    expect(onDelete).toHaveBeenCalledTimes(1)
  })
})
