import { redirect } from "@sveltejs/kit";
import {
  adminServiceFindOrganizations,
  adminServiceGetCurrentUser,
} from "../client";
import { ADMIN_URL } from "../client/http-client";

// The current user is not available on the web server, so SSR would screw up the redirect logic
export const ssr = false;

export async function load() {
  const user = await adminServiceGetCurrentUser();

  if (user.user) {
    const orgs = await adminServiceFindOrganizations();
    throw redirect(307, `/${orgs.organization[0].name}`);
  } else {
    throw redirect(307, `${ADMIN_URL}/auth/login?redirect=${window.origin}`);
  }
}
