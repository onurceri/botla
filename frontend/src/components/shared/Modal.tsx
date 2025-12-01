import { FC, PropsWithChildren } from 'react'

type Props = PropsWithChildren<{
  open: boolean
}>

const Modal: FC<Props> = ({ open, children }) => {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-lg rounded-lg border bg-popover p-4 text-popover-foreground shadow-lg">
        {children}
      </div>
    </div>
  )
}

export default Modal
