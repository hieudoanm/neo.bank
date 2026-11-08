import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { authApi, HttpError } from '@/lib/api/client'
import { useAuth } from '@/context/AuthContext'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { Alert } from '@/components/ui/Alert'
import { validateEmail, validatePassword, validateConfirmPassword } from '@/lib/validation/auth'

interface FormData {
  email: string
  password: string
  confirmPassword: string
}

export const RegisterPage = () => {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [form, setForm] = useState<FormData>({ email: '', password: '', confirmPassword: '' })
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [serverError, setServerError] = useState<string | null>(null)
  const [networkError, setNetworkError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: (data) => {
      setServerError(null)
      setNetworkError(null)
      login(
        { userId: data.userId, accountId: data.accountId, email: form.email },
        data.accessToken,
        data.refreshToken,
      )
      navigate('/', { replace: true })
    },
    onError: (err) => {
      setFieldErrors({})
      setServerError(null)
      setNetworkError(null)
      const status = err instanceof HttpError ? err.status : 0
      const message = err instanceof HttpError ? err.error.message : undefined
      if (status === 400) {
        setServerError(message ?? 'Request could not be completed')
      } else if (status === 409) {
        setFieldErrors({ email: 'Email is already registered' })
        setServerError('Email is already registered')
      } else if (status >= 500 || status === 0) {
        setNetworkError(
          status === 0
            ? 'Unable to connect to the server. Please try again later.'
            : 'The server is temporarily unavailable. Please try again later.',
        )
      } else {
        setServerError(message ?? 'An unexpected error occurred')
      }
    },
  })

  const handleChange = (key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [key]: e.target.value }))
    if (fieldErrors[key]) {
      setFieldErrors((prev) => {
        const next = { ...prev }
        delete next[key]
        return next
      })
    }
    if (serverError) setServerError(null)
  }

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setFieldErrors({})
    setServerError(null)
    setNetworkError(null)

    const next: Record<string, string> = {}
    const emailError = validateEmail(form.email)
    const passwordError = validatePassword(form.password)
    const confirmError = validateConfirmPassword(form.password, form.confirmPassword)
    if (emailError) next.email = emailError
    if (passwordError) next.password = passwordError
    if (confirmError) next.confirmPassword = confirmError
    if (Object.keys(next).length > 0) {
      setFieldErrors(next)
      return
    }

    mutation.mutate({ email: form.email, password: form.password })
  }

  const isLoading = mutation.isPending

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-6">
      <Card className="w-full max-w-sm mx-auto">
        <h1 className="text-text-primary text-2xl font-semibold tracking-tight text-center mb-1">
          Create your account
        </h1>
        <p className="text-text-secondary text-center text-sm mb-6">Join NEO Bank today</p>

        {networkError && (
          <Alert type="error" className="mb-4" role="alert">
            {networkError}
          </Alert>
        )}
        {serverError && (
          <Alert type="error" className="mb-4" role="alert">
            {serverError}
          </Alert>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Email"
            type="email"
            name="email"
            autoComplete="email"
            value={form.email}
            onChange={handleChange('email')}
            error={fieldErrors.email}
            required
          />
          <Input
            label="Password"
            type="password"
            name="password"
            autoComplete="new-password"
            value={form.password}
            onChange={handleChange('password')}
            error={fieldErrors.password}
            required
          />
          <Input
            label="Confirm password"
            type="password"
            name="confirmPassword"
            autoComplete="new-password"
            value={form.confirmPassword}
            onChange={handleChange('confirmPassword')}
            error={fieldErrors.confirmPassword}
            required
          />
          <Button type="submit" variant="primary" className="w-full" disabled={isLoading}>
            {isLoading ? 'Creating account...' : 'Create account'}
          </Button>
        </form>

        <p className="text-center text-text-secondary text-sm mt-6">
          Already have an account?{' '}
          <a href="/login" className="text-primary hover:text-primary-hover underline">
            Sign in
          </a>
        </p>
      </Card>
    </div>
  )
}
