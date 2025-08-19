<template>
  <div class="d-flex flex-column gap-2">
    <p><strong>Model:</strong> {{ props.image.modelName }}</p>
    <p><strong>Model Hash:</strong> {{ props.image.modelHash }}</p>
    <p><strong>NSFW:</strong> {{ props.image.nsfw ? 'Yes' : 'No' }}</p>
    <div>
      <strong>Loras:</strong>
      <div v-if="props.image.loras && props.image.loras.length">
        <div v-for="(l, i) in props.image.loras" :key="i">{{ l.name }} ({{ l.hash }})</div>
      </div>
      <div v-else class="text-muted">None</div>
    </div>
    <div>
      <strong>Prompt:</strong>
      <div class="small">{{ props.image.prompt }}</div>
    </div>
    <div>
      <strong>Negative Prompt:</strong>
      <div class="small">{{ props.image.negativePrompt }}</div>
    </div>
    <p><strong>Sampler:</strong> {{ props.image.sampler }}</p>
    <p><strong>Steps:</strong> {{ props.image.steps }}</p>
    <p><strong>CFG:</strong> {{ props.image.cfgScale }}</p>
    <p><strong>Seed:</strong> {{ props.image.seed }}</p>
    <p><strong>Scheduler:</strong> {{ props.image.scheduler }}</p>
    <p><strong>Clip Skip:</strong> {{ props.image.clipSkip }}</p>
    <p><strong>Source App:</strong> {{ props.image.sourceApp }}</p>
    <div>
      <strong>Tags:</strong>
      <span v-for="t in props.image.tags" :key="t.id" class="badge text-bg-secondary me-1">{{ t.name }}</span>
    </div>
    <div class="d-grid mt-3">
      <button class="btn btn-primary" @click="$emit('edit')">Edit</button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ image: any }>()
defineEmits(['edit'])
</script>
