<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import auth0, { WebAuth } from "auth0-js";
  import type { AuthOptions } from "auth0-js";
  import { onMount } from "svelte";
  import { LOGIN_OPTIONS } from "../config";
  import AuthContainer from "./AuthContainer.svelte";
  import EmailPasswordForm from "./EmailPasswordForm.svelte";
  import { getConnectionFromEmail } from "./utils";
  import OrSeparator from "./OrSeparator.svelte";
  import EmailSubmissionForm from "./EmailSubmissionForm.svelte";
  import Disclaimer from "./Disclaimer.svelte";
  import Spacer from "./Spacer.svelte";
  import { AuthStep, type Config } from "../types";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";

  export let configParams: string;
  export let cloudClientIDs = "";
  export let disableForgotPassDomains = "";
  export let connectionMap = "{}";
  export let auth0Domain = "";
  export let auth0ClientID = "";
  export let auth0ClientSecret = "";

  const connectionMapObj = JSON.parse(connectionMap);
  const cloudClientIDsArr = cloudClientIDs.split(",");
  const disableForgotPassDomainsArr = disableForgotPassDomains.split(",");
  const AUTH0_DOMAIN = auth0Domain;
  const AUTH0_CLIENT_ID = auth0ClientID;
  const AUTH0_CLIENT_SECRET = auth0ClientSecret;

  let email = "";
  let step: AuthStep = AuthStep.Base;
  let webAuth: WebAuth;
  let isExistingUser = false;
  let tokenExpiry = 0;
  let managementApiToken = null;

  $: errorText = "";
  $: isAllowedClient = false;
  $: domainDisabled = isDomainDisabled(email);
  $: headingText = getHeadingText(step, isExistingUser, email);
  $: subheadingText = getSubheadingText(step, email);

  function isDomainDisabled(email: string): boolean {
    return disableForgotPassDomainsArr.some((domain) =>
      email.toLowerCase().endsWith(domain.toLowerCase()),
    );
  }

  function configureDevMode() {
    if (
      import.meta.env.DEV &&
      (!configParams || configParams === "undefined")
    ) {
      console.warn(
        "No auth config provided. In development mode - auth flows will not work.",
      );
      errorText = "Authentication is not configured in development mode";
      isAllowedClient = true;

      step = AuthStep.Base;
      return;
    }
  }

  function init() {
    try {
      configureDevMode();

      const config = JSON.parse(
        decodeURIComponent(escape(window.atob(configParams))),
      ) as Config;

      if (!cloudClientIDsArr.includes(config?.clientID)) {
        errorText = "Authentication is not available for this client";
        isAllowedClient = false;
        return;
      }
      isAllowedClient = true;

      const isSignup = config?.extraParams?.screen_hint === "signup";

      if (isSignup) {
        step = AuthStep.SignUp;
      }

      const authOptions: AuthOptions = Object.assign(
        {
          overrides: {
            __tenant: config.auth0Tenant,
            __token_issuer: config.authorizationServer.issuer,
          },
          domain: config.auth0Domain,
          clientID: config.clientID,
          redirectUri: config.callbackURL,
          responseType: "code",
        },
        config.internalOptions,
      );

      webAuth = new auth0.WebAuth(authOptions);
    } catch (e) {
      console.error("Failed to initialize auth:", e);
      errorText = "Failed to initialize authentication in development mode";
    }
  }

  function authorizeSSO(email: string, connectionName: string) {
    if (import.meta.env.DEV) {
      errorText = "SSO authentication is not available in development mode";
      return;
    }

    webAuth.authorize({
      connection: connectionName,
      login_hint: email,
      prompt: "login",
    });
  }

  // TODO: Serverless function
  async function getManagementApiToken() {
    const now = Math.floor(Date.now() / 1000);

    if (managementApiToken && now < tokenExpiry) {
      return managementApiToken;
    }

    try {
      const response = await fetch(`https://${AUTH0_DOMAIN}/oauth/token`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type",
        },
        body: JSON.stringify({
          client_id: AUTH0_CLIENT_ID,
          client_secret: AUTH0_CLIENT_SECRET,
          audience: `https://${AUTH0_DOMAIN}/api/v2/`,
          grant_type: "client_credentials",
        }),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to obtain management API token: ${response.statusText}`,
        );
      }

      const data = await response.json();
      managementApiToken = data.access_token;
      tokenExpiry = now + data.expires_in;

      return managementApiToken;
    } catch (error) {
      console.error("Error obtaining management API token:", error);
      step = AuthStep.Base;
      errorText = "Unable to verify user existence. Please try again.";
      throw error;
    }
  }

  // TODO: Serverless function
  async function checkUserExists(email: string) {
    if (import.meta.env.DEV) {
      errorText = "User existence check is not available in development mode";
      return;
    }

    try {
      const token = await getManagementApiToken();
      // TODO: use our own API
      const response = await fetch(
        `https://${AUTH0_DOMAIN}/api/v2/users-by-email?email=${encodeURIComponent(email)}&fields=email`,
        {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error(
          `Failed to check user existence: ${response.statusText}`,
        );
      }

      const users = await response.json();
      isExistingUser = users.length > 0;
      console.log("User existence check:", { isExistingUser });

      step = isExistingUser ? AuthStep.Login : AuthStep.SignUp;
    } catch (error) {
      console.error("Error checking user existence:", error);
      errorText = "Unable to verify user existence. Please try again.";
      step = AuthStep.Base;
    }
  }

  async function processEmailSubmission(event) {
    errorText = "";
    isExistingUser = false;
    email = event.detail.email;
    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    step = AuthStep.Loading;
    headingText = getHeadingText(step, isExistingUser, email);
    subheadingText = getSubheadingText(step, email);

    if (connectionName) {
      authorizeSSO(email, connectionName);
    } else {
      await checkUserExists(email);

      console.log("after checking: ", email, isExistingUser);

      if (isExistingUser) {
        step = AuthStep.Login;
      } else {
        step = AuthStep.SignUp;
      }

      headingText = getHeadingText(step, isExistingUser, email);
      subheadingText = getSubheadingText(step, email);
    }
  }

  function getHeadingText(
    step: AuthStep,
    isExisting: boolean,
    email: string,
  ): string {
    switch (step) {
      case AuthStep.Base:
        return "Continue to Rill";
      case AuthStep.Login:
        return "Log in with email";
      case AuthStep.SignUp:
        return isExisting
          ? `Welcome back to Rill`
          : `Sign up with <span class="font-medium">${email}</span>`;
      case AuthStep.Loading:
        return "Checking...";
      default:
        return "";
    }
  }

  function getSubheadingText(step: AuthStep, email: string): string {
    switch (step) {
      case AuthStep.Login:
        return `Log in using <span class="font-medium">${email}</span>`;
      case AuthStep.Loading:
        return "";
      default:
        return "";
    }
  }

  function backToBaseStep() {
    step = AuthStep.Base;
    errorText = "";
    isExistingUser = false;
  }

  onMount(() => {
    if (import.meta.env.DEV) {
      errorText = "Unable to initialize auth0 client in development mode";
      return;
    }

    init();
  });
