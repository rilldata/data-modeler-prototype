import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
import type { MetricsEvent } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { TelemetryClient } from "@rilldata/web-common/metrics/service/RillIntakeClient";

export class RillAdminTelemetryClient implements TelemetryClient {
  public async fireEvent(event: MetricsEvent) {
    try {
      const resp = await fetch(`${ADMIN_URL}/v1/telemetry/events`, {
        method: "POST",
        body: JSON.stringify({
          events: [event],
        }),
        credentials: "include",
      });
      if (!resp.ok)
        console.error(`Failed to send ${event.event_type}. ${resp.statusText}`);
    } catch (err) {
      console.error(`Failed to send ${event.event_type}. ${err.message}`);
    }
  }
}
