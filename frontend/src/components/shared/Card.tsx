import { FC, PropsWithChildren } from 'react'

type Props = PropsWithChildren<{
  title?: string
}>

const Card: FC<Props> = ({ title, children }) => {
  return (
    <div className="rounded-lg border bg-card p-4 text-card-foreground">
      {title && <div className="mb-2 text-base font-semibold">{title}</div>}
      {children}
    </div>
  )
}

export default Card
