import { redirect } from "@sveltejs/kit";

/**
 * Redirect `/dashboard/[name]` to `/explore/[name]`.
 * Maintains backwards compatibility with legacy URLs.
 */
export const load = ({ params }) => {
  throw redirect(307, `/explore/${params.name}`);
};
