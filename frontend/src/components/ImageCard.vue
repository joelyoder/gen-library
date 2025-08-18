<template>
  <div class="card shadow-sm">
    <img
      v-if="image.thumbUrl"
      :src="apiBase + image.thumbUrl"
      class="card-img-top"
      :alt="image.fileName"
      loading="lazy"
    />
    <div class="card-body p-2">
      <div class="d-flex justify-content-between align-items-center">
        <small class="text-truncate" style="max-width: 80%">{{ image.fileName }}</small>
        <div class="d-flex align-items-center gap-1">
          <span v-if="image.nsfw" class="badge text-bg-danger">NSFW</span>
          <button class="btn btn-sm btn-outline-primary" @click="onMetadata">Metadata</button>
          <button class="btn btn-sm btn-outline-danger" @click="onDelete">Delete</button>
        </div>
      </div>
      <small v-if="image.modelName" class="text-muted">{{ image.modelName }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import { deleteImage } from '../api'

const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081'
const props = defineProps<{ image: any }>()
const emit = defineEmits(['deleted', 'metadata'])

async function onDelete() {
  if (!confirm('Delete this image?')) return
  await deleteImage(props.image.id)
  emit('deleted', props.image.id)
}

function onMetadata() {
  emit('metadata', props.image)
}
</script>
