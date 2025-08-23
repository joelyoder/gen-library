<template>
  <div class="col-12 col-lg-6">
    <h2 class="h4 mb-3">Settings</h2>
    <form @submit.prevent="save">
      <div class="mb-3">
        <label class="form-label">Library Folder</label>
        <input
          v-model="path"
          class="form-control"
          placeholder="D:\\AI\\library"
        />
      </div>
      <button class="btn btn-primary" type="submit">Save</button>
    </form>
    <div class="mt-4">
      <span>Watcher:</span>
      <span
        :class="watcherRunning ? 'text-success' : 'text-danger'"
        class="ms-1"
      >
        {{ watcherRunning ? "Running" : "Stopped" }}
      </span>
      <button
        class="btn btn-secondary ms-3"
        type="button"
        @click="toggleWatcher"
      >
        {{ watcherRunning ? "Stop" : "Start" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import {
  getLibraryPath,
  setLibraryPath,
  getWatcherStatus,
  startWatcher,
  stopWatcher,
} from "../api";

const path = ref("");
const watcherRunning = ref(false);

onMounted(async () => {
  path.value = await getLibraryPath();
  watcherRunning.value = await getWatcherStatus();
});

async function save() {
  if (!path.value.trim()) {
    alert("Please enter a folder path");
    return;
  }
  await setLibraryPath(path.value.trim());
  alert("Saved");
}

async function toggleWatcher() {
  if (watcherRunning.value) {
    await stopWatcher();
  } else {
    await startWatcher();
  }
  watcherRunning.value = await getWatcherStatus();
}
</script>
