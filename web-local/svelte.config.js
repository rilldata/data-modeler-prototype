import adapter from "@sveltejs/adapter-static";
import preprocess from "svelte-preprocess";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  preprocess: preprocess(),

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

// import adapter from "@sveltejs/adapter-auto";
// import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

// /** @type {import('@sveltejs/kit').Config} */
// const config = {
//   // Consult https://kit.svelte.dev/docs/integrations#preprocessors
//   // for more information about preprocessors
//   preprocess: vitePreprocess(),

//   kit: {
//     // adapter-auto only supports some environments, see https://kit.svelte.dev/docs/adapter-auto for a list.
//     // If your environment is not supported or you settled on a specific environment, switch out the adapter.
//     // See https://kit.svelte.dev/docs/adapters for more information about adapters.
//     adapter: adapter(),
//     files: {
//       assets: "../web-common/static",
//     },
//   },
// };

// export default config;
