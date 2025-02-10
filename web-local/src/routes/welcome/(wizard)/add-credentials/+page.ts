import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export async function load({ parent }) {
  const { onboardingState } = await parent();

  const { managementType, olapDriver } = onboardingState;
  if (!get(managementType) || !get(olapDriver)) {
    throw redirect(302, "/welcome/select-connectors");
  }

  return {
    onboardingState,
  };
}
