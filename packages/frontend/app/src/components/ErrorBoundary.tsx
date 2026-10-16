import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Alert } from '@/components/ui/Alert'
import { Button } from '@/components/ui/Button'
import { logger } from '@/lib/log'

interface ErrorBoundaryProps {
  children: ReactNode
}

interface ErrorBoundaryState {
  error: Error | null
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    logger.error('unhandled ui error', { error, componentStack: info.componentStack })
  }

  render(): ReactNode {
    if (!this.state.error) return this.props.children
    return (
      <div className="flex min-h-screen items-center justify-center bg-background-secondary p-6">
        <div className="w-full max-w-md">
          <Alert type="error" title="Something went wrong">
            <p className="text-text-secondary">
              An unexpected error occurred. Reload the page to try again.
            </p>
            <Button className="mt-4" onClick={() => window.location.reload()}>
              Reload page
            </Button>
          </Alert>
        </div>
      </div>
    )
  }
}
