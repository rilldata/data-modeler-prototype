export const ssr = false;

import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export async function load({ url, depends, untrack }) {
  depends("init");

  const initialized = await isProjectInitialized(get(runtime).instanceId);

  const inOnboardingFlow = untrack(() => url.pathname.startsWith("/welcome"));

  if (!initialized && !inOnboardingFlow) {
    throw redirect(303, "/welcome");
  }
}
