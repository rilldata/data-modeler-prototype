<script lang="ts">
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import RadixH1 from "@rilldata/web-common/components/typography/RadixH1.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import Button from "../../components/button/Button.svelte";
</script>

<section class="flex flex-col gap-y-6 items-center text-center">
  <RillLogoSquareNegative size="84px" gradient />
  <RadixH1>
    <span
      class="bg-gradient-to-r from-primary-900 to-primary-800 text-transparent bg-clip-text opacity-75"
    >
      Welcome to Rill
    </span>
  </RadixH1>
  <div class="flex flex-col gap-y-2">
    <Subheading twColor="text-slate-600">
      Build fast operational dashboards that your team will actually use.
    </Subheading>
    <Subheading twColor="text-slate-600">Let's get started.</Subheading>
  </div>
  <Button type="primary" large href="/welcome/select-connectors">
    Connect your data
  </Button>
</section>
