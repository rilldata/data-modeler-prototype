import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export function load({ params }) {
  // TODO: Check to see if the sourceId exists server-side
  const sourceExists = true;

  if (sourceExists) {
    return {
      sourceId: params.id,
    };
  }

  console.log("params", params);
  throw error(404, "Source not found");
}
