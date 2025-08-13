<template>
  <div>
    <nav class="navbar navbar-expand navbar-light bg-light mb-3">
      <div class="container">
        <a class="navbar-brand" href="#">AI Image Library</a>
        <div class="navbar-nav">
          <RouterLink to="/" class="nav-link">Library</RouterLink>
          <RouterLink to="/settings" class="nav-link">Settings</RouterLink>
        </div>
      </div>
    </nav>
    <div class="container py-3">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>
