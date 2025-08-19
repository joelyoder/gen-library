import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8081',
      '/thumbs': 'http://localhost:8081',
    },
    host: '0.0.0.0', // <-- this is key
    port: 5174
  }
})