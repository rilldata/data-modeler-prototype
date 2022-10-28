module.exports = {
  /** Once we have applied dark styling to all UI elements, remove this line */
  darkMode: 'class',
  content: [
    "./**/*.html",
    "./**/*.svelte",
    "./src/lib/duckdb-data-types.ts",
    "./src/lib/components/chip/chip-types.ts"
  ],
  theme: {
    extend: {},
  },
  plugins: [],
  safelist: [
    'text-blue-800', // needed for blue text in filter pills
  ]
};
