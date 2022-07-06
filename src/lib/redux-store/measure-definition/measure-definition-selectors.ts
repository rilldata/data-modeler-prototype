import { generateFilteredEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export const {
  singleSelector: selectMeasureById,
  manySelector: selectMeasuresByMetricsId,
  manySelectorByIds: selectMeasuresByIds,
} = generateFilteredEntitySelectors<[string], MeasureDefinitionEntity>(
  "measureDefinition",
  (entity: MeasureDefinitionEntity, metricsDefId: string) =>
    entity.metricsDefId === metricsDefId
);
