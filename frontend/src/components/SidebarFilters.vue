<template>
  <div class="card">
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label">Search</label>
        <input
          class="form-control"
          v-model="localQ"
          @keyup.enter="$emit('search')"
          placeholder="prompt, model, metadata"
        />
      </div>

      <div class="mb-3">
        <label class="form-label">Tags (comma separated)</label>
        <input class="form-control" v-model="tagsCsv" />
      </div>

      <div class="mb-3">
        <label class="form-label">Rating</label>
        <select class="form-select" v-model="localRating">
          <option value="">Any</option>
          <option value="0">0 Stars</option>
          <option value="1">1 Star</option>
          <option value="2">2 Stars</option>
          <option value="3">3 Stars</option>
          <option value="4">4 Stars</option>
          <option value="5">5 Stars</option>
        </select>
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

      <div class="d-grid gap-2 mt-3">
        <button class="btn btn-primary" @click="apply">Apply</button>
        <button class="btn btn-secondary" @click="$emit('scan')">
          Scan Library
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";

const props = defineProps<{
  q: string;
  tags: string[];
  sort: "created_time" | "imported_at" | "file_name";
  order: "asc" | "desc";
  rating: number | null;
}>();
const emit = defineEmits([
  "update:q",
  "update:tags",
  "update:sort",
  "update:order",
  "update:rating",
  "search",
  "scan",
]);

const localQ = computed({
  get: () => props.q,
  set: (v) => emit("update:q", v),
});
const localSort = computed({
  get: () => props.sort,
  set: (v) => emit("update:sort", v),
});
const localOrder = computed({
  get: () => props.order,
  set: (v) => emit("update:order", v),
});

const localRating = computed({
  get: () => (props.rating == null ? "" : String(props.rating)),
  set: (v) => emit("update:rating", v === "" ? null : Number(v)),
});

const tagsCsv = computed({
  get: () => props.tags.join(", "),
  set: (v) =>
    emit(
      "update:tags",
      v
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean),
    ),
});

function apply() {
  emit("search");
}

const searchTimer = ref<number | null>(null);
watch(localQ, () => {
  if (searchTimer.value) {
    clearTimeout(searchTimer.value);
  }
  searchTimer.value = window.setTimeout(() => emit("search"), 300);
});
</script>
