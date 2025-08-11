import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import LibraryView from './views/LibraryView.vue'
import SettingsView from './views/SettingsView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', component: LibraryView },
  { path: '/settings', component: SettingsView },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
