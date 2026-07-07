import { fileURLToPath, URL } from 'node:url'
import react from '@vitejs/plugin-react-swc'
import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  base: './',
  plugins: [react(), tsconfigPaths()],
  css: {
    postcss: './postcss.config.cjs'
  },
  define: {
    __DOMAIN_BACKEND__: JSON.stringify(''),
    __NODE_ENV__: JSON.stringify(process.env.NODE_ENV || 'production'),
    __DOMAIN_OVERRIDE__: JSON.stringify('0')
  },
  server: {
    host: '0.0.0.0',
    port: 3333,
    cors: true,
    strictPort: true,
    allowedHosts: true,
    proxy: {
      '/api': 'http://127.0.0.1:3000'
    }
  },
  resolve: {
    alias: {
      '@app': fileURLToPath(new URL('./src/app', import.meta.url)),
      '@entities': fileURLToPath(new URL('./src/entities', import.meta.url)),
      '@features': fileURLToPath(new URL('./src/features', import.meta.url)),
      '@pages': fileURLToPath(new URL('./src/pages', import.meta.url)),
      '@shared': fileURLToPath(new URL('./src/shared', import.meta.url))
    }
  },
  build: {
    target: 'esNext',
    outDir: 'dist',
    chunkSizeWarningLimit: 1000000,
    commonjsOptions: {
      include: [/node_modules/, /vendor\/\@exodus/],
      transformMixedEsModules: true
    }
  }
})
