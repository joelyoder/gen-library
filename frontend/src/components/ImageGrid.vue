<template>
  <div
    class="grid"
    :style="{ gridTemplateColumns: `repeat(${columnCount}, 1fr)` }"
  >
    <div v-for="img in images" :key="img.id">
      <ImageCard
        :image="img"
        @deleted="emit('deleted', $event)"
        @metadata="emit('metadata', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import ImageCard from './ImageCard.vue'

defineProps<{ images: any[] }>()
const emit = defineEmits(['deleted', 'metadata'])

const columnCount = Math.max(1, Math.min(5, Math.floor(window.innerWidth / 320)))
</script>

<style scoped>
.grid {
  display: grid;
  gap: 1rem;
}
</style>

