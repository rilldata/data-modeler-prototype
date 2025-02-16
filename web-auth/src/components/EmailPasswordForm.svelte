<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import Eye from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import { createEventDispatcher, onMount } from "svelte";
  import { ArrowLeftIcon } from "lucide-svelte";
  import { WebAuth } from "auth0-js";
  import { DATABASE_CONNECTION } from "../constants";
  import type { Auth0Error } from "auth0-js";
  import { AuthStep } from "../types";

  const dispatch = createEventDispatcher();

  export let disabled = false;
  export let email = "";
  export let isDomainDisabled = false;
  export let webAuth: WebAuth;
  export let step: AuthStep;

  let showForgetPassword = false;
  let password = "";
  let showPassword = false;
  let errorText = "";
  let inputEl: HTMLInputElement;

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-primary-ry-300 focus:outline-none";

  $: type = showPassword ? "text" : "password";

  function handleInput(event) {
    password = event.target.value;

    // Clear error text if password is cleared
    if (password.length === 0) {
      errorText = "";
    }
  }

  function togglePasswordVisibility() {
    showPassword = !showPassword;
  }

  function handleClick() {
    errorText = "";

    authenticateUser(email, password);
  }

  function displayError(err: any) {
    errorText = err.message;
  }

  function handleResetPassword() {
    errorText = "";

    if (!email) return displayError({ message: "Please enter an email" });
    if (isDomainDisabled) {
      return displayError({
        message: "Password reset is not available. Please contact your admin.",
      });
    }

    webAuth.changePassword(
      {
        connection: DATABASE_CONNECTION,
        email: email,
      },
      (err, resp) => {
        if (err) displayError({ message: err?.description });
        else alert(resp);
      },
    );
  }

  function handleAuthError(err: Auth0Error) {
    // NOTE: Auth0 is not consistent in the naming of the error description field
    const errorText =
      typeof err?.description === "string"
        ? err.description
        : typeof err?.policy === "string"
          ? err.policy
          : typeof err?.error_description === "string"
            ? err.error_description
            : err?.errorDescription;

    displayError({ message: errorText });
    disabled = false;
  }

  async function authenticateUser(email: string, password: string) {
    disabled = true;
    errorText = "";
    showForgetPassword = false;

    try {
      if (step === AuthStep.Login) {
        webAuth.login(
          {
            realm: DATABASE_CONNECTION,
            username: email,
            password: password,
          },
          (loginErr) => {
            if (loginErr) {
              displayError({ message: loginErr?.description });
              showForgetPassword = true;
            } else {
              disabled = false;
            }
          },
        );
      } else {
        webAuth.redirect.signupAndLogin(
          {
            connection: DATABASE_CONNECTION,
            email: email,
            password: password,
          },
          (err) => {
            if (err) {
              handleAuthError(err);
              showForgetPassword = true;
            } else {
              disabled = false;
            }
          },
        );
      }
    } catch (err) {
      handleAuthError(err);
      showForgetPassword = true;
    }
  }

  $: disabled = !(password.length > 0);

  onMount(() => {
    if (inputEl) {
      inputEl.focus();
    }
  });
</script>

<div class="flex flex-col gap-y-4" style:max-width="400px">
  <div>
    <div class="relative flex items-center">
      <input
        bind:this={inputEl}
        class="{inputClasses} {focusClasses} flex-grow pr-10"
        style:width="100%"
        {type}
        on:input={handleInput}
        id="password"
        placeholder="Password"
        required
        on:keydown={(e) => {
          if (e.key === "Enter") {
            handleClick();
          }
        }}
      />

      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <span
        role="button"
        tabindex="0"
        class="absolute right-3 cursor-pointer"
        on:click={togglePasswordVisibility}
      >
        {#if !showPassword}
          <Eye />
        {:else}
          <EyeInvisible />
        {/if}
      </span>
    </div>

    {#if errorText}
      <div class="text-red-500 text-sm mt-3">
        {errorText}
      </div>
    {/if}
  </div>

  {#if showForgetPassword}
    <div class="text-left">
      <button
        type="button"
        on:click={handleResetPassword}
        class="text-sm text-slate-500 pl-1 font-medium">Forgot password?</button
      >
    </div>
  {/if}

  <CtaButton {disabled} variant="primary" on:click={handleClick}>
    <div class="flex justify-center font-medium">Continue</div>
  </CtaButton>
  <CtaButton
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
