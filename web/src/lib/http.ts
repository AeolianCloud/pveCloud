export interface RequestOptions extends RequestInit {
  bodyJson?: unknown
}

const baseURL = '/api'

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')

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
    throw new Error(`request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}
