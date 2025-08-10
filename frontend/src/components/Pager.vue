<template>
  <nav class="d-flex justify-content-between align-items-center mt-3" aria-label="pager">
    <button class="btn btn-outline-secondary" :disabled="page <= 1" @click="$emit('change', page - 1)">Prev</button>
    <div>
      <span class="me-2">Page {{ page }}</span>
      <span class="text-muted">• {{ total }} items</span>
    </div>
    <button class="btn btn-outline-secondary" :disabled="page >= maxPage" @click="$emit('change', page + 1)">Next</button>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ page: number; pageSize: number; total: number }>()
const maxPage = computed(() => Math.max(1, Math.ceil(props.total / (props.pageSize || 50))))
</script>