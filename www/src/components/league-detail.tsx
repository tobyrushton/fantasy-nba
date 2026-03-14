import { useEffect, useMemo, useState } from "react"

import { ApiError, Client, type LeagueResponse, type PlayerResponse, type RosterResponse } from "@/client/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
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

function normalizeForSearch(value: string): string {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9\s]/g, " ")
    .replace(/\s+/g, " ")
    .trim()
}

export function LeagueDetail({ leagueId }: LeagueDetailProps) {
  const client = useMemo(() => new Client(getBrowserApiUrl()), [])

  const [authState, setAuthState] = useState<AuthState>("loading")
  const [token, setToken] = useState<string | null>(null)
  const [userId, setUserId] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmittingRoster, setIsSubmittingRoster] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [league, setLeague] = useState<LeagueResponse | null>(null)
  const [rosters, setRosters] = useState<RosterResponse[]>([])
  const [players, setPlayers] = useState<PlayerResponse[]>([])
  const [playerSearch, setPlayerSearch] = useState("")
  const [selectedPlayerIds, setSelectedPlayerIds] = useState<number[]>([])
  const [removeSearch, setRemoveSearch] = useState("")
  const [addSearch, setAddSearch] = useState("")
  const [selectedRemovePlayerIds, setSelectedRemovePlayerIds] = useState<number[]>([])
  const [selectedAddPlayerIds, setSelectedAddPlayerIds] = useState<number[]>([])

  function parseUserID(tokenValue: string): number | null {
    const parts = tokenValue.split(".")
    if (parts.length < 2) {
      return null
    }

    try {
      const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/")
      const padded = payload.padEnd(payload.length + ((4 - (payload.length % 4)) % 4), "=")
      const decoded = window.atob(padded)
      const claims = JSON.parse(decoded) as Record<string, unknown>
      const claimUserID = claims.user_id

      if (typeof claimUserID === "number" && Number.isFinite(claimUserID)) {
        return claimUserID
      }

      return null
    } catch {
      return null
    }
  }

  useEffect(() => {
    const storedToken = getStoredAuthToken()
    if (!storedToken) {
      setAuthState("unauthenticated")
      return
    }

    const parsedUserID = parseUserID(storedToken)
    if (!parsedUserID) {
      clearStoredAuthToken()
      setAuthState("unauthenticated")
      return
    }

    setToken(storedToken)
    setUserId(parsedUserID)
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
        const [leagueResponse, rostersResponse, playersResponse] = await Promise.all([
          client.getLeagueById(leagueId),
          client.getLeagueRosters(leagueId),
          client.getPlayers(),
        ])

        if (!cancelled) {
          setLeague(leagueResponse)
          setRosters(rostersResponse)
          setPlayers(playersResponse)
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

  const filteredPlayers = useMemo(() => {
    const query = normalizeForSearch(playerSearch)
    if (query.length === 0) {
      return players
    }

    return players.filter((player) => {
      const fullName = normalizeForSearch(`${player.first_name} ${player.last_name}`)
      const position = normalizeForSearch(player.position ?? "")
      return fullName.includes(query) || position.includes(query)
    })
  }, [playerSearch, players])

  const currentUserRoster = useMemo(() => {
    if (!userId) {
      return undefined
    }

    return rosters.find((roster) => roster.user.id === userId)
  }, [rosters, userId])
  const isLeagueMember = Boolean(currentUserRoster)
  const rosterPlayers = useMemo(() => currentUserRoster?.players ?? [], [currentUserRoster])
  const hasTeam = rosterPlayers.length > 0

  const filteredRosterPlayersForRemoval = useMemo(() => {
    const query = normalizeForSearch(removeSearch)
    if (query.length === 0) {
      return rosterPlayers
    }

    return rosterPlayers.filter((player) => {
      const fullName = normalizeForSearch(`${player.first_name} ${player.last_name}`)
      const position = normalizeForSearch(player.position ?? "")
      return fullName.includes(query) || position.includes(query)
    })
  }, [removeSearch, rosterPlayers])

  const addablePlayers = useMemo(() => {
    const rosterIDs = new Set(rosterPlayers.map((player) => player.id))
    return players.filter((player) => !rosterIDs.has(player.id))
  }, [players, rosterPlayers])

  const filteredAddablePlayers = useMemo(() => {
    const query = normalizeForSearch(addSearch)
    if (query.length === 0) {
      return addablePlayers
    }

    return addablePlayers.filter((player) => {
      const fullName = normalizeForSearch(`${player.first_name} ${player.last_name}`)
      const position = normalizeForSearch(player.position ?? "")
      return fullName.includes(query) || position.includes(query)
    })
  }, [addSearch, addablePlayers])

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

  async function handleCreateTeam() {
    if (!token || !userId) {
      return
    }

    if (selectedPlayerIds.length !== 10) {
      setError("Select exactly 10 players to create your team.")
      return
    }

    setIsSubmittingRoster(true)
    setError(null)
    setSuccess(null)
    client.setToken(token)

    try {
      await client.createRoster({
        league_id: leagueId,
        player_ids: selectedPlayerIds,
        user_id: userId,
      })

      const refreshedRosters = await client.getLeagueRosters(leagueId)
      setRosters(refreshedRosters)
      setSelectedPlayerIds([])
      setSuccess("Team created successfully.")
    } catch (createError) {
      if (isUnauthorizedError(createError)) {
        clearStoredAuthToken()
        setToken(null)
        setAuthState("unauthenticated")
        return
      }

      setError(getApiErrorMessage(createError))
    } finally {
      setIsSubmittingRoster(false)
    }
  }

  async function handleUpdateTeam() {
    if (!token || !userId) {
      return
    }

    if (selectedRemovePlayerIds.length === 0) {
      setError("Select at least one player to remove.")
      return
    }

    if (selectedRemovePlayerIds.length !== selectedAddPlayerIds.length) {
      setError("You must add the same number of players you remove.")
      return
    }

    setIsSubmittingRoster(true)
    setError(null)
    setSuccess(null)
    client.setToken(token)

    try {
      await client.updateRoster({
        add_players: selectedAddPlayerIds,
        league_id: leagueId,
        remove_players: selectedRemovePlayerIds,
        user_id: userId,
      })

      const refreshedRosters = await client.getLeagueRosters(leagueId)
      setRosters(refreshedRosters)
      setSelectedRemovePlayerIds([])
      setSelectedAddPlayerIds([])
      setSuccess("Team updated successfully.")
    } catch (updateError) {
      if (isUnauthorizedError(updateError)) {
        clearStoredAuthToken()
        setToken(null)
        setAuthState("unauthenticated")
        return
      }

      setError(getApiErrorMessage(updateError))
    } finally {
      setIsSubmittingRoster(false)
    }
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
            {rosters.map((roster) => {
              const rosterPlayers = roster.players ?? []

              return (
                <li key={roster.user.id} className="px-5 py-4">
                  <p className="font-medium text-slate-900">{roster.user.username}</p>
                  <p className="mt-1 text-xs text-slate-500">{rosterPlayers.length} players</p>
                  <div className="mt-3 flex flex-wrap gap-2">
                    {rosterPlayers.map((player) => (
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
              )
            })}
          </ul>
        )}
      </section>

      {isLeagueMember && !hasTeam ? (
        <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
          <header className="border-b border-slate-200 px-5 py-4">
            <h2 className="text-base font-semibold text-slate-900">Create Your Team</h2>
            <p className="mt-1 text-sm text-slate-600">Pick exactly 10 players.</p>
          </header>

          <div className="px-5 py-4">
            <p className="mb-3 text-sm text-slate-600">
              Selected: <span className="font-semibold text-slate-900">{selectedPlayerIds.length}</span>/10
            </p>
            <div className="mb-3 space-y-1">
              <label htmlFor="roster-player-search" className="text-sm font-medium text-slate-700">
                Search players
              </label>
              <Input
                id="roster-player-search"
                value={playerSearch}
                onChange={(event) => setPlayerSearch(event.target.value)}
                placeholder="Search by player name or position"
                className="h-10 border-slate-300 bg-white"
              />
            </div>
            <div className="max-h-64 overflow-y-auto rounded-xl border border-slate-200 p-2">
              <div className="grid gap-2 md:grid-cols-2">
                {filteredPlayers.map((player) => {
                  const isSelected = selectedPlayerIds.includes(player.id)

                  return (
                    <div
                      key={player.id}
                      className={`rounded-lg border px-3 py-2 text-sm transition-colors ${
                        isSelected
                          ? "border-emerald-300 bg-emerald-50"
                          : "border-slate-200 bg-white"
                      }`}
                    >
                      <button
                        type="button"
                        className="w-full text-left text-slate-800"
                        onClick={() => {
                          setSelectedPlayerIds((current) => {
                            if (current.includes(player.id)) {
                              return current.filter((id) => id !== player.id)
                            }

                            if (current.length >= 10) {
                              return current
                            }

                            return [...current, player.id]
                          })
                        }}
                      >
                        <span className="font-medium">{player.first_name} {player.last_name}</span>
                      </button>
                      <div className="mt-1 flex items-center justify-between gap-2 text-xs text-slate-600">
                        <span>Position: {player.position || "N/A"}</span>
                        <a href={`/players/${player.id}`} className="font-medium text-slate-800 underline underline-offset-2">
                          View full stats
                        </a>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>

            <div className="mt-4 flex gap-2">
              <Button type="button" onClick={handleCreateTeam} disabled={isSubmittingRoster || selectedPlayerIds.length !== 10}>
                {isSubmittingRoster ? "Creating team..." : "Create team"}
              </Button>
              <Button type="button" variant="outline" onClick={() => setSelectedPlayerIds([])} disabled={isSubmittingRoster || selectedPlayerIds.length === 0}>
                Clear selection
              </Button>
            </div>
          </div>
        </section>
      ) : null}

      {isLeagueMember && hasTeam ? (
        <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
          <header className="border-b border-slate-200 px-5 py-4">
            <h2 className="text-base font-semibold text-slate-900">Update Your Team</h2>
            <p className="mt-1 text-sm text-slate-600">Remove and add the same number of players.</p>
          </header>

          <div className="grid gap-4 px-5 py-4 md:grid-cols-2">
            <div>
              <p className="mb-2 text-sm font-medium text-slate-800">Remove from roster</p>
              <Input
                value={removeSearch}
                onChange={(event) => setRemoveSearch(event.target.value)}
                placeholder="Search your roster"
                className="mb-2 h-10 border-slate-300 bg-white"
              />
              <div className="max-h-64 overflow-y-auto rounded-xl border border-slate-200 p-2">
                <div className="space-y-2">
                  {filteredRosterPlayersForRemoval.map((player) => {
                    const isSelected = selectedRemovePlayerIds.includes(player.id)
                    return (
                      <button
                        key={player.id}
                        type="button"
                        className={`w-full rounded-lg border px-3 py-2 text-left text-sm ${
                          isSelected
                            ? "border-rose-300 bg-rose-50 text-rose-800"
                            : "border-slate-200 bg-white text-slate-700"
                        }`}
                        onClick={() => {
                          setSelectedRemovePlayerIds((current) => {
                            if (current.includes(player.id)) {
                              return current.filter((id) => id !== player.id)
                            }
                            return [...current, player.id]
                          })
                        }}
                      >
                        {player.first_name} {player.last_name}
                        <span className="ml-2 text-xs text-slate-500">({player.position || "N/A"})</span>
                      </button>
                    )
                  })}
                </div>
              </div>
            </div>

            <div>
              <p className="mb-2 text-sm font-medium text-slate-800">Add to roster</p>
              <Input
                value={addSearch}
                onChange={(event) => setAddSearch(event.target.value)}
                placeholder="Search available players"
                className="mb-2 h-10 border-slate-300 bg-white"
              />
              <div className="max-h-64 overflow-y-auto rounded-xl border border-slate-200 p-2">
                <div className="space-y-2">
                  {filteredAddablePlayers.map((player) => {
                    const isSelected = selectedAddPlayerIds.includes(player.id)
                    return (
                      <div key={player.id} className={`rounded-lg border px-3 py-2 text-sm ${
                        isSelected
                          ? "border-emerald-300 bg-emerald-50"
                          : "border-slate-200 bg-white"
                      }`}>
                        <button
                          type="button"
                          className="w-full text-left text-slate-800"
                          onClick={() => {
                            setSelectedAddPlayerIds((current) => {
                              if (current.includes(player.id)) {
                                return current.filter((id) => id !== player.id)
                              }
                              return [...current, player.id]
                            })
                          }}
                        >
                          {player.first_name} {player.last_name}
                          <span className="ml-2 text-xs text-slate-500">({player.position || "N/A"})</span>
                        </button>
                        <div className="mt-1 text-right">
                          <a href={`/players/${player.id}`} className="text-xs font-medium text-slate-800 underline underline-offset-2">
                            View full stats
                          </a>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>
          </div>

          <div className="border-t border-slate-200 px-5 py-4">
            <p className="mb-3 text-sm text-slate-600">
              Removing <span className="font-semibold text-slate-900">{selectedRemovePlayerIds.length}</span> and adding <span className="font-semibold text-slate-900">{selectedAddPlayerIds.length}</span>
            </p>
            <div className="flex gap-2">
              <Button
                type="button"
                onClick={handleUpdateTeam}
                disabled={
                  isSubmittingRoster ||
                  selectedRemovePlayerIds.length === 0 ||
                  selectedRemovePlayerIds.length !== selectedAddPlayerIds.length
                }
              >
                {isSubmittingRoster ? "Updating team..." : "Update team"}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setSelectedRemovePlayerIds([])
                  setSelectedAddPlayerIds([])
                }}
                disabled={isSubmittingRoster || (selectedRemovePlayerIds.length === 0 && selectedAddPlayerIds.length === 0)}
              >
                Clear changes
              </Button>
            </div>
          </div>
        </section>
      ) : null}

      {!isLeagueMember ? (
        <p className="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          Join this league before creating a team.
        </p>
      ) : null}

      {success ? <p className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</p> : null}
    </section>
  )
}
