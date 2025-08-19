<template>
  <div class="d-flex" style="gap: 1rem">
    <div v-for="(col, cIdx) in columns" :key="cIdx" class="flex-fill">
      <div v-for="img in col" :key="img.id" class="mb-3">
        <ImageCard
          :image="img"
          @deleted="emit('deleted', $event)"
          @metadata="emit('metadata', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ImageCard from './ImageCard.vue'

const props = defineProps<{ images: any[] }>()
const emit = defineEmits(['deleted', 'metadata'])

const columnCount = Math.max(1, Math.min(5, Math.floor(window.innerWidth / 320)))
const columns = computed(() => {
  const rows = Math.ceil(props.images.length / columnCount)
  return Array.from({ length: columnCount }, (_, i) =>
    props.images.slice(i * rows, (i + 1) * rows)
  )
})
</script>

<style scoped>
</style>

