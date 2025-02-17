<script lang="ts">
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { createUserFacingError } from "../components/errors/user-facing-errors";

  $: ({ status, error } = $page); // Triggered by `throw error(status, message)`

  $: ({ statusCode, header, body, detail } = createUserFacingError(
    status,
    error.message,
  ));
</script>

<ErrorPage {statusCode} {header} {body} {detail} />
