import { ButtonHTMLAttributes, FC } from 'react'

type Props = ButtonHTMLAttributes<HTMLButtonElement>

const Button: FC<Props> = ({ className = '', ...props }) => {
  return <button {...props} className={`inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground ${className}`} />
}

export default Button
