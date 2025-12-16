import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'
import { readFileSync, writeFileSync } from 'fs'
import { resolve } from 'path'

// Plugin to fix preview.html widget path for production
function previewHtmlPlugin() {
  return {
    name: 'preview-html-transform',
    closeBundle() {
      // After build, modify preview.html in dist folder
      const previewPath = resolve(__dirname, 'dist/preview.html')
      try {
        let content = readFileSync(previewPath, 'utf-8')
        content = content.replace('/src/widget.tsx', './widget.js')
        writeFileSync(previewPath, content)
        console.log('✓ Fixed preview.html widget path for production')
      } catch (e) {
        console.warn('Could not transform preview.html:', e.message)
      }
    }
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [preact({ jsxImportSource: 'preact' }), previewHtmlPlugin()],
  test: {
    environment: 'jsdom'
  },
  resolve: {
    alias: {
      react: 'preact/compat',
      'react-dom': 'preact/compat',
      'react-dom/client': 'preact/compat',
      'react/jsx-runtime': 'preact/jsx-runtime',
    },
  },
  build: {
    lib: {
      entry: 'src/widget.tsx',
      name: 'ChatbotWidget',
      fileName: () => 'widget.js',
      formats: ['iife'],
    },
    rollupOptions: {
      output: {
        inlineDynamicImports: true,
      },
    },
  },
})
