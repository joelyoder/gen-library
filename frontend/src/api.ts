import axios from 'axios'

// falls back to 8081 if env missing
const base = import.meta.env.VITE_API_BASE || 'http://localhost:8081'
const api = axios.create({ baseURL: base })

export default api

export interface ListParams {
  page?: number
  pageSize?: number
  q?: string
  tags?: string[]
  nsfw?: 'hide'|'show'|'only'
  sort?: 'created_time'|'imported_at'|'file_name'
  order?: 'asc'|'desc'
}

export async function listImages(params: ListParams) {
  const p = new URLSearchParams()
  p.set('page', String(params.page ?? 1))
  p.set('pageSize', String(params.pageSize ?? 50))
  if (params.q) p.set('q', params.q)
  if (params.tags && params.tags.length) p.set('tags', params.tags.join(','))
  p.set('nsfw', params.nsfw ?? 'hide')
  p.set('sort', params.sort ?? 'imported_at')
  p.set('order', params.order ?? 'desc')
  const { data } = await api.get(`/api/images?${p.toString()}`)
  return data
}

export async function getImage(id: number) {
  const { data } = await api.get(`/api/images/${id}`)
  return data
}

export async function getLibraryFolder() {
  const { data } = await api.get('/api/settings/libraryFolder')
  return data
}

export async function setLibraryFolder(path: string) {
  const { data } = await api.put('/api/settings/libraryFolder', { path })
  return data
}

export async function importLibrary() {
  const { data } = await api.post('/api/settings/import')
  return data
}