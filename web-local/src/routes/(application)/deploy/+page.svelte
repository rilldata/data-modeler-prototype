<script lang="ts">
  import { goto } from "$app/navigation";
  import { getNeverSubscribedIssue } from "@rilldata/web-common/features/billing/issues";
  import TrialDetailsDialog from "@rilldata/web-common/features/billing/TrialDetailsDialog.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import OrgSelector from "@rilldata/web-common/features/project/OrgSelector.svelte";
  import { ProjectDeployer } from "@rilldata/web-common/features/project/ProjectDeployer";
  import DeployError from "@rilldata/web-common/features/project/DeployError.svelte";
  import { onMount } from "svelte";
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  const deployer = new ProjectDeployer(data.orgParam);
  const metadata = deployer.metadata;
  const user = deployer.user;
  const project = deployer.project;
  const orgsMetadata = deployer.orgsMetadata;
  const deployerStatus = deployer.getStatus();
  const promptOrgSelection = deployer.promptOrgSelection;
  // This org is set by the deployer.
  // 1. When there is no org is present it is auto created based on user's email.
  // 2. Otherwise, it will be the based on the selection. Will be equal to 'selectedOrg' in this case.
  const org = deployer.org;
  let isEmptyOrg = false;
  $: {
    const om = $orgsMetadata?.data?.orgs.find((o) => o.name === $org);
    isEmptyOrg = !!om?.issues && !!getNeverSubscribedIssue(om.issues);
  }

  // This is specifically the org selected using the OrgSelector.
  // Used to retrigger the deploy after the user confirms deploy on an empty org.
  let selectedOrg = "";
  let deployConfirmOpen = false;

  function onOrgSelect(org: string) {
    // Disabling for now. We should get user's quota and not show if it is hit
    // const om = orgMetadata.orgs.find((o) => o.name === org);
    // if (om?.issues && getNeverSubscribedIssue(om.issues)) {
    //   // if the selected org is empty, show the trial starting dialog
    //   // We need this since we wouldn't have shown it in DeployCTA
    //   deployConfirmOpen = true;
    //   selectedOrg = org;
    //   return;
    // }

    promptOrgSelection.set(false);
    return deployer.deploy(org);
  }

  function onStartTrialConfirm() {
    promptOrgSelection.set(false);
    return deployer.deploy(selectedOrg);
  }

  function onBack() {
    if ($orgsMetadata.data?.orgs?.length) {
      promptOrgSelection.set(true);
    } else {
      void goto("/");
    }
  }

  function onRetry() {
    return deployer.deploy($org);
  }

  onMount(() => {
    void deployer.loginOrDeploy();
  });
</script>

<!-- This seems to be necessary to trigger tanstack query to update the query object -->
<!-- TODO: find a config to avoid this -->
<div class="hidden">
  {$user.isLoading}-{$metadata.isLoading}-{$project.isLoading}
</div>

<CTALayoutContainer>
  <CTAContentContainer>
    {#if $promptOrgSelection}
      <OrgSelector
        orgs={$user.data?.rillUserOrgs ?? []}
        onSelect={onOrgSelect}
      />
    {:else if $deployerStatus.isLoading}
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      <CTAHeader variant="bold">
        Hang tight! We're deploying your project...
      </CTAHeader>
      <CTANeedHelp />
    {:else if $deployerStatus.error}
      <DeployError
        error={$deployerStatus.error}
        org={$org}
        adminUrl={$metadata.data?.adminUrl ?? ""}
        {isEmptyOrg}
        {onRetry}
        {onBack}
      />
    {/if}
  </CTAContentContainer>
</CTALayoutContainer>

<TrialDetailsDialog
  bind:open={deployConfirmOpen}
  onContinue={onStartTrialConfirm}
/>
