import { MetricsEventFactory } from "$common/metrics-service/MetricsEventFactory";
import type { ActiveEvent, CommonFields, CommonUserFields } from "$common/metrics-service/MetricsTypes";

export class ProductHealthEventFactory extends MetricsEventFactory {
    public activeEvent(commonFields: CommonFields, commonUserFields: CommonUserFields,
                       durationMilSec: number, totalInFocus: number): ActiveEvent {
        const event = this.getBaseMetricsEvent("active", commonFields, commonUserFields) as ActiveEvent;
        event.duration_sec = Math.round(durationMilSec / 1000);
        event.total_in_focus = totalInFocus;
        return event;
    }
}
