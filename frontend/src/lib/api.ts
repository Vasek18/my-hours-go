// Thin fetch wrapper around the Go JSON API.
// Always sends cookies (credentials: 'include') so the session works, and
// normalizes the backend's error shape into an ApiError.

const BASE_URL = '/api'

export interface ValidationErrors {
  [field: string]: string
}

export class ApiError extends Error {
  status: number
  errors: ValidationErrors

  constructor(status: number, message: string, errors: ValidationErrors = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.errors = errors
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(BASE_URL + path, {
    method,
    credentials: 'include',
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })

  // 204 / empty body
  const text = await res.text()
  const data = text ? JSON.parse(text) : {}

  if (!res.ok) {
    const message = (data && data.error) || 'Something went wrong.'
    const errors = (data && data.errors) || {}
    throw new ApiError(res.status, message, errors)
  }

  return data as T
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
}
