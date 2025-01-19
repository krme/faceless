/** @type {import('tailwindcss').Config} */
// const colors = require('tailwindcss/colors');
const plugin = require('tailwindcss/plugin');

module.exports = {
  darkMode: 'selector',
  important: true,
  content: [
    'web/view/**/*.templ',
    'web/static/**/*.html',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Oxanium', 'sans-serif'],
        serif: ['Merriweather', 'serif'],
        icons: ['MaterialIcons'],
      },
      transitionDuration: {
        '10000': '10000ms',
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    plugin(function({ addVariant }) {
      addVariant('children', '&>*')
    }),
    require('daisyui'),
  ],
  safelist: [
    {
      pattern: /^(?!(?:scroll|bottom)$)m\w?-/,
      variants: ['sm', 'md', 'lg', 'xl', '2xl'],
    },
    {
      pattern: /^(?!(?:scroll|bottom)$)p\w?-/,
      variants: ['sm', 'md', 'lg', 'xl', '2xl'],
    },
  ],
}

