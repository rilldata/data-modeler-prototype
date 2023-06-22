import Google from "@rilldata/web-common/components/icons/Google.svelte";
import Microsoft from "@rilldata/web-common/components//icons/Microsoft.svelte";

export const LOGIN_OPTIONS = [
  {
    name: "Google",
    icon: Google,
    connection: "google-oauth2",
    label: "Continue with Google",
    style: "primary",
  },
  {
    name: "Microsoft",
    icon: Microsoft,
    connection: "windowslive",
    label: "Continue with Microsoft",
    style: "secondary",
  },
];
