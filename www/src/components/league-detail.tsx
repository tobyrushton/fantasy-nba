import { useEffect, useMemo, useState } from "react"

import { ApiError, Client, type LeagueResponse, type RosterResponse } from "@/client/api"
import { Button } from "@/components/ui/button"
import { clearStoredAuthToken, getBrowserApiUrl, getStoredAuthToken } from "@/lib/auth"

type AuthState = "loading" | "unauthenticated" | "ready"

type LeagueDetailProps = {
  leagueId: number
}

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

export function LeagueDetail({ leagueId }: LeagueDetailProps) {
  const client = useMemo(() => new Client(getBrowserApiUrl()), [])

  const [authState, setAuthState] = useState<AuthState>("loading")
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [league, setLeague] = useState<LeagueResponse | null>(null)
  const [rosters, setRosters] = useState<RosterResponse[]>([])

  useEffect(() => {
    const storedToken = getStoredAuthToken()
    if (!storedToken) {
      setAuthState("unauthenticated")
      return
    }

    setToken(storedToken)
    setAuthState("ready")
  }, [])

  useEffect(() => {
    if (authState === "unauthenticated") {
      window.location.replace("/login")
    }
  }, [authState])

  useEffect(() => {
    if (authState !== "ready" || !token) {
      return
    }

    const authToken = token

    let cancelled = false

    async function loadLeague() {
      setIsLoading(true)
      setError(null)
      client.setToken(authToken)

      try {
        const [leagueResponse, rostersResponse] = await Promise.all([
          client.getLeagueById(leagueId),
          client.getLeagueRosters(leagueId),
        ])

        if (!cancelled) {
          setLeague(leagueResponse)
          setRosters(rostersResponse)
        }
      } catch (loadError) {
        if (isUnauthorizedError(loadError)) {
          clearStoredAuthToken()
          if (!cancelled) {
            setToken(null)
            setAuthState("unauthenticated")
          }
          return
        }

        if (!cancelled) {
          setError(getApiErrorMessage(loadError))
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    void loadLeague()

    return () => {
      cancelled = true
    }
  }, [authState, client, leagueId, token])

  if (authState === "loading") {
    return <div className="rounded-2xl border border-slate-200 bg-white p-8 text-sm text-slate-600">Checking your session...</div>
  }

  if (authState === "unauthenticated") {
    return <div className="rounded-2xl border border-slate-200 bg-white p-8 text-sm text-slate-600">Redirecting to login...</div>
  }

  if (isLoading) {
    return <div className="rounded-2xl border border-slate-200 bg-white p-8 text-sm text-slate-600">Loading league...</div>
  }

  if (error) {
    return <div className="rounded-2xl border border-rose-200 bg-rose-50 p-8 text-sm text-rose-700">{error}</div>
  }

  if (!league) {
    return <div className="rounded-2xl border border-slate-200 bg-white p-8 text-sm text-slate-600">League not found.</div>
  }

  return (
    <section className="space-y-6">
      <header className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <p className="text-xs font-semibold tracking-[0.18em] text-slate-500 uppercase">League</p>
        <h1 className="mt-2 text-3xl font-semibold text-slate-900">{league.name}</h1>
        <p className="mt-2 text-sm text-slate-600">
          League ID: {league.id} • Creator: {league.creator_username} • Members: {league.member_count}
        </p>
      </header>

      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
        <header className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
          <h2 className="text-base font-semibold text-slate-900">Rosters</h2>
          <Button asChild size="sm" variant="outline">
            <a href="/leagues">Back to leagues</a>
          </Button>
        </header>

        {rosters.length === 0 ? (
          <p className="px-5 py-8 text-sm text-slate-600">No rosters have been submitted for this league yet.</p>
        ) : (
          <ul className="divide-y divide-slate-200">
            {rosters.map((roster) => (
              <li key={roster.user.id} className="px-5 py-4">
                <p className="font-medium text-slate-900">{roster.user.username}</p>
                <p className="mt-1 text-xs text-slate-500">{roster.players.length} players</p>
                <div className="mt-3 flex flex-wrap gap-2">
                  {roster.players.map((player) => (
                    <a
                      key={player.id}
                      href={`/players/${player.id}`}
                      className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs text-slate-700 hover:bg-slate-100"
                    >
                      {player.first_name} {player.last_name}
                    </a>
                  ))}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  )
}
