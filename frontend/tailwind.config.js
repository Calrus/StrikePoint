/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: '#0a0a0a', // Deep Matte Black
        surface: '#1e293b',    // Dark Slate Gray
        primary: '#8b5cf6',    // Neon Purple (Degen)
        success: '#10b981',    // Emerald Green (Profit/Safe)
      },
      fontFamily: {
        mono: ['"JetBrains Mono"', '"Roboto Mono"', 'monospace'],
        sans: ['Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
