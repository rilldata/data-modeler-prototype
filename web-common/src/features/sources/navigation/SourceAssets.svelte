<script lang="ts">
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import GenerateChartYAMLPrompt from "@rilldata/web-common/features/charts/prompt/GenerateChartYAMLPrompt.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { slide } from "svelte/transition";
  import { appScreen } from "../../../layout/app-store";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import NavigationEntry from "../../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../../layout/navigation/NavigationHeader.svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import AddAssetButton from "../../entity-management/AddAssetButton.svelte";
  import { EntityType } from "../../entity-management/types";
  import { createModelFromSource } from "../createModel";
  import SourceMenuItems from "./SourceMenuItems.svelte";
  import SourceTooltip from "./SourceTooltip.svelte";
  import { addSourceModal } from "../modal/add-source-visibility";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  let showTables = true;
  let showRenameTableModal = false;
  let renameTableName: null | string = null;

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: modelNames = useModelFileNames($runtime.instanceId);

  $: hasNoAssets = $sourceNames.data?.length === 0;

  async function openShowAddSourceModal() {
    addSourceModal.open();

    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceAdd,
      BehaviourEventMedium.Button,
      $appScreen.type,
      MetricsEventSpace.LeftPanel,
    );
  }

  async function queryHandler(tableName: string) {
    await createModelFromSource(
      $runtime.instanceId,
      $modelNames?.data ?? [],
      tableName,
      tableName,
    );
    // TODO: fire telemetry
  }

  function openRenameTableModal(tableName: string) {
    showRenameTableModal = true;
    renameTableName = tableName;
  }

  let showGenerateChartModal = false;
  let generateChartTable = "";
  let generateChartConnector = "";
  function openGenerateChartModal(tableName: string, connector: string) {
    showGenerateChartModal = true;
    generateChartTable = tableName;
    generateChartConnector = connector;
  }
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showTables}>Sources</NavigationHeader>

  {#if showTables}
    <ol transition:slide={{ duration }}>
      {#if $sourceNames?.data}
        {#each $sourceNames.data as sourceName (sourceName)}
          <li
            animate:flip={{ duration: 200 }}
            transition:slide={{ duration }}
            aria-label={sourceName}
          >
            <NavigationEntry
              expandable
              name={sourceName}
              context="source"
              open={$page.url.pathname === `/source/${sourceName}`}
              on:command-click={() => queryHandler(sourceName)}
            >
              <div slot="more" transition:slide={{ duration }}>
                <ColumnProfile indentLevel={1} objectName={sourceName} />
              </div>

              <SourceTooltip slot="tooltip-content" {sourceName} connector="" />

              <SourceMenuItems
                slot="menu-items"
                {sourceName}
                on:rename-asset={() => {
                  openRenameTableModal(sourceName);
                }}
                on:generate-chart={({ detail: { table, connector } }) => {
                  openGenerateChartModal(table, connector);
                }}
              />
            </NavigationEntry>
          </li>
        {/each}
      {/if}
      <AddAssetButton
        id="add-table"
        label="Add source"
        bold={hasNoAssets}
        on:click={openShowAddSourceModal}
      />
    </ol>
  {/if}
</div>

{#if showRenameTableModal && renameTableName !== null}
  <RenameAssetModal
    entityType={EntityType.Table}
    closeModal={() => (showRenameTableModal = false)}
    currentAssetName={renameTableName}
  />
{/if}

{#if showGenerateChartModal}
  <GenerateChartYAMLPrompt
    bind:open={showGenerateChartModal}
    table={generateChartTable}
    connector={generateChartConnector}
  />
{/if}
