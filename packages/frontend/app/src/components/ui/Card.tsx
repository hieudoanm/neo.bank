import type { ReactNode } from 'react'

interface CardProps {
  className?: string
  children: ReactNode
}

export const Card = ({ className = '', children }: CardProps) => {
  return (
    <section className={`card card-body bg-base-100 shadow-sm ${className}`}>
      {children}
    </section>
  )
}
