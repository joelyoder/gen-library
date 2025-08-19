<template>
  <div class="card shadow-sm">
    <img
      v-if="image.thumbUrl"
      :src="apiBase + image.thumbUrl"
      class="card-img-top"
      :alt="image.fileName"
      loading="lazy"
      @click="onView"
    />
    <div class="card-body p-2">
      <div class="d-flex justify-content-end align-items-center gap-1">
        <span v-if="image.nsfw" class="badge text-bg-danger">NSFW</span>
        <button class="btn btn-sm btn-outline-danger" @click="onDelete">
          <i class="bi bi-trash"></i>
        </button>
      </div>
    </div>
  </div>
</template>

  <script setup lang="ts">
  import { deleteImage, apiBase } from '../api'
  const props = defineProps<{ image: any }>()
  const emit = defineEmits(['deleted', 'metadata'])

  async function onDelete() {
    if (!confirm('Delete this image?')) return
    await deleteImage(props.image.id)
    emit('deleted', props.image.id)
  }

  function onView() {
    emit('metadata', props.image)
  }
</script>
