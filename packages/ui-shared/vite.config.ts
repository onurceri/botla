import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'BotlaUIShared',
      fileName: 'index',
      formats: ['es'],
    },
    rollupOptions: {
      external: ['react', 'react-dom', 'preact', 'preact/hooks'],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          preact: 'preact',
          'preact/hooks': 'preactHooks',
        },
      },
    },
  },
})
