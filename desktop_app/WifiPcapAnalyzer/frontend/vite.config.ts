import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  // Wails specific build options
  build: {
    // Output directory for Wails
    outDir: 'dist',
    // Sourcemaps for debugging
    sourcemap: process.env.NODE_ENV !== 'production',
  },
  esbuild: {
    drop: process.env.NODE_ENV === 'production' ? ['console', 'debugger'] : [],
  },
  // Optional: server configuration if needed during dev
  server: {
    port: 3000, // Or any other port you prefer for Vite dev server
  },
})