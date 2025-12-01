export type Chatbot = {
  id: string
  name: string
  description?: string
}

export type Source = {
  id: string
  type: 'pdf' | 'url'
  url?: string
}
