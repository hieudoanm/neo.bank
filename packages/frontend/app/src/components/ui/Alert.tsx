import type { ReactNode } from 'react'

interface AlertProps {
  type: 'error' | 'warning' | 'success' | 'info'
  title?: string
  children?: ReactNode
  className?: string
  role?: string
}

const typeClass: Record<AlertProps['type'], string> = {
  error: 'alert alert-error',
  warning: 'alert alert-warning',
  success: 'alert alert-success',
  info: 'alert alert-info',
}

export const Alert = ({ type, title, children, className = '', role = 'alert' }: AlertProps) => {
  return (
    <div className={`alert ${typeClass[type]} ${className}`} role={role}>
      {title && <h3 className="alert-title">{title}</h3>}
      {children && <div>{children}</div>}
    </div>
  )
}
