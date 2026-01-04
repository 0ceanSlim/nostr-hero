import { defineConfig } from 'vite';
import { resolve } from 'path';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [tailwindcss()],

  // Root directory for source files
  root: '.',

  // Public directory for static assets
  publicDir: 'www/res',

  // Build configuration
  build: {
    // Output directory (relative to project root)
    outDir: 'www/dist',

    // Don't empty outDir before build (preserve other files)
    emptyOutDir: false,

    // Generate source maps for debugging
    sourcemap: true,

    // Enable CSS code splitting for separate entry point CSS files
    cssCodeSplit: true,

    // Multiple entry points
    rollupOptions: {
      input: {
        // JavaScript entries (first one will import CSS)
        index: resolve(__dirname, 'src/entries/index.js'),
        game: resolve(__dirname, 'src/entries/game.js'),
        gameIntro: resolve(__dirname, 'src/entries/gameIntro.js'),
        newGame: resolve(__dirname, 'src/entries/newGame.js'),
        settings: resolve(__dirname, 'src/entries/settings.js'),
        discover: resolve(__dirname, 'src/entries/discover.js'),
        saves: resolve(__dirname, 'src/entries/saves.js'),
      },
      output: {
        // Output file naming pattern
        entryFileNames: '[name].js',
        chunkFileNames: 'chunks/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          // Keep CSS filenames predictable (no hash) for easy linking in Go templates
          if (assetInfo.name && assetInfo.name.endsWith('.css')) {
            // Use the entry point name for CSS files (no hash)
            return '[name].css';
          }
          return 'assets/[name]-[hash][extname]';
        }
      }
    },

    // Minification
    minify: 'terser',
    terserOptions: {
      compress: {
        // Remove console.log in production
        drop_console: false, // Keep for now during testing
      }
    }
  },

  // Development server configuration
  server: {
    port: 5173,
    open: false,
    watch: {
      // Prevent infinite rebuild loops
      ignored: ['**/www/dist/**']
    },
    proxy: {
      // Proxy API requests to Go server during development
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },

  // Define global constants
  define: {
    __DEV__: JSON.stringify(process.env.NODE_ENV !== 'production'),
    __PROD__: JSON.stringify(process.env.NODE_ENV === 'production')
  },

  // Resolve configuration
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
});
