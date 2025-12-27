import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'
import { readFileSync, writeFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

import { visualizer } from 'rollup-plugin-visualizer'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// Plugin to fix preview.html widget path for production
function previewHtmlPlugin() {
  return {
    name: 'preview-html-transform',
    closeBundle() {
      // After build, modify preview.html and index.html in dist folder
      const files = ['preview.html']  // index.html no longer loads widget
      
      files.forEach(file => {
        const filePath = resolve(__dirname, `dist/${file}`)
        try {
          let content = readFileSync(filePath, 'utf-8')
          content = content.replace('/src/widget.tsx', './widget.js')
          writeFileSync(filePath, content)
          console.log(`✓ Fixed ${file} widget path for production`)
        } catch (e) {
          // It's okay if file doesn't exist (e.g. clean build might not have copied it yet if something went wrong, but usually it should be there)
          // Actually public dir is copied before closeBundle?
          // Vite copies public dir assets.
          console.warn(`Could not transform ${file}:`, e.message)
        }
      })
    }
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    preact({ jsxImportSource: 'preact' }), 
    previewHtmlPlugin(),
    visualizer({
      filename: 'dist/stats.html',
      open: false,
      gzipSize: true,
    }),
  ],
  test: {
    environment: 'jsdom',
    exclude: ['**/node_modules/**', '**/dist/**', '**/e2e/**'],
    alias: {
      react: 'preact/compat',
      'react-dom': 'preact/compat',
      'react-dom/client': 'preact/compat',
      'react/jsx-runtime': 'preact/jsx-runtime',
    },
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
