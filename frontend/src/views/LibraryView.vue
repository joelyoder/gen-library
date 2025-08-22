<template>
  <div class="row g-3">
    <div class="col-12 col-lg-3">
      <SidebarFilters
        v-model:q="q"
        v-model:tags="tags"
        v-model:sort="sort"
        v-model:order="order"
        v-model:rating="rating"
        @search="reload"
        @scan="onScan"
      />
    </div>
    <div class="col-12 col-lg-9">
      <ImageGrid :images="items" @deleted="onDeleted" @metadata="onMetadata" />
      <Pager
        :page="page"
        :page-size="pageSize"
        :total="total"
        @change="onPage"
      />
    </div>
  </div>

  <div v-if="metadataOpen">
    <div class="modal fade show d-block" tabindex="-1">
      <div class="modal-dialog modal-fullscreen">
        <div class="modal-content bg-dark text-light">
          <div class="modal-header">
            <h5 class="modal-title">{{ selectedImage?.fileName }}</h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              @click="closeMetadata"
            ></button>
          </div>
          <div class="modal-body p-0">
            <div class="row g-0 h-100">
              <div
                class="col-md-8 d-flex align-items-center justify-content-center bg-black position-relative"
              >
                <button
                  v-if="selectedIndex > 0"
                  class="btn btn-dark position-absolute top-50 start-0 translate-middle-y opacity-50"
                  @click="prevImage"
                >
                  <i class="bi bi-chevron-left fs-1"></i>
                </button>
                <img
                  v-if="selectedImage"
                  :src="
                    apiBase +
                    '/api/images/' +
                    selectedImage.id +
                    '/file?sha=' +
                    selectedImage.sha256
                  "
                  class="img-fluid"
                  :alt="selectedImage.fileName"
                  style="max-height: 100vh; width: auto"
                />
                <button
                  v-if="selectedIndex < items.length - 1"
                  class="btn btn-dark position-absolute top-50 end-0 translate-middle-y opacity-50"
                  @click="nextImage"
                >
                  <i class="bi bi-chevron-right fs-1"></i>
                </button>
              </div>
              <div class="col-md-4 overflow-auto p-3" v-if="selectedImage">
                <div class="d-flex justify-content-end mb-3">
                  <button
                    class="btn btn-outline-danger"
                    @click="onDeleteSelected"
                  >
                    Delete
                  </button>
                </div>
                <MetadataPanel
                  v-if="metadataEditing"
                  :image="selectedImage"
                  @saved="onMetadataSaved"
                  @cancel="metadataEditing = false"
                />
                <div v-else>
                  <MetadataDisplay
                    :image="selectedImage"
                    @edit="metadataEditing = true"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="modal-backdrop fade show"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watchEffect, watch, onUnmounted } from "vue";
import {
  listImages,
  scanLibrary,
  getLibraryPath,
  getImage,
  deleteImage,
  updateImageMetadata,
  apiBase,
} from "../api";
import SidebarFilters from "../components/SidebarFilters.vue";
import ImageGrid from "../components/ImageGrid.vue";
import Pager from "../components/Pager.vue";
import MetadataPanel from "../components/MetadataPanel.vue";
import MetadataDisplay from "../components/MetadataDisplay.vue";
import { nsfw } from "../nsfw";

const page = ref(1);
const pageSize = ref(50);
const q = ref("");
const tags = ref<string[]>([]);
const sort = ref<"created_time" | "imported_at" | "file_name">("created_time");
const order = ref<"asc" | "desc">("desc");
const rating = ref<number | null>(null);

const items = ref<any[]>([]);
const total = ref(0);
const metadataOpen = ref(false);
const selectedImage = ref<any | null>(null);
const metadataEditing = ref(false);
const selectedIndex = ref(-1);
const reloadKey = ref(0);

function reload() {
  page.value = 1;
  reloadKey.value++;
}

function onPage(newPage: number) {
  page.value = newPage;
}

async function onScan() {
  const root = await getLibraryPath();
  if (!root) {
    alert("Please set a library path in Settings first");
    return;
  }
  await scanLibrary(root);
  reload();
}

function onDeleted(id: number) {
  items.value = items.value.filter((img) => img.id !== id);
}

