<script lang="ts">
  import Auth from "./components/Auth.svelte";

  const connectionMap = import.meta.env.VITE_CONNECTION_MAP;
  const cloudClientIDs = import.meta.env.VITE_RILL_CLOUD_AUTH0_CLIENT_IDS;
  const disableForgotPassDomains = import.meta.env
    .VITE_DISABLE_FORGOT_PASS_DOMAINS;
  const auth0Domain = import.meta.env.VITE_AUTH0_DOMAIN;
  const auth0ClientID = import.meta.env.VITE_AUTH0_CLIENT_ID;
  const auth0ClientSecret = import.meta.env.VITE_AUTH0_CLIENT_SECRET;

  // This gets populated by Auth0 runtime
  const configParams = "@@config@@";
</script>

<svelte:head>
  {#if import.meta.env.PROD}
    <meta
      http-equiv="Content-Security-Policy"
      content="
        default-src 'none';
        connect-src 'self' https:;
        font-src https:;
        img-src https:;
        object-src 'none';
        script-src https: 'unsafe-inline';
        style-src 'unsafe-inline' https:
        "
    />
  {/if}
</svelte:head>

<main class="size-full">
  <Auth
    {configParams}
    {cloudClientIDs}
    {disableForgotPassDomains}
    {connectionMap}
    {auth0Domain}
    {auth0ClientID}
    {auth0ClientSecret}
  />
</main>
