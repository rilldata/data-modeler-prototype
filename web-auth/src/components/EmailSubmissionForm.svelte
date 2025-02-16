<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import { validateEmail } from "./utils";
  import { createEventDispatcher } from "svelte";

  export let disabled = false;

  let email = "";
  let errorText = "";

  const dispatch = createEventDispatcher();

  const inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  const focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-primary-ry-300 focus:outline-none";

  function handleClick() {
    if (!email) {
      errorText = "Please enter your email";
      return;
    }

    if (!validateEmail(email)) {
      errorText = "Please enter a valid email address";
      return;
    }

    errorText = "";

    dispatch("submit", { email });
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      handleClick();
    }
  }

  $: buttonDisabled = disabled || !(email.length > 0 && validateEmail(email));
</script>

<div class="mb-4 flex flex-col gap-y-4">
  <input
    class="{inputClasses} {focusClasses}"
    style:width="400px"
    type="email"
    placeholder="Enter your email address"
    id="email"
    bind:value={email}
    on:keydown={handleKeyDown}
  />

  {#if errorText}
    <div class="text-red-500 text-sm -mt-2">
      {errorText}
    </div>
  {/if}
</div>

<CtaButton disabled={buttonDisabled} variant="secondary" on:click={handleClick}>
  <div class="flex justify-center font-medium">
    <span>Continue with email</span>
  </div>
</CtaButton>
