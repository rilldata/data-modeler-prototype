import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useMainEntityFiles } from "../entity-management/file-selectors";

export function useChartFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "charts");
}

export const useChart = (instanceId: string, chartName: string) => {
  return useResource(instanceId, chartName, ResourceKind.Chart);
};
