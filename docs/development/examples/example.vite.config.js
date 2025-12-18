import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig(({ mode }) => {
  const isDev = mode === 'development';

  return {
    root: '.', // Root is project directory
    publicDir: 'www/res', // Static assets

    build: {
      outDir: `www/dist/${isDev ? 'dev' : 'prod'}`,
      emptyOutDir: true,
      sourcemap: isDev,
      minify: !isDev ? 'esbuild' : false,

      rollupOptions: {
        input: {
          test: resolve(__dirname, 'src/pages/test.js'),
          game: resolve(__dirname, 'src/pages/game.js'),
          intro: resolve(__dirname, 'src/pages/intro.js'),
          'new-game': resolve(__dirname, 'src/pages/new-game.js'),
          index: resolve(__dirname, 'src/pages/index.js')
        },
        output: {
          entryFileNames: isDev ? '[name].js' : '[name].[hash].js',
          chunkFileNames: isDev ? 'chunks/[name].js' : 'chunks/[name].[hash].js',
          assetFileNames: 'assets/[name].[hash][extname]',

          // Split vendor code for better caching
          manualChunks(id) {
            if (id.includes('src/lib/')) {
              return 'vendor';
            }
            if (id.includes('src/ui/core/')) {
              return 'ui-core';
            }
          }
        }
      },

      // Enable tree shaking in production
      treeshake: !isDev,

      // Target modern browsers
      target: 'es2020'
    },

    server: {
      port: 5173,
      proxy: {
        // Proxy API calls to Go backend
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        }
      }
    },

    // Define global constants
    define: {
      __DEV__: isDev,
      __PROD__: !isDev,
      __VERSION__: JSON.stringify('1.0.0')
    },

    // Resolve configuration
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
        '@lib': resolve(__dirname, 'src/lib'),
        '@data': resolve(__dirname, 'src/data'),
        '@state': resolve(__dirname, 'src/state'),
        '@logic': resolve(__dirname, 'src/logic'),
        '@systems': resolve(__dirname, 'src/systems'),
        '@ui': resolve(__dirname, 'src/ui'),
        '@pages': resolve(__dirname, 'src/pages'),
        '@components': resolve(__dirname, 'src/components'),
        '@config': resolve(__dirname, 'src/config')
      }
    }
  };
});
