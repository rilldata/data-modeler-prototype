<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import { getConnectionFromEmail } from "./utils";
  import { ArrowLeftIcon } from "lucide-svelte";
  import { WebAuth } from "auth0-js";

  const dispatch = createEventDispatcher();

  export let disabled = false;
  export let email = "";
  export let webAuth: WebAuth;
  export let connectionMapObj: Record<string, string[]>;

  let errorText = "";

  function handleClick() {
    void authorizeSSO(email.toLowerCase());
  }

  function displayError(err: any) {
    errorText = err.message;
  }

  function authorizeSSO(email: string) {
    disabled = true;
    errorText = "";

    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    if (!connectionName) {
      displayError({
        message: `IDP for the email ${email} not found. Please contact your administrator.`,
      });
      disabled = false;
      return;
    }

    webAuth.authorize({
      connection: connectionName,
      login_hint: email,
      prompt: "login",
    });
  }
</script>

<div class="flex flex-col gap-y-4">
  <CtaButton {disabled} variant="primary" on:click={handleClick}>
    <div class="flex justify-center font-medium">
      <span>Continue with SAML SSO</span>
    </div>
  </CtaButton>
  <CtaButton
    {disabled}
    variant="secondary"
    gray
    on:click={() => {
      dispatch("back");
    }}
  >
    <div class="flex justify-center items-center font-medium">
      <ArrowLeftIcon class="mr-1" size={14} />
      <span>Back</span>
    </div>
  </CtaButton>
</div>

{#if errorText}
  <div class="mt-2 text-red-500 text-sm">{errorText}</div>
{/if}
