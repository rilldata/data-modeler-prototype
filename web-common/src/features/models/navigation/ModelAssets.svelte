<script lang="ts">
  import { page } from "$app/stores";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import GenerateChartYAMLPrompt from "@rilldata/web-common/features/charts/prompt/GenerateChartYAMLPrompt.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { useSourceFileNames } from "@rilldata/web-common/features/sources/selectors";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import NavigationEntry from "../../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import AddAssetButton from "../../entity-management/AddAssetButton.svelte";
  import { getName } from "../../entity-management/name-utils";
  import { createModel } from "../createModel";
  import ModelMenuItems from "./ModelMenuItems.svelte";
  import ModelTooltip from "./ModelTooltip.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let assets: V1Resource[];

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: useModelNames = useModelFileNames($runtime.instanceId);

  $: modelNames = $useModelNames?.data ?? [];
  let showModels = true;

  async function handleAddModel() {
    await createModel($runtime.instanceId, getName("model", modelNames));
    // if the models are not visible in the assets list, show them.
    if (!showModels) {
      showModels = true;
    }
  }

  /** rename the model */
  let showRenameModelModal = false;
  let renameModelName: string | null = null;
  const openRenameModelModal = (modelName: string) => {
    showRenameModelModal = true;
    renameModelName = modelName;
  };

  $: hasSourceButNoModels =
    $sourceNames?.data?.length !== undefined &&
    $sourceNames?.data?.length > 0 &&
    modelNames.length === 0;

  let showGenerateChartModal = false;
  let generateChartTable = "";
  let generateChartConnector = "";
  function openGenerateChartModal(tableName: string, connector: string) {
    showGenerateChartModal = true;
    generateChartTable = tableName;
    generateChartConnector = connector;
  }
</script>

<div class="flex flex-col h-fit gap-0">
  <NavigationHeader bind:show={showModels}>Models</NavigationHeader>

  {#if showModels}
    <ol transition:slide={{ duration }} id="assets-model-list">
      {#each assets as asset (asset.meta?.name?.name)}
        {@const modelName = asset.meta?.name?.name}
        <li animate:flip={{ duration }} aria-label={modelName}>
          <NavigationEntry
            expandable
            name={modelName}
            context="model"
            open={$page.url.pathname === `/model/${modelName}`}
          >
            <div transition:slide={{ duration }} slot="more">
              <ColumnProfile indentLevel={1} objectName={modelName} />
            </div>

            <ModelTooltip slot="tooltip-content" {modelName} />

            <ModelMenuItems
              slot="menu-items"
              {modelName}
              on:rename-asset={() => {
                openRenameModelModal(modelName);
              }}
              on:generate-chart={({ detail: { table, connector } }) => {
                openGenerateChartModal(table, connector);
              }}
            />
          </NavigationEntry>
        </li>
      {/each}
      <AddAssetButton
        id="create-model-button"
        label="Add model"
        bold={hasSourceButNoModels}
        on:click={handleAddModel}
      />
    </ol>
  {/if}
</div>

{#if showRenameModelModal && renameModelName !== null}
  <RenameAssetModal
    entityType={EntityType.Model}
    closeModal={() => (showRenameModelModal = false)}
    currentAssetName={renameModelName.replace(".sql", "")}
  />
{/if}

{#if showGenerateChartModal}
  <GenerateChartYAMLPrompt
    bind:open={showGenerateChartModal}
    table={generateChartTable}
    connector={generateChartConnector}
  />
{/if}
