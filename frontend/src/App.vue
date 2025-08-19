<template>
  <div class="container py-3">
    <div class="d-flex align-items-center mb-3">
      <h1 class="h3 mb-0 me-3">Gen Library</h1>
      <nav class="me-auto">
        <RouterLink to="/" class="me-3">Library</RouterLink>
        <RouterLink to="/settings">Settings</RouterLink>
      </nav>
      <select class="form-select w-auto" v-model="nsfw">
        <option value="hide">Hide NSFW</option>
        <option value="show">Show NSFW</option>
        <option value="only">Only NSFW</option>
      </select>
    </div>
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { nsfw } from './nsfw'
</script>