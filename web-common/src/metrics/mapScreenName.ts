import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Page } from "@sveltejs/kit";

/**
 * mapScreenName maps page route id to {@link MetricsEventScreenName}
 * This is only for web-local. But since {@link initMetrics} is in web-common this has to be here for now.
 */
export function mapScreenName(page: Page): MetricsEventScreenName {
  switch (page.route.id) {
    case "/(application)/(developer)/files/dashboards/[name]":
      return MetricsEventScreenName.Dashboard;
    case "/(application)/(developer)/files/metrics/[name]/edit":
      return MetricsEventScreenName.MetricsDefinition;
    case "/(application)/(developer)/files/sources/[name]":
      return MetricsEventScreenName.Source;
    case "/(application)/(developer)/files/models/[name]":
      return MetricsEventScreenName.Model;
    case "/(application)/(developer)":
      return MetricsEventScreenName.Home;
  }
  return MetricsEventScreenName.Unknown;
}
