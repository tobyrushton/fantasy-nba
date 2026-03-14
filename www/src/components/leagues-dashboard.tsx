import { useCallback, useEffect, useMemo, useState, type ComponentProps } from "react"

import { ApiError, Client, type LeagueResponse } from "@/client/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { clearStoredAuthToken, getBrowserApiUrl, getStoredAuthToken } from "@/lib/auth"

type AuthState = "loading" | "unauthenticated" | "ready"
type FormSubmitHandler = NonNullable<ComponentProps<"form">["onSubmit"]>

function getApiErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.body && typeof error.body === "object") {
    const record = error.body as Record<string, unknown>
    const message = record.error ?? record.message
    if (typeof message === "string" && message.length > 0) {
      return message
    }
  }

  if (error instanceof Error && error.message.length > 0) {
    return error.message
  }

  return "Request failed"
}

function isUnauthorizedError(error: unknown): error is ApiError {
  return error instanceof ApiError && error.status === 401
}

function parseUserIdFromJwt(token: string): number | null {
  const parts = token.split(".")
  if (parts.length < 2) {
    return null
  }

  try {
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/")
    const padded = payload.padEnd(payload.length + ((4 - (payload.length % 4)) % 4), "=")
    const decoded = window.atob(padded)
    const claims = JSON.parse(decoded) as Record<string, unknown>
    const userId = claims.user_id

    if (typeof userId === "number" && Number.isFinite(userId)) {
      return userId
    }

    return null
  } catch {
    return null
  }
}