</script>

<AuthContainer>
  <RillLogoSquareNegative size="84px" />
  <Spacer />
  <div class="flex flex-col items-center gap-y-2 text-center">
    <div class="text-xl text-slate-800">
      {@html headingText}
    </div>
    {#if subheadingText}
      <div class="text-base text-gray-500">
        {@html subheadingText}
      </div>
    {:else}
      <Spacer />
    {/if}
  </div>

  <div class="flex flex-col gap-y-4 mt-6" style:width="400px">
    {#if step === AuthStep.Base}
      {#each LOGIN_OPTIONS as { label, icon, style, connection } (connection)}
        <CtaButton
          variant={style === "primary" ? "primary" : "secondary"}
          on:click={() => {
            errorText = "";
            if (import.meta.env.DEV) {
              errorText =
                "OAuth authentication is not available in development mode";
              return;
            }
            webAuth.authorize({ connection });
          }}
        >
          <div class="flex justify-center items-center gap-x-2 font-medium">
            {#if icon}
              <svelte:component this={icon} />
            {/if}
            <div>{label}</div>
          </div>
        </CtaButton>
      {/each}

      <OrSeparator />

      <EmailSubmissionForm on:submit={processEmailSubmission} />
    {/if}

    {#if step === AuthStep.Login || step === AuthStep.SignUp}
      <EmailPasswordForm
        {email}
        {step}
        isDomainDisabled={domainDisabled}
        {webAuth}
        on:back={backToBaseStep}
        disabled={!isAllowedClient}
      />
    {/if}
  </div>

  {#if errorText}
    <div style:max-width="400px" class="text-red-500 text-sm mt-3">
      {errorText}
    </div>
  {/if}

  <Disclaimer />
  <div class="mt-6 text-center">
    <CtaNeedHelp />
  </div>
</AuthContainer>
