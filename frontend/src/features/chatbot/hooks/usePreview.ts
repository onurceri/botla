import { useEffect, useState } from 'react'

export function usePreview() {
  const [previewOpen, setPreviewOpen] = useState(true)
  const [sessionId, setSessionId] = useState('')

  useEffect(() => {
    setSessionId(`playground-${Math.random().toString(36).substring(2, 15)}`)
  }, [])

  return { previewOpen, setPreviewOpen, sessionId }
}

