import axios from "axios";

// Use a relative base URL by default so that Vite's dev server proxy
// can forward API requests to the backend. This avoids hard-coding
// "localhost", which breaks when accessing the app from other devices.
export const apiBase = import.meta.env.VITE_API_BASE_URL || "";

const api = axios.create({
  baseURL: apiBase,
});

export interface ListParams {
  page?: number;
  pageSize?: number;
  q?: string;
  tags?: string[];
  nsfw?: "hide" | "show" | "only";
  sort?: "created_time" | "imported_at" | "file_name";
  order?: "asc" | "desc";
  rating?: number;
}

export async function listImages(params: ListParams) {
  const p = new URLSearchParams();
  p.set("page", String(params.page ?? 1));
  p.set("pageSize", String(params.pageSize ?? 50));
  if (params.q) p.set("q", params.q);
  if (params.tags && params.tags.length) p.set("tags", params.tags.join(","));
  p.set("nsfw", params.nsfw ?? "hide");
  p.set("sort", params.sort ?? "imported_at");
  p.set("order", params.order ?? "desc");
  if (params.rating !== undefined) p.set("rating", String(params.rating));
  const { data } = await api.get(`/api/images?${p.toString()}`);
  return data;
}

export async function getImage(id: number) {
  const { data } = await api.get(`/api/images/${id}`);
  return data;
}

export async function getLibraryPath(): Promise<string> {
  const { data } = await api.get("/api/settings/library_path");
  return data.value ?? "";
}

export async function setLibraryPath(path: string) {
  await api.put("/api/settings/library_path", { value: path });
}

export async function scanLibrary(root?: string) {
  const { data } = await api.post("/api/scan", root ? { root } : undefined);
  return data;
}

export async function deleteImage(
  id: number,
  mode: "trash" | "hard" = "trash",
) {
  await api.delete(`/api/images/${id}`, { params: { mode } });
}

export async function updateImageMetadata(id: number, metadata: any) {
  const { data } = await api.put(`/api/images/${id}/metadata`, metadata);
  return data;
}

export async function addTags(id: number, tags: string[]) {
  const { data } = await api.post(`/api/images/${id}/tags`, { tags });
  return data;
}

export async function removeTags(id: number, tags: string[]) {
  const { data } = await api.delete(`/api/images/${id}/tags`, {
    data: { tags },
  });
  return data;
}
