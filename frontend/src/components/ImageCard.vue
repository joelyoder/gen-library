<template>
  <div class="card shadow-sm">
    <div class="position-relative">
      <img
        v-if="image.thumbUrl"
        :src="apiBase + image.thumbUrl"
        class="card-img-top"
        :alt="image.fileName"
        loading="lazy"
        @click="onView"
      />
      <button
        class="btn btn-sm"
        :class="image.nsfw ? 'btn-danger' : 'btn-outline-danger'"
        style="position: absolute; top: 0.5rem; right: 0.5rem"
        @click.stop="onToggleNSFW"
      >
        <i :class="['bi', image.nsfw ? 'bi-eye-slash-fill' : 'bi-eye-fill']"></i>
      </button>
    </div>
    <div class="card-body p-2">
      <div class="d-flex justify-content-between align-items-center">
        <i
          v-if="image.favorite"
          class="bi bi-star-fill text-warning"
        ></i>
        <div class="d-flex align-items-center">
          <button class="btn btn-sm btn-outline-danger" @click="onDelete">
            <i class="bi bi-trash"></i>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

  <script setup lang="ts">
  import { deleteImage, apiBase, updateImageMetadata } from '../api'
  const props = defineProps<{ image: any }>()
  const emit = defineEmits(['deleted', 'metadata', 'nsfw-changed'])

  async function onDelete() {
    if (!confirm('Delete this image?')) return
    await deleteImage(props.image.id)
    emit('deleted', props.image.id)
  }

  function onView() {
    emit('metadata', props.image)
  }

  async function onToggleNSFW() {
    const newVal = !props.image.nsfw
    try {
      const updated = await updateImageMetadata(props.image.id, { nsfw: newVal })
      ;(props.image as any).nsfw = updated.nsfw
      emit('nsfw-changed', { id: props.image.id, nsfw: updated.nsfw })
    } catch (err) {
      console.error('Failed to update NSFW status', err)
    }
  }
</script>
