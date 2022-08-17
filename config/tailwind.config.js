module.exports = {
  mode: 'jit',
  content: [
    './www/**/*.{ts,tsx}',
  ],
  plugins: [
    require('@tailwindcss/aspect-ratio'),
    require('@tailwindcss/typography'),
  ],
}
