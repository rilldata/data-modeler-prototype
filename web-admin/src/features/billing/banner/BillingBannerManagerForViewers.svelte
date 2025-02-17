<script lang="ts">
  import { areAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;

  $: allProjectsHibernating = areAllProjectsHibernating(organization);

  $: if ($allProjectsHibernating.data) {
    // we have a generic banner for viewers when org is defunct for some reason and projects are hibernating
    eventBus.emit("banner", {
      type: "default",
      message:
        "This org’s projects are hibernating. Please reach out to your administrator to regain access.",
      iconType: "sleep",
    });
  }
</script>
