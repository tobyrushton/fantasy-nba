export const AUTH_TOKEN_STORAGE_KEY = "fantasy_nba_token"

export function getStoredAuthToken(): string | null {
  if (typeof window === "undefined") {
    return null
  }

  return window.localStorage.getItem(AUTH_TOKEN_STORAGE_KEY)
}

export function setStoredAuthToken(token: string): void {
  if (typeof window === "undefined") {
    return
  }

  window.localStorage.setItem(AUTH_TOKEN_STORAGE_KEY, token)
}

export function clearStoredAuthToken(): void {
  if (typeof window === "undefined") {
    return
  }

  window.localStorage.removeItem(AUTH_TOKEN_STORAGE_KEY)
}

export function getBrowserApiUrl(): string {
  const configured = import.meta.env.PUBLIC_API_URL
  if (configured && configured.length > 0) {
    return configured
  }

  return "http://127.0.0.1:8080"
}
