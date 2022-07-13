import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";

import TableCellInput from "$lib/components/table-editable/TableCellInput.svelte";
import TableCellSelector from "../table-editable/TableCellSelector.svelte";

export const initDimensionColumns = (inputChangeHandler, dimensionOptions) =>
  [
    {
      name: "labelSingle",
      label: "label (single)",
      tooltip: "a human readable name for this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
    },

    {
      name: "dimensionColumn",
      label: "dimension column",
      tooltip:
        "a categorical column from the data model that this metrics set is based on",
      renderer: TableCellSelector,
      onchange: inputChangeHandler,
      options: dimensionOptions,
      validation: (row: DimensionDefinitionEntity) => row.dimensionIsValid,
    },
    {
      name: "description",
      tooltip: "a human readable description of this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
    },

    {
      name: "labelPlural",
      label: "label (plural)",
      tooltip: "an optional pluralized human readable name for this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
    },
    // FIXME will be needed later for API
    // {
    //   name: "sqlName",
    //   label: "identifier",
    //   tooltip: "a unique SQL identifier for this dimension",
    //   renderer: TableCellInput,
    //   onchange: inputChangeHandler,
    //   validation: (row: DimensionDefinitionEntity) => row.sqlNameIsValid,
    // },

    // FIXME: willbe needed later for cardinality summary
    // {
    //   name: "id",
    //   label: "unique values",
    //   tooltip: "the number of unique values present in this dimension",
    //   renderer: TabelCellCardinality
    // },
  ] as ColumnConfig[];
