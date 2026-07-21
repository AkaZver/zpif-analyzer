/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#7c5cbf',
          light: '#9b7ed8',
          dark: '#5d507e',
        },
        surface: {
          DEFAULT: '#1a1a1a',
          light: '#2a2a2a',
          card: '#333333',
        },
        accent: {
          DEFAULT: '#e94560',
          light: '#ff6b6b',
        },
        text: {
          primary: '#e0e0e0',
          secondary: '#a0a0a0',
          muted: '#6c6c6c',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
