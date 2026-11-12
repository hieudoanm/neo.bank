import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Alert } from './Alert'

describe('Alert', () => {
  it('renders children inside an alert region', () => {
    render(<Alert type="error">Something failed</Alert>)
    expect(screen.getByRole('alert')).toHaveTextContent('Something failed')
  })

  it('renders a title when provided', () => {
    render(
      <Alert type="info" title="Heads up">
        Body
      </Alert>,
    )
    expect(screen.getByRole('heading', { name: 'Heads up' })).toBeInTheDocument()
  })

  it('omits the title heading when no title is given', () => {
    render(<Alert type="success">Done</Alert>)
    expect(screen.queryByRole('heading')).not.toBeInTheDocument()
  })

  it('allows the role and className to be overridden', () => {
    render(
      <Alert type="warning" role="status" className="extra">
        Careful
      </Alert>,
    )
    expect(screen.getByRole('status')).toHaveClass('alert-warning', 'extra')
  })
})
