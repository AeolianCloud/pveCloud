export interface RequestOptions extends RequestInit {
  bodyJson?: unknown
}

const baseURL = '/api'

const authStorageKey = 'pvecloud-web-auth'

interface APIErrorShape {
  error?: {
    code?: string
    message?: string
  }
}

export function readStoredToken(): string {
  if (typeof window === 'undefined') {
    return ''
  }

  return window.localStorage.getItem(authStorageKey) ?? ''
}

export function writeStoredToken(token: string) {
  if (typeof window === 'undefined') {
    return
  }

  if (token) {
    window.localStorage.setItem(authStorageKey, token)
    return
  }

  window.localStorage.removeItem(authStorageKey)
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  const token = readStoredToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  let body = options.body
  if (options.bodyJson !== undefined) {
    headers.set('Content-Type', 'application/json')
    body = JSON.stringify(options.bodyJson)
  }

  const response = await fetch(`${baseURL}${path}`, {
    ...options,
    headers,
    body,
  })

  if (!response.ok) {
    let message = `request failed: ${response.status}`

    try {
      const payload = (await response.json()) as APIErrorShape
      message = payload.error?.message ?? message
    } catch {
      // Ignore non-JSON error bodies.
    }

    if (response.status === 401) {
      writeStoredToken('')
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        window.location.assign('/login')
      }
    }

    throw new Error(message)
  }

  return response.json() as Promise<T>
}
