import { page } from "$app/stores";
import { getScreenNameFromPage } from "@rilldata/web-admin/features/navigation/nav-utils";
import { RillAdminTelemetryClient } from "@rilldata/web-admin/features/telemetry/RillAdminTelemetryClient";
import { collectCommonUserFields } from "@rilldata/web-common/metrics/collectCommonUserFields";
import { ErrorEventHandler } from "@rilldata/web-common/metrics/ErrorEventHandler";
import {
  metricsService,
  setErrorEvent,
  setMetricsService,
} from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventFactory } from "@rilldata/web-common/metrics/service/BehaviourEventFactory";
import { ErrorEventFactory } from "@rilldata/web-common/metrics/service/ErrorEventFactory";
import { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-common/metrics/service/ProductHealthEventFactory";
import { get } from "svelte/store";

export async function initCloudMetrics() {
  setMetricsService(
    new MetricsService(new RillAdminTelemetryClient(), [
      new ProductHealthEventFactory(),
      new BehaviourEventFactory(),
      new ErrorEventFactory(),
    ]),
  );

  const commonUserMetrics = await collectCommonUserFields();
  setErrorEvent(
    new ErrorEventHandler(
      metricsService,
      commonUserMetrics,
      window.location.host.startsWith("localhost"),
      () => getScreenNameFromPage(get(page)),
    ),
  );
  // TODO: add other handlers and callers
}
