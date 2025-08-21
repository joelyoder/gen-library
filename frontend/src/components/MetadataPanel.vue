<template>
  <div class="card">
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label">Rating</label>
        <div>
          <i
            v-for="n in 5"
            :key="n"
            class="bi text-warning"
            :class="n <= form.rating ? 'bi-star-fill' : 'bi-star'"
            @click="form.rating = form.rating === n ? 0 : n"
            style="cursor: pointer"
          ></i>
        </div>
      </div>
      <div class="mb-3">
        <label class="form-label">Model</label>
        <input class="form-control" v-model="form.modelName" />
      </div>
      <div class="mb-3">
        <label class="form-label">Model Hash</label>
        <input class="form-control" v-model="form.modelHash" />
      </div>
      <div class="mb-3">
        <label class="form-label">Resolution</label>
        <p class="form-control-plaintext">
          {{ props.image.width }}x{{ props.image.height }}
        </p>
      </div>
      <div class="mb-3">
        <label class="form-label">Loras</label>
        <div v-for="(l, i) in loras" :key="i" class="input-group mb-1">
          <input class="form-control" placeholder="Name" v-model="l.name" />
          <input class="form-control" placeholder="Hash" v-model="l.hash" />
          <button
            class="btn btn-outline-danger"
            type="button"
            @click="removeLora(i)"
          >
            &times;
          </button>
        </div>
        <button
          class="btn btn-outline-primary btn-sm"
          type="button"
          @click="addLora"
        >
          Add Lora
        </button>
      </div>
      <div class="mb-3">
        <label class="form-label">Prompt</label>
        <textarea
          class="form-control"
          rows="3"
          v-model="form.prompt"
        ></textarea>
      </div>
      <div class="mb-3">
        <label class="form-label">Negative Prompt</label>
        <textarea
          class="form-control"
          rows="3"
          v-model="form.negativePrompt"
        ></textarea>
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
          <input
            class="form-control"
            type="number"
            step="0.1"
            v-model="form.cfgScale"
          />
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

      <div class="form-check form-switch mb-3">
        <input
          class="form-check-input"
          type="checkbox"
          id="nsfwCheck"
          v-model="form.nsfw"
        />
        <label class="form-check-label" for="nsfwCheck">NSFW</label>
      </div>

      <div class="mb-3">
        <button
          class="btn btn-link p-0"
          type="button"
          @click="rawOpen = !rawOpen"
        >
          {{ rawOpen ? "Hide Raw JSON" : "Show Raw JSON" }}
        </button>
        <pre v-if="rawOpen" class="mt-2"><code>{{ rawJson }}</code></pre>
      </div>

      <div class="d-flex gap-2">
        <button class="btn btn-primary" type="button" @click="onSave">Save</button>
        <button class="btn btn-secondary" type="button" @click="emit('cancel')">
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, computed } from "vue";
import TagEditor from "./TagEditor.vue";
import { updateImageMetadata } from "../api";

const props = defineProps<{ image: any }>();
const emit = defineEmits(["saved", "cancel"]);

const form = reactive({
  rating: 0,
  modelName: "",
  modelHash: "",
  prompt: "",
  negativePrompt: "",
  sampler: "",
  steps: "" as any,
  cfgScale: "" as any,
  seed: "",
  scheduler: "",
  clipSkip: "" as any,
  sourceApp: "",
  nsfw: false,
});

const tags = ref<string[]>([]);
const loras = ref<{ name: string; hash: string }[]>([]);
const rawOpen = ref(false);

watch(
  () => props.image,
  (img) => {
    form.modelName = img?.modelName ?? "";
    form.modelHash = img?.modelHash ?? "";
    form.prompt = img?.prompt ?? "";
    form.negativePrompt = img?.negativePrompt ?? "";
    form.sampler = img?.sampler ?? "";
    form.steps = img?.steps ?? "";
    form.cfgScale = img?.cfgScale ?? "";
    form.seed = img?.seed ?? "";
    form.scheduler = img?.scheduler ?? "";
    form.clipSkip = img?.clipSkip ?? "";
    form.sourceApp = img?.sourceApp ?? "";
    form.nsfw = !!img?.nsfw;
    form.rating = img?.rating ?? 0;
    tags.value = img?.tags?.map((t: any) => t.name) ?? [];
    loras.value =
      img?.loras?.map((l: any) => ({ name: l.name, hash: l.hash })) ?? [];
    rawOpen.value = false;
  },
  { immediate: true },
);

const rawJson = computed(() => {
  try {
    return JSON.stringify(props.image?.rawMetadata || {}, null, 2);
  } catch {
    return "";
  }
});

async function onSave() {
  const payload: any = {
    rating: form.rating,
    modelName: form.modelName || null,
    modelHash: form.modelHash || null,
    prompt: form.prompt || null,
    negativePrompt: form.negativePrompt || null,
    sampler: form.sampler || null,
    steps: form.steps !== "" ? Number(form.steps) : null,
    cfgScale: form.cfgScale !== "" ? Number(form.cfgScale) : null,
    seed: form.seed || null,
    scheduler: form.scheduler || null,
    clipSkip: form.clipSkip !== "" ? Number(form.clipSkip) : null,
    sourceApp: form.sourceApp || null,
    loras: loras.value.filter((l) => l.name || l.hash),
    nsfw: form.nsfw,
  };
  await updateImageMetadata(props.image.id, payload);
  emit("saved");
}

function addLora() {
  loras.value.push({ name: "", hash: "" });
}

function removeLora(i: number) {
  loras.value.splice(i, 1);
}
</script>
