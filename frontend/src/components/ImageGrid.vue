<template>
  <div class="masonry" :style="{ columnCount }">
    <div v-for="img in images" :key="img.id" class="mb-3 break-inside-avoid">
      <ImageCard :image="img" />
    </div>
  </div>
</template>

<script setup lang="ts">
import ImageCard from './ImageCard.vue'

const props = defineProps<{ images: any[] }>()

const columnCount = Math.max(1, Math.min(5, Math.floor(window.innerWidth / 320)))
</script>

<style scoped>
.masonry { column-gap: 1rem; }
.break-inside-avoid { break-inside: avoid; }
</style>