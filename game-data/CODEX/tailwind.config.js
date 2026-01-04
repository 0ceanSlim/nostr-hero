/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./html/**/*.html",
    "./src/**/*.js",
  ],
  theme: {
    extend: {
      colors: {
        // Dracula theme colors
        'codex-bg-primary': '#282a36',
        'codex-bg-secondary': '#44475a',
        'codex-bg-tertiary': '#6272a4',
        'codex-text-primary': '#f8f8f2',
        'codex-text-secondary': '#e0e0e0',
        'codex-text-muted': '#6272a4',
        'codex-green': '#50fa7b',
        'codex-cyan': '#8be9fd',
        'codex-purple': '#bd93f9',
        'codex-pink': '#ff79c6',
        'codex-yellow': '#f1fa8c',
        'codex-orange': '#ffb86c',
        'codex-red': '#ff5555',
      }
    },
  },
  plugins: [],
}
