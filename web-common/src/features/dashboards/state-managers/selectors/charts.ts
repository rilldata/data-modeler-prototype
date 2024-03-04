import { timeControlsState } from "@rilldata/web-common/features/dashboards/state-managers/selectors/time-range";
import {
  getDurationFromMS,
  getOffset,
  getTimeWidth,
} from "@rilldata/web-common/lib/time/transforms";
import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
import type { DashboardDataSources } from "./types";

export const chartSelectors = {
  canPanLeft: (dashData: DashboardDataSources) => {
    const timeControls = timeControlsState(dashData);
    const startRange = timeControls.allTimeRange?.start;
    const selectedStart = timeControls.selectedTimeRange?.start;
    return (
      (selectedStart?.getTime() || Infinity) >=
      (startRange?.getTime() || -Infinity)
    );
  },
  canPanRight: (dashData: DashboardDataSources) => {
    const timeControls = timeControlsState(dashData);
    const endRange = timeControls?.allTimeRange?.end;
    const selectedEnd = timeControls.selectedTimeRange?.end;
    return (
      (selectedEnd?.getTime() || -Infinity) <= (endRange?.getTime() || Infinity)
    );
  },
  getNewPanRage: (dashData: DashboardDataSources) => {
    const timeControls = timeControlsState(dashData);

    return (direction: "left" | "right") => {
      const selectedTimeRange = timeControls?.selectedTimeRange;
      if (!selectedTimeRange) return;

      const timeZone = dashData?.dashboard?.selectedTimezone || "UTC";
      const { start, end } = selectedTimeRange;

      if (!start || !end) return;

      const offsetType =
        direction === "left" ? TimeOffsetType.SUBTRACT : TimeOffsetType.ADD;

      const currentRangeWidth = getTimeWidth(start, end);
      const panAmount = getDurationFromMS(currentRangeWidth);

      const newStart = getOffset(start, panAmount, offsetType, timeZone);
      const newEnd = getOffset(end, panAmount, offsetType, timeZone);

      return { start: newStart, end: newEnd };
    };
  },
};
