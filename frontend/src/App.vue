<template>
  <div class="container py-3">
    <h1 class="h3 mb-3">AI Image Library</h1>
    <nav class="mb-3">
      <RouterLink to="/" class="me-3">Library</RouterLink>
      <RouterLink to="/settings">Settings</RouterLink>
    </nav>
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>