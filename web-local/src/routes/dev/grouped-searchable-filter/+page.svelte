<script lang="ts">
  import SearchableFilterDropdown from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterDropdown.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";

  const selectableGroups: Array<SearchableFilterSelectableGroup> = [
    {
      name: "Measures",
      items: [
        { name: "total_records", label: "Total Records" },
        { name: "bid_price", label: "Bid Price" },
      ],
    },
    {
      name: "Dimensions",
      items: [
        { name: "publisher", label: "Publisher" },
        { name: "domain", label: "Domain" },
      ],
    },
  ];

  let singleSelected = selectableGroups.map((s) => s.items.map(() => false));
  const multiSelected = selectableGroups.map((s) => s.items.map(() => false));
</script>

<span class="p-2">Single Select</span>
<SearchableFilterDropdown
  allowMultiSelect={false}
  on:focus
  on:hover
  on:item-clicked={(e) => {
    singleSelected = selectableGroups.map((s) => s.items.map(() => false));
    for (let i = 0; i < selectableGroups.length; i++) {
      for (let j = 0; j < selectableGroups[i].items.length; j++) {
        if (selectableGroups[i].items[j].name === e.detail.name) {
          singleSelected[i][j] = !singleSelected[i][j];
          return;
        }
      }
    }
  }}
  {selectableGroups}
  selectedItems={singleSelected}
/>
<br />
<span class="p-2">Multi Select</span>
<SearchableFilterDropdown
  allowMultiSelect={true}
  on:focus
  on:hover
  on:item-clicked={(e) => {
    for (let i = 0; i < selectableGroups.length; i++) {
      for (let j = 0; j < selectableGroups[i].items.length; j++) {
        if (selectableGroups[i].items[j].name === e.detail.name) {
          multiSelected[i][j] = !multiSelected[i][j];
          return;
        }
      }
    }
  }}
  {selectableGroups}
  selectedItems={multiSelected}
/>
<br />
<span class="p-2">Single Select No Icon</span>
<SearchableFilterDropdown
  allowMultiSelect={false}
  on:focus
  on:hover
  on:item-clicked={(e) => {
    singleSelected = selectableGroups.map((s) => s.items.map(() => false));
    for (let i = 0; i < selectableGroups.length; i++) {
      for (let j = 0; j < selectableGroups[i].items.length; j++) {
        if (selectableGroups[i].items[j].name === e.detail.name) {
          singleSelected[i][j] = !singleSelected[i][j];
          return;
        }
      }
    }
  }}
  {selectableGroups}
  selectedItems={singleSelected}
  showSelection={false}
/>
