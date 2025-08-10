<template>
  <div class="container py-3">
    <h1 class="h3 mb-3">AI Image Library</h1>
    <nav class="mb-3">
      <ul class="nav nav-pills">
        <li class="nav-item">
          <router-link to="/" class="nav-link" active-class="active" exact-active-class="active">Library</router-link>
        </li>
        <li class="nav-item">
          <router-link to="/settings" class="nav-link" active-class="active">Settings</router-link>
        </li>
      </ul>
    </nav>
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>