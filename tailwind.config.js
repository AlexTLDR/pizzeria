/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.html", "./internal/**/*.go", "./static/**/*.js"],
  safelist: [
    'gradient-background',
    'animate-gradient',
  ],
  theme: {
    extend: {
      colors: {
        'pizza-red': '#D62828',
        'pizza-green': '#2B9348',
        'pizza-beige': '#F2E8CF',
      },
      fontFamily: {
        'display': ['Playfair Display', 'serif'],
        'body': ['Roboto', 'sans-serif'],
      }
    },
  },
  plugins: [],
}
