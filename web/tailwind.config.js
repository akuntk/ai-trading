/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Binance theme colors
        'binance-yellow': '#f0b90b',
        'binance-yellow-dark': '#c99400',
        'binance-yellow-light': '#fcd535',
        'binance-yellow-glow': 'rgba(240, 185, 11, 0.2)',
        'binance-green': '#0ecb81',
        'binance-red': '#f6465d',

        // Dark background colors
        background: '#000000',
        'background-elevated': '#0a0a0a',
        'panel-bg': '#0a0a0a',
        'panel-bg-hover': '#111111',
        'panel-border': '#1a1a1a',
        'panel-border-hover': '#2a2a2a',

        // Text colors
        'text-primary': '#eaecef',
        'text-secondary': '#848e9c',
        'text-tertiary': '#5e6673',
        'text-disabled': '#474d57',

        // Default Tailwind colors (for compatibility)
        border: 'hsl(214.3 31.8% 91.4%)',
        input: 'hsl(214.3 31.8% 91.4%)',
        ring: 'hsl(222.2 84% 4.9%)',

        // Shadcn/ui colors (modified for dark theme)
        background: '#000000',
        foreground: '#eaecef',
        card: '#0a0a0a',
        'card-foreground': '#eaecef',
        popover: '#0a0a0a0',
        'popover-foreground': '#eaecef',
        primary: {
          DEFAULT: '#f0b90b',
          foreground: '#000000',
        },
        'primary-foreground': '#000000',
        secondary: {
          DEFAULT: '#1a1a1a',
          foreground: '#eaecef',
        },
        'secondary-foreground': '#eaecef',
        muted: {
          DEFAULT: '#1a1a1a',
          foreground: '#848e9c',
        },
        'muted-foreground': '#848e9c',
        accent: {
          DEFAULT: '#1a1a1a',
          foreground: '#eaecef',
        },
        'accent-foreground': '#eaecef',
        destructive: {
          DEFAULT: '#f6465d',
          foreground: '#ffffff',
        },
        'destructive-foreground': '#ffffff',
      },
    },
  },
  plugins: [],
}
