import type { BeforeNavigate } from "@sveltejs/kit";
import { writable } from "svelte/store";

export type FeatureFlags = {
  readOnly: boolean;
};
export const featureFlags = writable<FeatureFlags>({
  readOnly: undefined,
});

export function retainFeaturesFlags(navigation: BeforeNavigate) {
  if (!navigation.from.url.searchParams.has("features")) {
    return;
  }

  navigation.to.url.searchParams.set(
    "features",
    navigation.from.url.searchParams.get("features")
  );
}
