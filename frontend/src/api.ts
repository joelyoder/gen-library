import axios from 'axios'

// Backend now defaults to port 8081, so point axios there
const api = axios.create({ baseURL: 'http://localhost:8081' })

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

export async function getLibraryPath(): Promise<string> {
  const { data } = await api.get('/api/settings/library_path')
  return data.value ?? ''
}

export async function setLibraryPath(path: string) {
  await api.put('/api/settings/library_path', { value: path })
}

export async function scanLibrary(root?: string) {
  const { data } = await api.post('/api/scan', root ? { root } : undefined)
  return data
}