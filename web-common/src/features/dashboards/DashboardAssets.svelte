<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import GenerateChartYAMLPrompt from "@rilldata/web-common/features/charts/prompt/GenerateChartYAMLPrompt.svelte";
  import DashboardMenuItems from "@rilldata/web-common/features/dashboards/DashboardMenuItems.svelte";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import { lastVisitedURLs } from "@rilldata/web-common/layout/navigation/last-visited-urls.js";
  import RenameAssetModal from "../entity-management/RenameAssetModal.svelte";

  $: instanceId = $runtime.instanceId;

  $: sourceNames = useSourceFileNames(instanceId);
  $: modelNames = useModelFileNames(instanceId);
  $: dashboardNames = useDashboardFileNames(instanceId);

  const createDashboard = createRuntimeServicePutFile();

  const { readOnly } = featureFlags;

  let showMetricsDefs = true;

  let showRenameMetricsDefinitionModal = false;
  let renameMetricsDefName: string | null = null;

  const openRenameMetricsDefModal = (metricsDefName: string) => {
    showRenameMetricsDefinitionModal = true;
    renameMetricsDefName = metricsDefName;
  };

  const dispatchAddEmptyMetricsDef = async () => {
    if (!showMetricsDefs) {
      showMetricsDefs = true;
    }
    const newDashboardName = getName("dashboard", $dashboardNames?.data ?? []);
    await $createDashboard.mutateAsync({
      instanceId,
      path: getFileAPIPathFromNameAndType(
        newDashboardName,
        EntityType.MetricsDefinition,
      ),
      data: {
        blob: "",
        create: true,
        createOnly: true,
      },
    });

    await goto(`/dashboard/${newDashboardName}`);
  };

  $: canAddDashboard = $readOnly === false;

  $: hasSourceAndModelButNoDashboards =
    $sourceNames?.data &&
    $modelNames?.data &&
    $sourceNames?.data?.length > 0 &&
    $modelNames?.data?.length > 0 &&
    $dashboardNames?.data?.length === 0;

  let showGenerateChartModal = false;
  let generateChartMetricsView = "";
  function openGenerateChartModal(metricsView: string) {
    showGenerateChartModal = true;
    generateChartMetricsView = metricsView;
  }

  const lastVisited = new Map();
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showMetricsDefs}>Dashboards</NavigationHeader>
  {#if showMetricsDefs}
    <ol transition:slide={{ duration }} id="assets-metrics-list">
      {#if $dashboardNames.data}
        {#each $dashboardNames.data as dashboardName (dashboardName)}
          <li animate:flip={{ duration }} aria-label={dashboardName}>
            <NavigationEntry
              showContextMenu={!$readOnly}
              name={dashboardName}
              context="dashboard"
              open={$page.url.pathname === `/dashboard/${dashboardName}` ||
                $page.url.pathname === `/dashboard/${dashboardName}/edit`}
            >
              <DashboardMenuItems
                slot="menu-items"
                metricsViewName={dashboardName}
                on:rename={() => openRenameMetricsDefModal(dashboardName)}
                on:generate-chart={() => openGenerateChartModal(dashboardName)}
              />
            </NavigationEntry>
          </li>
        {/each}
        {#if canAddDashboard}
          <AddAssetButton
            id="add-dashboard"
            label="Add dashboard"
            bold={hasSourceAndModelButNoDashboards ?? false}
            on:click={() => dispatchAddEmptyMetricsDef()}
          />
        {/if}
      {/if}
    </ol>
  {/if}
</div>

{#if showRenameMetricsDefinitionModal && renameMetricsDefName}
  <RenameAssetModal
    entityType={EntityType.MetricsDefinition}
    closeModal={() => (showRenameMetricsDefinitionModal = false)}
    currentAssetName={renameMetricsDefName}
  />
{/if}

{#if showGenerateChartModal}
  <GenerateChartYAMLPrompt
    bind:open={showGenerateChartModal}
    metricsView={generateChartMetricsView}
  />
{/if}
