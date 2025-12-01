import { FC } from 'react'

type Props = {
  name: string
  description?: string
}

const ChatbotCard: FC<Props> = ({ name, description }) => {
  return (
    <div className="rounded-lg border bg-card p-4 text-card-foreground">
      <div className="text-lg font-semibold">{name}</div>
      {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
    </div>
  )
}

export default ChatbotCard
