import { useMemo, useState, type ComponentProps } from "react"

import { ApiError, Client } from "@/client/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { getBrowserApiUrl, setStoredAuthToken } from "@/lib/auth"

type AuthMode = "login" | "register"

type AuthFormProps = {
  mode: AuthMode
}

type FormSubmitHandler = NonNullable<ComponentProps<"form">["onSubmit"]>

function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.body && typeof error.body === "object") {
    const message = (error.body as Record<string, unknown>).message
    if (typeof message === "string" && message.length > 0) {
      return message
    }
  }

  if (error instanceof Error && error.message.length > 0) {
    return error.message
  }

  return "Something went wrong. Please try again."
}

export function AuthForm({ mode }: AuthFormProps) {
  const client = useMemo(() => new Client(getBrowserApiUrl()), [])

  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const isLogin = mode === "login"

  const handleSubmit: FormSubmitHandler = async (event) => {
    event.preventDefault()
    setError(null)
    setSuccess(null)

    if (username.trim().length < 3) {
      setError("Username must be at least 3 characters.")
      return
    }

    if (password.length < 6) {
      setError("Password must be at least 6 characters.")
      return
    }

    setIsLoading(true)

    try {
      if (isLogin) {
        const response = await client.login({ username: username.trim(), password })
        setStoredAuthToken(response.token)
        setSuccess("Logged in. Redirecting...")

        window.setTimeout(() => {
          window.location.href = "/"
        }, 450)
      } else {
        await client.register({ username: username.trim(), password })
        setSuccess("Account created. Redirecting to login...")

        window.setTimeout(() => {
          window.location.href = "/login"
        }, 550)
      }
    } catch (submitError) {
      setError(getErrorMessage(submitError))
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-1">
        <label htmlFor="username" className="text-sm font-medium text-slate-700">
          Username
        </label>
        <Input
          id="username"
          name="username"
          autoComplete="username"
          value={username}
          onChange={(event) => setUsername(event.target.value)}
          className="h-11 border-slate-300 bg-white"
          placeholder="your-username"
          disabled={isLoading}
          required
        />
      </div>

      <div className="space-y-1">
        <label htmlFor="password" className="text-sm font-medium text-slate-700">
          Password
        </label>
        <Input
          id="password"
          name="password"
          type="password"
          autoComplete={isLogin ? "current-password" : "new-password"}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="h-11 border-slate-300 bg-white"
          placeholder="At least 6 characters"
          disabled={isLoading}
          required
        />
      </div>

      {error ? <p className="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">{error}</p> : null}
      {success ? <p className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{success}</p> : null}

      <Button type="submit" className="h-11 w-full" disabled={isLoading}>
        {isLoading ? "Please wait..." : isLogin ? "Log in" : "Create account"}
      </Button>
    </form>
  )
}
