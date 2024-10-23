import adapter from "@sveltejs/adapter-static";
import { sveltePreprocess } from "svelte-preprocess";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  preprocess: sveltePreprocess(),

  kit: {
    adapter: adapter({
      fallback: "index.html",
    }),
    files: {
      assets: "../web-common/static",
    },
  },
};

export default config;
