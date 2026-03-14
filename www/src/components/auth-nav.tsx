import { useState } from "react"

import { Button } from "@/components/ui/button"
import { clearStoredAuthToken, getStoredAuthToken } from "@/lib/auth"

export function AuthNav() {
  const [hasToken, setHasToken] = useState(() => Boolean(getStoredAuthToken()))

  if (hasToken) {
    return (
      <div className="flex items-center gap-2">
        <Button asChild variant="outline" size="sm">
          <a href="/">Players</a>
        </Button>
        <Button asChild variant="outline" size="sm">
          <a href="/leagues">Leagues</a>
        </Button>
        <Button
          size="sm"
          onClick={() => {
            clearStoredAuthToken()
            setHasToken(false)
          }}
        >
          Log out
        </Button>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      <Button asChild variant="outline" size="sm">
        <a href="/login">Log in</a>
      </Button>
      <Button asChild size="sm">
        <a href="/register">Sign up</a>
      </Button>
    </div>
  )
}
