<template>
  <div class="row g-3">
      <div class="col-12 col-lg-3">
        <SidebarFilters
          v-model:q="q"
          v-model:tags="tags"
          v-model:sort="sort"
          v-model:order="order"
          @search="reload"
        />
      </div>
    <div class="col-12 col-lg-9">
      <div class="d-flex justify-content-end mb-2">
        <button class="btn btn-sm btn-secondary" @click="onScan">Scan Library</button>
      </div>
      <ImageGrid :images="items" @deleted="onDeleted" @metadata="onMetadata" />
      <Pager :page="page" :page-size="pageSize" :total="total" @change="onPage" />
    </div>
  </div>

  <div v-if="metadataOpen">
    <div class="modal fade show d-block" tabindex="-1">
      <div class="modal-dialog modal-fullscreen">
        <div class="modal-content bg-dark text-light">
          <div class="modal-header">
            <h5 class="modal-title">{{ selectedImage?.fileName }}</h5>
            <button type="button" class="btn-close btn-close-white" @click="closeMetadata"></button>
          </div>
          <div class="modal-body p-0">
            <div class="row g-0 h-100">
              <div class="col-md-8 d-flex align-items-center justify-content-center bg-black">
                <img
                  v-if="selectedImage"
                  :src="apiBase + '/api/images/' + selectedImage.id + '/file'"
                  class="img-fluid"
                  :alt="selectedImage.fileName"
                />
              </div>
              <div class="col-md-4 overflow-auto p-3" v-if="selectedImage">
                <div class="d-flex justify-content-end mb-3">
                  <button class="btn btn-outline-danger" @click="onDeleteSelected">Delete</button>
                </div>
                <MetadataPanel
                  v-if="metadataEditing"
                  :image="selectedImage"
                  @saved="onMetadataSaved"
                  @cancel="metadataEditing = false"
                />
                <div v-else>
                  <MetadataDisplay :image="selectedImage" @edit="metadataEditing = true" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="modal-backdrop fade show"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { listImages, scanLibrary, getLibraryPath, getImage, deleteImage } from '../api'
import SidebarFilters from '../components/SidebarFilters.vue'
import ImageGrid from '../components/ImageGrid.vue'
import Pager from '../components/Pager.vue'
import MetadataPanel from '../components/MetadataPanel.vue'
import MetadataDisplay from '../components/MetadataDisplay.vue'
import { nsfw } from '../nsfw'

const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081'

const page = ref(1)
const pageSize = ref(50)
const q = ref('')
const tags = ref<string[]>([])
const sort = ref<'created_time'|'imported_at'|'file_name'>('imported_at')
const order = ref<'asc'|'desc'>('desc')

const items = ref<any[]>([])
const total = ref(0)
const metadataOpen = ref(false)
const selectedImage = ref<any|null>(null)
const metadataEditing = ref(false)

function reload() {
  page.value = 1
}

function onPage(newPage: number) {
  page.value = newPage
}

async function onScan() {
  const root = await getLibraryPath()
  if (!root) {
    alert('Please set a library path in Settings first')
    return
  }
  await scanLibrary(root)
  reload()
}

function onDeleted(id: number) {
  items.value = items.value.filter(img => img.id !== id)
}

async function onMetadata(img: any) {
  selectedImage.value = await getImage(img.id)
  metadataOpen.value = true
  metadataEditing.value = false
}

function closeMetadata() {
  metadataOpen.value = false
  selectedImage.value = null
  metadataEditing.value = false
}

async function onMetadataSaved() {
  if (selectedImage.value) {
    const updated = await getImage(selectedImage.value.id)
    selectedImage.value = updated
    const idx = items.value.findIndex(i => i.id === updated.id)
    if (idx !== -1) items.value[idx] = updated
  }
  metadataEditing.value = false
}

async function onDeleteSelected() {
  if (!selectedImage.value) return
  if (!confirm('Delete this image?')) return
  await deleteImage(selectedImage.value.id)
  items.value = items.value.filter(img => img.id !== selectedImage.value?.id)
  closeMetadata()
}

watchEffect(async () => {
  const data = await listImages({
    page: page.value,
    pageSize: pageSize.value,
    q: q.value || undefined,
    tags: tags.value,
    nsfw: nsfw.value,
    sort: sort.value,
    order: order.value,
  })
  items.value = data.items
  total.value = data.total
})
</script>

