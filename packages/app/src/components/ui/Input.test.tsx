import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Input } from './Input'

describe('Input', () => {
  it('associates the label with the input using the name when no id is given', () => {
    render(<Input label="Email" name="email" />)
    expect(screen.getByLabelText('Email')).toHaveAttribute('id', 'email')
  })

  it('prefers an explicit id', () => {
    render(<Input label="Email" id="custom" name="email" />)
    expect(screen.getByLabelText('Email')).toHaveAttribute('id', 'custom')
  })

  it('renders without a label when none is provided', () => {
    render(<Input name="email" />)
    expect(screen.queryByText('Email')).not.toBeInTheDocument()
  })

  it('marks the input invalid and links the error message', () => {
    render(<Input label="Amount" name="amount" error="Amount is required" />)
    const input = screen.getByLabelText('Amount')
    expect(input).toHaveAttribute('aria-invalid', 'true')
    expect(input).toHaveAttribute('aria-describedby', 'amount-error')
    expect(screen.getByText('Amount is required')).toHaveAttribute('id', 'amount-error')
  })

  it('does not set aria attributes when there is no error', () => {
    render(<Input label="Amount" name="amount" />)
    const input = screen.getByLabelText('Amount')
    expect(input).not.toHaveAttribute('aria-invalid')
    expect(input).not.toHaveAttribute('aria-describedby')
  })
})
