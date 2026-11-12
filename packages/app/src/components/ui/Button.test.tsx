import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from './Button'

describe('Button', () => {
  it('renders children and defaults to the primary variant', () => {
    render(<Button>Save</Button>)
    expect(screen.getByRole('button', { name: 'Save' })).toHaveClass('btn', 'btn-primary')
  })

  it('applies the requested variant and extra className', () => {
    render(
      <Button variant="danger" className="extra">
        Delete
      </Button>,
    )
    expect(screen.getByRole('button', { name: 'Delete' })).toHaveClass('btn-error', 'extra')
  })

  it('fires onClick and forwards native props', () => {
    const onClick = vi.fn()
    render(
      <Button type="submit" onClick={onClick}>
        Go
      </Button>,
    )
    const button = screen.getByRole('button', { name: 'Go' })
    expect(button).toHaveAttribute('type', 'submit')
    fireEvent.click(button)
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('does not fire onClick when disabled', () => {
    const onClick = vi.fn()
    render(
      <Button disabled onClick={onClick}>
        Disabled
      </Button>,
    )
    fireEvent.click(screen.getByRole('button', { name: 'Disabled' }))
    expect(onClick).not.toHaveBeenCalled()
  })
})