async function onMetadata(img: any) {
  const idx = items.value.findIndex((i) => i.id === img.id);
  await showImageAt(idx);
}

function closeMetadata() {
  metadataOpen.value = false;
  selectedImage.value = null;
  metadataEditing.value = false;
  selectedIndex.value = -1;
}

async function showImageAt(idx: number) {
  if (idx < 0 || idx >= items.value.length) return;
  selectedIndex.value = idx;
  const img = await getImage(items.value[idx].id);
  selectedImage.value = img;
  metadataOpen.value = true;
  metadataEditing.value = false;
}

function prevImage() {
  showImageAt(selectedIndex.value - 1);
}

function nextImage() {
  showImageAt(selectedIndex.value + 1);
}

async function toggleNsfw() {
  if (!selectedImage.value) return;
  const newVal = !selectedImage.value.nsfw;
  await updateImageMetadata(selectedImage.value.id, { nsfw: newVal });
  selectedImage.value.nsfw = newVal;
  const idx = items.value.findIndex((i) => i.id === selectedImage.value?.id);
  if (idx !== -1) {
    if (
      (nsfw.value === "hide" && newVal) ||
      (nsfw.value === "only" && !newVal)
    ) {
      items.value.splice(idx, 1);
    } else {
      items.value[idx].nsfw = newVal;
    }
  }
}

async function setRating(n: number) {
  if (!selectedImage.value) return;
  await updateImageMetadata(selectedImage.value.id, { rating: n });
  selectedImage.value.rating = n;
  const idx = items.value.findIndex((i) => i.id === selectedImage.value?.id);
  if (idx !== -1) {
    if (rating.value !== null && n !== rating.value) {
      items.value.splice(idx, 1);
    } else {
      items.value[idx].rating = n;
    }
  }
}

async function onMetadataSaved() {
  if (selectedImage.value) {
    const updated = await getImage(selectedImage.value.id);
    selectedImage.value = updated;
    const idx = items.value.findIndex((i) => i.id === updated.id);
    if (idx !== -1) {
      if (
        (nsfw.value === "hide" && updated.nsfw) ||
        (nsfw.value === "only" && !updated.nsfw)
      ) {
        items.value.splice(idx, 1);
      } else {
        items.value[idx] = updated;
      }
    }
  }
  metadataEditing.value = false;
}

async function onDeleteSelected() {
  if (!selectedImage.value) return;
  if (!confirm("Delete this image?")) return;
  await deleteImage(selectedImage.value.id);
  items.value = items.value.filter((img) => img.id !== selectedImage.value?.id);
  closeMetadata();
}

function onKeydown(e: KeyboardEvent) {
  if (!metadataOpen.value || metadataEditing.value) return;
  const tag = (e.target as HTMLElement).tagName;
  if (["INPUT", "TEXTAREA", "SELECT"].includes(tag)) return;
  switch (e.key) {
    case "ArrowLeft":
      e.preventDefault();
      prevImage();
      break;
    case "ArrowRight":
      e.preventDefault();
      nextImage();
      break;
    case "Escape":
      e.preventDefault();
      closeMetadata();
      break;
    case "f":
      e.preventDefault();
      toggleNsfw();
      break;
    case "d":
      e.preventDefault();
      onDeleteSelected();
      break;
    case "e":
      e.preventDefault();
      metadataEditing.value = true;
      break;
    case "0":
    case "1":
    case "2":
    case "3":
    case "4":
    case "5":
      e.preventDefault();
      setRating(Number(e.key));
      break;
  }
}

watch(metadataOpen, (open) => {
  if (open) {
    window.addEventListener("keydown", onKeydown);
  } else {
    window.removeEventListener("keydown", onKeydown);
  }
});

onUnmounted(() => {
  window.removeEventListener("keydown", onKeydown);
});

watchEffect(async () => {
  reloadKey.value;
  const data = await listImages({
    page: page.value,
    pageSize: pageSize.value,
    q: q.value || undefined,
    tags: tags.value,
    nsfw: nsfw.value,
    sort: sort.value,
    order: order.value,
    rating: rating.value ?? undefined,
  });
  items.value = data.items;
  total.value = data.total;
});
</script>
