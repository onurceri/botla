import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import HeaderActions from '../HeaderActions'

describe('HeaderActions', () => {
  it('renders title and create button for new chatbot', () => {
    const onCreate = vi.fn()
    render(
      <HeaderActions
        isNew
        name=""
        isDeleting={false}
        isCreating={false}
        onDelete={() => {}}
        onCreate={onCreate}
      />
    )
    expect(screen.getByText('Yeni Chatbot')).toBeInTheDocument()
    const createButton = screen.getByRole('button', { name: /Oluştur/i })
    fireEvent.click(createButton)
    expect(onCreate).toHaveBeenCalledTimes(1)
  })

  it('renders name and delete button without create for existing chatbot', () => {
    const onDelete = vi.fn()
    render(
      <HeaderActions
        isNew={false}
        name="Destek Botu"
        isDeleting={false}
        onDelete={onDelete}
      />
    )
    expect(screen.getByText('Destek Botu')).toBeInTheDocument()
    const deleteButton = screen.getByLabelText('Sil')
    fireEvent.click(deleteButton)
    expect(onDelete).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('button', { name: /Değişiklikleri Kaydet/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Oluştur/i })).not.toBeInTheDocument()
  })
})
