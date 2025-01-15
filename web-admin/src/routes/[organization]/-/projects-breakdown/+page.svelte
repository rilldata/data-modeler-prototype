<script lang="ts">
  import ProjectBreakdownDashboard from "@rilldata/web-admin/features/organizations/project-breakdown/ProjectBreakdownDashboard.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ runtime } = data);
  // Seeing the data using the same project but with different jwt will not get automatically invalidated.
  // Since we do not have jwt as part of the query key we need to invalidate the queries for this instanceId
  $: invalidateRuntimeQueries(queryClient, runtime.instanceId);
</script>

<RuntimeProvider
  instanceId={runtime.instanceId}
  host={runtime.host}
  jwt={runtime.jwt.token}
  authContext={runtime.jwt.authContext}
>
  <ProjectBreakdownDashboard />
</RuntimeProvider>
