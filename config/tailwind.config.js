module.exports = {
  mode: 'jit',
  content: [
    './www/**/*.{ts,tsx}',
  ],
  plugins: [
    require('@tailwindcss/aspect-ratio'),
    require('@tailwindcss/typography'),
  ],
  theme: {
    extend: {
      colors: {
        'desktop-green': '#00807F',
        'emphasis-text': '#2D2E5F',
        'highlight-color': '#FFFF00',
        'pane-gray': '#D4D0C9',
        'outset-dark-shadow': '#404040',
        'outset-light-shadow': '#808080',
        'inset-dark-shadow': 'black',
        'inset-light-shadow': '#808080',
        'hr-dark-shadow': '#808080',
        'hr-light-shadow': '#FFFFFF',
        'title-bar-text': '#FEFEFE',
        'title-bar-left': '#08008E',
        'title-bar-right': '#75B3D3',
        'logo-tile': '#095FA8',
        'link-blue': 'blue'
      }
    }
  }
}
