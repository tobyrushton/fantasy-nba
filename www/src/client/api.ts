
export type ApiMessageResponse = Record<string, string>

export interface AuthRequest {
    username: string
    password: string
}

export interface AuthTokenResponse {
    token: string
}

export interface CreateLeagueRequest {
    name: string
    user_id: number
}

export interface DeleteLeagueRequest {
    id: number
    user_id: number
}

export interface JoinLeagueRequest {
    league_id: number
    user_id: number
}

export interface CreateRosterRequest {
    league_id: number
    player_ids: number[]
    user_id: number
}

export interface UpdateRosterRequest {
    add_players: number[]
    league_id: number
    remove_players: number[]
    user_id: number
}

export interface LeagueResponse {
    creator_id: number
    creator_username: string
    id: number
    member_count: number
    name: string
}

export interface PlayerGameStatsResponse {
    assists: number
    blocks: number
    did_not_play: boolean
    game_date: string
    made_free_throws: number
    made_three_pointers: number
    player_id: number
    points: number
    rebounds: number
    steals: number
    turnovers: number
}

export interface PlayerResponse {
    first_name: string
    id: number
    last_name: string
    nba_id: string
    position: string
    team_name: string
    team_id: number
}

export interface PlayerStatsResponse {
    games: PlayerGameStatsResponse[]
    player: PlayerResponse
}

export interface UserResponse {
    id: number
    username: string
}

export interface RosterResponse {
    players: PlayerResponse[]
    user: UserResponse
}

export interface ListPlayersQuery {
    position?: string
    search?: string
}

export interface ClientOptions {
    defaultHeaders?: HeadersInit
    fetchFn?: typeof fetch
    token?: string
}

type RequestOptions = {
    auth?: boolean
    body?: unknown
    expectedStatus: number | number[]
    method: "GET" | "POST" | "PUT" | "DELETE"
}

export class ApiError<TError = ApiMessageResponse> extends Error {
    status: number
    body: TError | undefined

    constructor(status: number, message: string, body?: TError) {
        super(message)
        this.name = "ApiError"
        this.status = status
        this.body = body
    }
}

export class Client {
    private apiUrl: string
    private token?: string
    private defaultHeaders: HeadersInit
    private fetchFn: typeof fetch

    constructor(url: string, options: ClientOptions = {}) {
        this.apiUrl = url.replace(/\/+$/, "")
        this.token = options.token
        this.defaultHeaders = options.defaultHeaders ?? {}
        const configuredFetch = options.fetchFn
        this.fetchFn = configuredFetch
            ? (input, init) => configuredFetch(input, init)
            : (input, init) => globalThis.fetch(input, init)
    }

    setToken(token: string | undefined): void {
        this.token = token
    }

    clearToken(): void {
        this.token = undefined
    }

    async register(request: AuthRequest): Promise<ApiMessageResponse> {
        return this.request<ApiMessageResponse>("/auth/register", {
            body: request,
            expectedStatus: 201,
            method: "POST",
        })
    }

    async login(request: AuthRequest): Promise<AuthTokenResponse> {
        return this.request<AuthTokenResponse>("/auth/login", {
            body: request,
            expectedStatus: 200,
            method: "POST",
        })
    }

    async getPlayers(query?: ListPlayersQuery): Promise<PlayerResponse[]> {
        return this.request<PlayerResponse[]>(
            this.withQuery("/players/", {
                position: query?.position,
                search: query?.search,
            }),
            {
            expectedStatus: 200,
            method: "GET",
            }
        )
    }

    async getPlayer(id: number): Promise<PlayerStatsResponse> {
        return this.request<PlayerStatsResponse>(`/players/${id}`, {
            expectedStatus: 200,
            method: "GET",
        })
    }

    async getLeagues(): Promise<LeagueResponse[]> {
        return this.request<LeagueResponse[]>("/league/", {
            auth: true,
            expectedStatus: 200,
            method: "GET",
        })
    }

    async getLeagueById(id: number): Promise<LeagueResponse> {
        return this.request<LeagueResponse>(`/league/${id}`, {
            auth: true,
            expectedStatus: 200,
            method: "GET",
        })
    }

    async createLeague(request: CreateLeagueRequest): Promise<LeagueResponse> {
        return this.request<LeagueResponse>("/league/", {
            auth: true,
            body: request,
            expectedStatus: 201,
            method: "POST",
        })
    }

    async deleteLeague(request: DeleteLeagueRequest): Promise<ApiMessageResponse> {
        return this.request<ApiMessageResponse>("/league/", {
            auth: true,
            body: request,
            expectedStatus: 200,
            method: "DELETE",
        })
    }

    async joinLeague(request: JoinLeagueRequest): Promise<ApiMessageResponse> {
        return this.request<ApiMessageResponse>("/league/join", {
            auth: true,
            body: request,
            expectedStatus: 200,
            method: "POST",
        })
    }

    async createRoster(request: CreateRosterRequest): Promise<ApiMessageResponse> {
        return this.request<ApiMessageResponse>("/league/roster", {
            auth: true,
            body: request,
            expectedStatus: 200,
            method: "POST",
        })
    }

    async updateRoster(request: UpdateRosterRequest): Promise<ApiMessageResponse> {
        return this.request<ApiMessageResponse>("/league/roster", {
            auth: true,
            body: request,
            expectedStatus: 200,
            method: "PUT",
        })
    }

    async getLeagueRosters(id: number): Promise<RosterResponse[]> {
        return this.request<RosterResponse[]>(`/league/${id}/rosters`, {
            auth: true,
            expectedStatus: 200,
            method: "GET",
        })
    }

    private withQuery(path: string, query?: Record<string, string | undefined>): string {
        if (!query) {
            return path
        }

        const params = new URLSearchParams()
        for (const [key, value] of Object.entries(query)) {
            if (value !== undefined && value.length > 0) {
                params.set(key, value)
            }
        }

        const encodedQuery = params.toString()
        return encodedQuery.length > 0 ? `${path}?${encodedQuery}` : path
    }

    private async request<TResponse>(path: string, options: RequestOptions): Promise<TResponse> {
        const expected = Array.isArray(options.expectedStatus)
            ? options.expectedStatus
            : [options.expectedStatus]

        const headers = new Headers(this.defaultHeaders)
        if (options.body !== undefined) {
            headers.set("Content-Type", "application/json")
        }

        if (options.auth) {
            if (!this.token) {
                throw new ApiError(0, "Missing auth token for authenticated request")
            }
            headers.set("Authorization", `Bearer ${this.token}`)
        }

        const response = await this.fetchFn(`${this.apiUrl}${path}`, {
            body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
            headers,
            method: options.method,
        })

        const responseBody = await this.parseJson(response)

        if (!expected.includes(response.status)) {
            throw new ApiError(response.status, `Request failed with status ${response.status}`, responseBody)
        }

        return responseBody as TResponse
    }

    private async parseJson(response: Response): Promise<unknown> {
        if (response.status === 204) {
            return undefined
        }

        const contentType = response.headers.get("content-type")
        if (!contentType || !contentType.toLowerCase().includes("application/json")) {
            return undefined
        }

        return response.json()
    }
}