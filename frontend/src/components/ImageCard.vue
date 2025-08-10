<template>
  <div class="card shadow-sm">
    <img v-if="image.thumbUrl" :src="image.thumbUrl" class="card-img-top" :alt="image.fileName" loading="lazy" />
    <div class="card-body p-2">
      <div class="d-flex justify-content-between align-items-center">
        <small class="text-truncate" style="max-width: 80%">{{ image.fileName }}</small>
        <span v-if="image.nsfw" class="badge text-bg-danger">NSFW</span>
      </div>
      <small v-if="image.modelName" class="text-muted">{{ image.modelName }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ image: any }>()
</script>