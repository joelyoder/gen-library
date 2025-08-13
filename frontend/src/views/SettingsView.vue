<template>
  <div>
    <div class="mb-3">
      <label class="form-label">Library Folder</label>
      <input v-model="path" type="text" class="form-control" />
    </div>
    <button class="btn btn-primary me-2" @click="save">Save</button>
    <button class="btn btn-secondary" @click="importNow">Import Now</button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getLibraryFolder, setLibraryFolder, importLibrary } from '../api'

const path = ref('')

onMounted(async () => {
  const data = await getLibraryFolder()
  path.value = data.path || ''
})

async function save() {
  await setLibraryFolder(path.value)
  alert('Saved')
}

async function importNow() {
  const res = await importLibrary()
  alert(`Added ${res.added}, updated ${res.updated}`)
  window.dispatchEvent(new Event('library-updated'))
}
</script>

