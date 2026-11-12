import { clearToken, getToken } from './session.js'

export async function api(path, { method = 'GET', body } = {}) {
  const token = getToken()
  const response = await fetch('/api' + path, {
    method,
    headers: {
      ...(body && { 'Content-Type': 'application/json' }),
      ...(token && { Authorization: 'Bearer ' + token }),
    },
    body: body && JSON.stringify(body),
  })

  // The token has expired or is no longer valid.
  if (response.status === 401 && token) {
    clearToken()
    window.location.assign('/signin')
  }
  if (response.status === 204) {
    return null
  }
  const data = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(data.error || 'Request failed')
  }
  return data
}
