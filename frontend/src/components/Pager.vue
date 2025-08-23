<template>
  <nav
    v-if="totalPages > 1"
    class="my-3"
    aria-label="pager"
  >
    <ul class="pagination justify-content-center align-items-center gap-1 mb-0">
      <li class="page-item" :class="{ disabled: page === 1 }">
        <a class="page-link" href="#" @click.prevent="changePage(1)">First</a>
      </li>
      <li class="page-item" :class="{ disabled: page === 1 }">
        <a class="page-link" href="#" @click.prevent="changePage(page - 1)">Previous</a>
      </li>
      <li class="d-flex align-items-center">
        <input
          type="number"
          min="1"
          :max="totalPages"
          v-model.number="pageInput"
          @keyup.enter="goToPage"
          class="form-control"
          style="width: 80px"
        />
        <span class="ms-1">/ {{ totalPages }}</span>
      </li>
      <li class="page-item" :class="{ disabled: page === totalPages }">
        <a class="page-link" href="#" @click.prevent="changePage(page + 1)">Next</a>
      </li>
      <li class="page-item" :class="{ disabled: page === totalPages }">
        <a class="page-link" href="#" @click.prevent="changePage(totalPages)">Last</a>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";

const props = defineProps<{ page: number; pageSize: number; total: number }>();
const emit = defineEmits<{
  (e: "change", page: number): void;
}>();

const totalPages = computed(() =>
  Math.max(1, Math.ceil(props.total / (props.pageSize || 50)))
);
const pageInput = ref(props.page);

watch(
  () => props.page,
  (p) => {
    pageInput.value = p;
  }
);

function changePage(p: number) {
  if (p < 1 || p > totalPages.value || p === props.page) return;
  emit("change", p);
}

function goToPage() {
  let p = pageInput.value;
  if (p < 1) p = 1;
  if (p > totalPages.value) p = totalPages.value;
  emit("change", p);
}
</script>

