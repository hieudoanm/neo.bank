import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card } from './Card'

describe('Card', () => {
  it('renders children and accepts an extra className', () => {
    render(<Card className="max-w-md">Balance</Card>)
    const card = screen.getByText('Balance').closest('section')!
    expect(card).toHaveClass('card', 'max-w-md')
    expect(screen.getByText('Balance')).toBeInTheDocument()
  })
})
