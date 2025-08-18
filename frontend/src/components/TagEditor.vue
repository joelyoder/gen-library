<template>
  <div>
    <div class="mb-2">
      <span
        v-for="tag in localTags"
        :key="tag"
        class="badge text-bg-secondary me-1 mb-1"
      >
        {{ tag }}
        <button
          type="button"
          class="btn-close btn-close-white btn-sm ms-1"
          aria-label="Remove"
          @click="onRemove(tag)"
        ></button>
      </span>
    </div>
    <div class="input-group">
      <input
        class="form-control"
        v-model="newTag"
        @keyup.enter="onAdd"
        placeholder="Add tag"
      />
      <button class="btn btn-outline-primary" @click="onAdd" :disabled="!newTag.trim()">Add</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { addTags, removeTags } from '../api'

const props = defineProps<{ imageId: number, tags: string[] }>()
const emit = defineEmits(['update:tags'])

const localTags = ref<string[]>([])
const newTag = ref('')

watch(() => props.tags, (v) => {
  localTags.value = [...v]
}, { immediate: true })

async function onAdd() {
  const tag = newTag.value.trim()
  if (!tag) return
  await addTags(props.imageId, [tag])
  const updated = [...localTags.value, tag]
  localTags.value = updated
  emit('update:tags', updated)
  newTag.value = ''
}

async function onRemove(tag: string) {
  await removeTags(props.imageId, [tag])
  const updated = localTags.value.filter(t => t !== tag)
  localTags.value = updated
  emit('update:tags', updated)
}
</script>

