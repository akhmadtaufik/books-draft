import { get, post, put, del } from '../composables/useApi.js'

export const chapterApi = {
  getById: (id) => get(`/api/chapters/${id}`),
  update: (id, payload) => put(`/api/chapters/${id}`, payload),
  saveSnapshot: (id, type) => post(`/api/chapters/${id}/versions`, { snapshot_type: type })
}
