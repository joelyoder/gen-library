<template>
  <div class="col-12 col-lg-6">
    <h2 class="h4 mb-3">Settings</h2>
    <form @submit.prevent="save">
      <div class="mb-3">
        <label class="form-label">Library Folder</label>
        <input v-model="path" class="form-control" placeholder="D:\\AI\\library" />
      </div>
      <button class="btn btn-primary" type="submit">Save</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getLibraryPath, setLibraryPath } from '../api'

const path = ref('')

onMounted(async () => {
  path.value = await getLibraryPath()
})

async function save() {
  await setLibraryPath(path.value)
  alert('Saved')
}
</script>
