<template>
  <div class="row g-3">
    <div class="col-12 col-lg-3">
      <SidebarFilters
        v-model:q="q"
        v-model:nsfw="nsfw"
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
      <ImageGrid :images="items" />
      <Pager :page="page" :page-size="pageSize" :total="total" @change="onPage" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { listImages, scanLibrary, getLibraryPath } from '../api'
import SidebarFilters from '../components/SidebarFilters.vue'
import ImageGrid from '../components/ImageGrid.vue'
import Pager from '../components/Pager.vue'

const page = ref(1)
const pageSize = ref(50)
const q = ref('')
const nsfw = ref<'hide'|'show'|'only'>('hide')
const tags = ref<string[]>([])
const sort = ref<'created_time'|'imported_at'|'file_name'>('imported_at')
const order = ref<'asc'|'desc'>('desc')

const items = ref<any[]>([])
const total = ref(0)

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