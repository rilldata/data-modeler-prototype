<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";

  const options = "abcdef".split("").map((value) => ({ value, label: value }));
  let values: string[] = [];
  let value: string;

  $: console.log(values, value);
</script>

<Select {options} bind:value id="test-single" label="Single" />
<Select {options} bind:value={values} id="test-multi" label="Multi" multiple />
