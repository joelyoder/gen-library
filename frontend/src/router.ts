import { createRouter, createWebHistory } from 'vue-router'
import LibraryView from './views/LibraryView.vue'
import SettingsView from './views/SettingsView.vue'

const routes = [
  { path: '/', component: LibraryView },
  { path: '/settings', component: SettingsView }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
