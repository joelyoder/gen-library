<template>
  <div class="card">
    <div class="card-body">
        <div class="mb-3">
          <label class="form-label">Search</label>
          <input class="form-control" v-model="localQ" @keyup.enter="$emit('search')" placeholder="prompt, model, metadata" />
        </div>

        <div class="mb-3">
          <label class="form-label">Tags (comma separated)</label>
          <input class="form-control" v-model="tagsCsv" />
        </div>

        <div class="row g-2">
          <div class="col-6">
            <label class="form-label">Sort</label>
            <select class="form-select" v-model="localSort">
              <option value="imported_at">Imported</option>
              <option value="created_time">Created</option>
              <option value="file_name">File name</option>
            </select>
          </div>
          <div class="col-6">
            <label class="form-label">Order</label>
            <select class="form-select" v-model="localOrder">
              <option value="desc">Desc</option>
              <option value="asc">Asc</option>
            </select>
          </div>
        </div>

        <div class="d-grid mt-3">
          <button class="btn btn-primary" @click="apply">Apply</button>
        </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  q: string
  tags: string[]
  sort: 'created_time'|'imported_at'|'file_name'
  order: 'asc'|'desc'
}>()
const emit = defineEmits(['update:q','update:tags','update:sort','update:order','search'])

const localQ = computed({
  get: () => props.q,
  set: v => emit('update:q', v)
})
const localSort = computed({
  get: () => props.sort,
  set: v => emit('update:sort', v)
})
const localOrder = computed({
  get: () => props.order,
  set: v => emit('update:order', v)
})

const tagsCsv = computed({
  get: () => props.tags.join(', '),
  set: v => emit('update:tags', v.split(',').map(s => s.trim()).filter(Boolean))
})

function apply() {
  emit('search')
}
</script>