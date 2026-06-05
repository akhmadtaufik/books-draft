import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    // Ensure the server binds to all network interfaces for Docker compatibility
    host: true,
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://backend:8080',
        changeOrigin: true,
      },
    },
    watch: {
      // Ignored paths to prevent infinite HMR reload loops caused by backend writes
      ignored: [
        '**/backend/**', 
        '**/database/**', 
        '**/.git/**', 
        '**/docker-compose.yml',
        '**/.env*'
      ]
    }
  },
})
