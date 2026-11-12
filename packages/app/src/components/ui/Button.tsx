import type { ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'secondary' | 'danger' | 'outline'

const variants: Record<Variant, string> = {
  primary: 'btn btn-primary',
  secondary: 'btn btn-secondary',
  danger: 'btn btn-error',
  outline: 'btn btn-outline',
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  className?: string
}

export const Button = ({
  variant = 'primary',
  className = '',
  children,
  disabled,
  ...props
}: ButtonProps) => {
  return (
    <button className={`${variants[variant]} ${className}`} disabled={disabled} {...props}>
      {children}
    </button>
  )
}
