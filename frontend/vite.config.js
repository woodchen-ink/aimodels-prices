import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  publicDir: 'public',
  server: {
    proxy: {
      '/api': {
        target: 'https://aimodels-prices.q58.pro',
        changeOrigin: true
      }
    }
  }
}) 