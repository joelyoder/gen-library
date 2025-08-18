<template>
  <div class="card">
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label">Model</label>
        <input class="form-control" v-model="form.modelName" />
      </div>
      <div class="mb-3">
        <label class="form-label">Model Hash</label>
        <input class="form-control" v-model="form.modelHash" />
      </div>
      <div class="mb-3">
        <label class="form-label">Prompt</label>
        <textarea class="form-control" rows="3" v-model="form.prompt"></textarea>
      </div>
      <div class="mb-3">
        <label class="form-label">Negative Prompt</label>
        <textarea class="form-control" rows="3" v-model="form.negativePrompt"></textarea>
      </div>
      <div class="row mb-3">
        <div class="col">
          <label class="form-label">Sampler</label>
          <input class="form-control" v-model="form.sampler" />
        </div>
        <div class="col">
          <label class="form-label">Steps</label>
          <input class="form-control" type="number" v-model="form.steps" />
        </div>
        <div class="col">
          <label class="form-label">CFG</label>
          <input class="form-control" type="number" step="0.1" v-model="form.cfgScale" />
        </div>
        <div class="col">
          <label class="form-label">Seed</label>
          <input class="form-control" v-model="form.seed" />
        </div>
      </div>
      <div class="row mb-3">
        <div class="col">
          <label class="form-label">Scheduler</label>
          <input class="form-control" v-model="form.scheduler" />
        </div>
        <div class="col">
          <label class="form-label">Clip Skip</label>
          <input class="form-control" type="number" v-model="form.clipSkip" />
        </div>
        <div class="col">
          <label class="form-label">Source App</label>
          <input class="form-control" v-model="form.sourceApp" />
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label">Tags</label>
        <TagEditor :image-id="props.image.id" v-model:tags="tags" />
      </div>

      <div class="mb-3">
        <button class="btn btn-link p-0" type="button" @click="rawOpen = !rawOpen">
          {{ rawOpen ? 'Hide Raw JSON' : 'Show Raw JSON' }}
        </button>
        <pre v-if="rawOpen" class="mt-2"><code>{{ rawJson }}</code></pre>
      </div>

      <div class="d-grid">
        <button class="btn btn-primary" @click="onSave">Save</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, computed } from 'vue'
import TagEditor from './TagEditor.vue'
import { updateImageMetadata } from '../api'

const props = defineProps<{ image: any }>()
const emit = defineEmits(['saved'])

const form = reactive({
  modelName: '',
  modelHash: '',
  prompt: '',
  negativePrompt: '',
  sampler: '',
  steps: '' as any,
  cfgScale: '' as any,
  seed: '',
  scheduler: '',
  clipSkip: '' as any,
  sourceApp: ''
})

const tags = ref<string[]>([])
const rawOpen = ref(false)

watch(() => props.image, (img) => {
  form.modelName = img?.modelName ?? ''
  form.modelHash = img?.modelHash ?? ''
  form.prompt = img?.prompt ?? ''
  form.negativePrompt = img?.negativePrompt ?? ''
  form.sampler = img?.sampler ?? ''
  form.steps = img?.steps ?? ''
  form.cfgScale = img?.cfgScale ?? ''
  form.seed = img?.seed ?? ''
  form.scheduler = img?.scheduler ?? ''
  form.clipSkip = img?.clipSkip ?? ''
  form.sourceApp = img?.sourceApp ?? ''
  tags.value = img?.tags?.map((t: any) => t.name) ?? []
  rawOpen.value = false
}, { immediate: true })

const rawJson = computed(() => {
  try {
    return JSON.stringify(props.image?.rawMetadata || {}, null, 2)
  } catch {
    return ''
  }
})

async function onSave() {
  const payload: any = {
    modelName: form.modelName || null,
    modelHash: form.modelHash || null,
    prompt: form.prompt || null,
    negativePrompt: form.negativePrompt || null,
    sampler: form.sampler || null,
    steps: form.steps !== '' ? Number(form.steps) : null,
    cfgScale: form.cfgScale !== '' ? Number(form.cfgScale) : null,
    seed: form.seed || null,
    scheduler: form.scheduler || null,
    clipSkip: form.clipSkip !== '' ? Number(form.clipSkip) : null,
    sourceApp: form.sourceApp || null
  }
  await updateImageMetadata(props.image.id, payload)
  emit('saved')
}
</script>

