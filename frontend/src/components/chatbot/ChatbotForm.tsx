import { FC, useState } from 'react'

type Props = {
  onSubmit?: (payload: { name: string; description?: string }) => void
}

const ChatbotForm: FC<Props> = ({ onSubmit }) => {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit?.({ name, description: description || undefined })
  }

  return (
    <form className="space-y-3" onSubmit={handleSubmit}>
      <div className="grid gap-1">
        <label className="text-sm">İsim</label>
        <input className="rounded border bg-input px-3 py-2" value={name} onChange={(e) => setName(e.target.value)} />
      </div>
      <div className="grid gap-1">
        <label className="text-sm">Açıklama</label>
        <textarea className="rounded border bg-input px-3 py-2" value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>
      <button type="submit" className="rounded bg-primary px-3 py-2 text-primary-foreground">Kaydet</button>
    </form>
  )
}

export default ChatbotForm