export function LeaguesDashboard() {
  const client = useMemo(() => new Client(getBrowserApiUrl()), [])

  const [authState, setAuthState] = useState<AuthState>("loading")
  const [token, setToken] = useState<string | null>(null)
  const [userId, setUserId] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [leagues, setLeagues] = useState<LeagueResponse[]>([])
  const [name, setName] = useState("")
  const [joinLeagueId, setJoinLeagueId] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const invalidateSession = useCallback(() => {
    clearStoredAuthToken()
    setToken(null)
    setUserId(null)
    setAuthState("unauthenticated")
  }, [])

  const loadLeagues = useCallback(async () => {
    if (!token) {
      return
    }

    setError(null)
    setIsLoading(true)
    client.setToken(token)

    try {
      const result = await client.getLeagues()
      setLeagues(result)
    } catch (loadError) {
      if (isUnauthorizedError(loadError)) {
        invalidateSession()
        return
      }
      setError(getApiErrorMessage(loadError))
    } finally {
      setIsLoading(false)
    }
  }, [client, invalidateSession, token])

  useEffect(() => {
    const storedToken = getStoredAuthToken()
    if (!storedToken) {
      setAuthState("unauthenticated")
      return
    }

    const parsedUserId = parseUserIdFromJwt(storedToken)
    if (!parsedUserId) {
      clearStoredAuthToken()
      setAuthState("unauthenticated")
      return
    }

    setToken(storedToken)
    setUserId(parsedUserId)
    setAuthState("ready")
  }, [])

  useEffect(() => {
    if (authState === "ready" && token) {
      void loadLeagues()
    }
  }, [authState, token, loadLeagues])

  useEffect(() => {
    if (authState === "unauthenticated") {
      window.location.replace("/login")
    }
  }, [authState])

  const handleCreateLeague: FormSubmitHandler = async (event) => {
    event.preventDefault()
    if (!token || !userId) {
      return
    }

    const trimmedName = name.trim()
    if (trimmedName.length === 0) {
      setError("League name is required.")
      return
    }

    setError(null)
    setSuccess(null)
    setIsLoading(true)
    client.setToken(token)

    try {
      const created = await client.createLeague({ name: trimmedName, user_id: userId })
      setLeagues((current) => [created, ...current])
      setName("")
      setSuccess(`Created league "${created.name}".`)
    } catch (createError) {
      if (isUnauthorizedError(createError)) {
        invalidateSession()
        return
      }
      setError(getApiErrorMessage(createError))
    } finally {
      setIsLoading(false)
    }
  }

  const handleJoinLeague: FormSubmitHandler = async (event) => {
    event.preventDefault()
    if (!token || !userId) {
      return
    }

    const parsedLeagueId = Number(joinLeagueId)
    if (!Number.isInteger(parsedLeagueId) || parsedLeagueId <= 0) {
      setError("Enter a valid league ID.")
      return
    }

    setError(null)
    setSuccess(null)
    setIsLoading(true)
    client.setToken(token)

    try {
      await client.joinLeague({ league_id: parsedLeagueId, user_id: userId })
      setJoinLeagueId("")
      setSuccess("Successfully joined league.")
      await loadLeagues()
    } catch (joinError) {
      if (isUnauthorizedError(joinError)) {
        invalidateSession()
        return
      }
      setError(getApiErrorMessage(joinError))
    } finally {
      setIsLoading(false)
    }
  }

  if (authState === "loading") {
    return <div className="rounded-2xl border border-slate-200 bg-white p-8 text-sm text-slate-600">Checking your session...</div>
  }

  if (authState === "unauthenticated") {
    return (
      <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-8 text-sm text-slate-700">
        Redirecting to login...
      </div>
    )
  }

  return (
    <section className="space-y-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs font-semibold tracking-[0.18em] text-slate-500 uppercase">League Hub</p>
          <h1 className="mt-2 text-3xl font-semibold text-slate-900">Your Leagues</h1>
          <p className="mt-2 text-sm text-slate-600">Create a league or join an existing one with its ID.</p>
        </div>
        <Button variant="outline" onClick={() => void loadLeagues()} disabled={isLoading}>
          Refresh
        </Button>
      </header>

      <div className="grid gap-4 md:grid-cols-2">
        <form onSubmit={handleCreateLeague} className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="text-base font-semibold text-slate-900">Create League</h2>
          <p className="mt-1 text-sm text-slate-600">Start a new competition as league creator.</p>
          <div className="mt-4 space-y-2">
            <label htmlFor="league-name" className="text-sm font-medium text-slate-700">League name</label>
            <Input
              id="league-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Weekend Warriors"
              disabled={isLoading}
              required
            />
          </div>
          <Button type="submit" className="mt-4 w-full" disabled={isLoading}>Create league</Button>
        </form>

        <form onSubmit={handleJoinLeague} className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="text-base font-semibold text-slate-900">Join League</h2>
          <p className="mt-1 text-sm text-slate-600">Enter an existing league ID to join.</p>
          <div className="mt-4 space-y-2">
            <label htmlFor="league-id" className="text-sm font-medium text-slate-700">League ID</label>
            <Input
              id="league-id"
              value={joinLeagueId}
              onChange={(event) => setJoinLeagueId(event.target.value)}
              inputMode="numeric"
              placeholder="123"
              disabled={isLoading}
              required
            />
          </div>
          <Button type="submit" className="mt-4 w-full" disabled={isLoading}>Join league</Button>
        </form>
      </div>

      {error ? <p className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</p> : null}
      {success ? <p className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</p> : null}

      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
        <header className="border-b border-slate-200 px-5 py-4">
          <h2 className="text-base font-semibold text-slate-900">League List</h2>
        </header>
        {isLoading && leagues.length === 0 ? (
          <p className="px-5 py-8 text-sm text-slate-600">Loading leagues...</p>
        ) : leagues.length === 0 ? (
          <p className="px-5 py-8 text-sm text-slate-600">No leagues yet. Create one to get started.</p>
        ) : (
          <ul className="divide-y divide-slate-200">
            {leagues.map((league) => (
              <li key={league.id} className="flex items-center justify-between gap-4 px-5 py-4">
                <div>
                  <p className="font-medium text-slate-900">{league.name}</p>
                  <p className="text-xs text-slate-500">Creator: {league.creator_username} • Members: {league.member_count}</p>
                </div>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-700">{league.member_count} members</span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  )
}
