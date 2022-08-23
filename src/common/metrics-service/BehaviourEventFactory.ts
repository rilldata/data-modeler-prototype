import { MetricsEventFactory } from "$common/metrics-service/MetricsEventFactory";
import type {
  CommonFields,
  CommonUserFields,
} from "$common/metrics-service/MetricsTypes";
import type { BehaviourEvent } from "$common/metrics-service/BehaviourEventTypes";
import type { BehaviourEventMedium } from "$common/metrics-service/BehaviourEventTypes";
import type {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "$common/metrics-service/MetricsTypes";

export class BehaviourEventFactory extends MetricsEventFactory {
  public navigationEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      commonFields,
      commonUserFields
    ) as BehaviourEvent;
    event.medium = medium;
    event.space = space;
    event.screen_name = screen_name;
    return event;
  }
}
