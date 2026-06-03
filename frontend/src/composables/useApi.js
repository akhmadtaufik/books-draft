/**
 * Thin fetch wrapper for the Novel Drafting API.
 * Uses VITE_API_BASE_URL if defined, otherwise relative paths.
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

async function request(url, options = {}) {
  const fullUrl = url.startsWith('http') ? url : `${API_BASE_URL}${url}`
  const res = await fetch(fullUrl, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })

  if (!res.ok) {
    let message = `Request failed: ${res.status} ${res.statusText}`
    try {
      const body = await res.json()
      if (body.error) message = body.error
      else if (body.message) message = body.message
    } catch {
      // response body isn't JSON — keep the default message
    }
    throw new Error(message)
  }

  // 204 No Content → return null
  if (res.status === 204) return null

  return res.json()
}

export function get(url) {
  return request(url, { method: 'GET' })
}

export function post(url, body) {
  return request(url, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function put(url, body) {
  return request(url, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export function del(url) {
  return request(url, { method: 'DELETE' })
}
