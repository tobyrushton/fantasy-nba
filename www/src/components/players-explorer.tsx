import { useMemo, useState } from "react"

import type { PlayerResponse } from "@/client/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

type PlayersExplorerProps = {
  players: PlayerResponse[]
}

function fullName(player: PlayerResponse): string {
  return `${player.first_name} ${player.last_name}`
}

function normalizeForSearch(value: string): string {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLocaleLowerCase()
    .replace(/[^a-z0-9\s]/g, " ")
    .replace(/\s+/g, " ")
    .trim()
}

export function PlayersExplorer({ players }: PlayersExplorerProps) {
  const [search, setSearch] = useState("")
  const [selectedPosition, setSelectedPosition] = useState("all")

  const positionOptions = useMemo(() => {
    const unique = new Set(
      players.map((player) => player.position?.trim()).filter((position): position is string => Boolean(position))
    )

    return Array.from(unique).sort((a, b) => a.localeCompare(b))
  }, [players])

  const filteredPlayers = useMemo(() => {
    const normalizedQuery = normalizeForSearch(search)
    const queryTerms = normalizedQuery.length > 0 ? normalizedQuery.split(" ") : []

    return players.filter((player) => {
      const name = normalizeForSearch(fullName(player))
      const position = normalizeForSearch(player.position ?? "")
      const nbaId = normalizeForSearch(player.nba_id)
      const searchableText = `${name} ${position} ${nbaId}`

      const matchesSearch =
        queryTerms.length === 0 || queryTerms.every((term) => searchableText.includes(term))
      const matchesPosition =
        selectedPosition === "all" || position === normalizeForSearch(selectedPosition)

      return matchesSearch && matchesPosition
    })
  }, [players, search, selectedPosition])

  return (
    <section className="relative mx-auto w-full max-w-6xl overflow-hidden rounded-3xl border border-slate-200 bg-white/95 p-6 shadow-xl md:p-10">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_right,#dbeafe,transparent_42%),radial-gradient(circle_at_bottom_left,#fef3c7,transparent_48%)]" />
      <div className="relative space-y-6">
        <header className="space-y-2">
          <p className="text-xs font-semibold tracking-[0.18em] text-muted-foreground uppercase">
            Fantasy NBA
          </p>
          <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">Players Directory</h1>
          <p className="max-w-2xl text-sm text-muted-foreground md:text-base">
            Search by player name, position, or NBA ID and browse the current roster pool.
          </p>
        </header>

        <div className="flex flex-col gap-3 md:flex-row md:items-center">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search players or NBA IDs"
            className="h-11 border-slate-300 bg-white text-slate-900 placeholder:text-slate-500"
          />
          <Select value={selectedPosition} onValueChange={setSelectedPosition}>
            <SelectTrigger
              className="h-11 min-w-44 border-slate-300 bg-white text-slate-900"
              aria-label="Filter players by position"
            >
              <SelectValue placeholder="All positions" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All positions</SelectItem>
              {positionOptions.map((position) => (
                <SelectItem key={position} value={position.toLowerCase()}>
                  {position}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            type="button"
            variant="outline"
            className="h-11 px-4"
            onClick={() => {
              setSearch("")
              setSelectedPosition("all")
            }}
            disabled={search.length === 0 && selectedPosition === "all"}
          >
            Clear
          </Button>
        </div>

        <div className="flex items-center justify-between text-sm text-slate-600">
          <span>
            Showing <strong className="text-slate-900">{filteredPlayers.length}</strong> of{" "}
            <strong className="text-slate-900">{players.length}</strong>
          </span>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {filteredPlayers.map((player) => (
            <article
              key={player.id}
              className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition-transform duration-200 hover:-translate-y-0.5 hover:shadow-md"
            >
              <h2 className="text-base font-semibold text-slate-900">{fullName(player)}</h2>
              <p className="mt-1 text-xs text-slate-600">NBA ID: {player.nba_id}</p>
              <div className="mt-4 flex items-center gap-2 text-xs">
                <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-1 font-medium text-slate-700">
                  Position: {player.position || "N/A"}
                </span>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-1 font-medium text-slate-700">
                  Team: {player.team_id}
                </span>
              </div>
              <Button asChild size="sm" className="mt-4 w-full">
                <a href={`/players/${player.id}`}>View stats</a>
              </Button>
            </article>
          ))}
        </div>

        {filteredPlayers.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-10 text-center text-sm text-slate-600">
            No players match your search.
          </div>
        ) : null}
      </div>
    </section>
  )
}