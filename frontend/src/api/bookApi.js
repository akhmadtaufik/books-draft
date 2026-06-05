import { get, post, put, del } from '../composables/useApi.js'

export const bookApi = {
  getAll: () => get('/api/books'),
  getPreview: (id) => get(`/api/books/${id}/preview`),
  getChapters: (id) => get(`/api/books/${id}/chapters`),
  create: (payload) => post('/api/books', payload),
  update: (id, payload) => put(`/api/books/${id}`, payload),
  delete: (id) => del(`/api/books/${id}`)
}
