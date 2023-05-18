import { describe, it } from "vitest";
import {
  AdBidsBaseFilter,
  AdBidsBidPriceMeasure,
  AdBidsClearedFilter,
  AdBidsDomainDimension,
  AdBidsExcludedFilter,
  AdBidsMirrorName,
  AdBidsName,
  AdBidsPublisherDimension,
  AllTimeParsedTestControls,
  assertMetricsView,
  createAdBidsInStore,
  createAdBidsMirrorInStore,
  CustomTestControls,
  Last6HoursTestControls,
  Last6HoursTestParsedControls,
} from "@rilldata/web-common/features/dashboard-stores-test-data";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";

describe("dashboard-stores", () => {
  it("Toggle filters", () => {
    createAdBidsInStore();
    assertMetricsView(AdBidsName);

    // add filters
    metricsExplorerStore.toggleFilter(
      AdBidsName,
      AdBidsPublisherDimension,
      "Google"
    );
    metricsExplorerStore.toggleFilter(
      AdBidsName,
      AdBidsPublisherDimension,
      "Facebook"
    );
    metricsExplorerStore.toggleFilter(
      AdBidsName,
      AdBidsDomainDimension,
      "google.com"
    );
    assertMetricsView(AdBidsName, AdBidsBaseFilter);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AdBidsMirrorName,
      AdBidsBaseFilter,
      AllTimeParsedTestControls
    );

    // toggle to exclude
    metricsExplorerStore.toggleFilterMode(AdBidsName, AdBidsPublisherDimension);
    assertMetricsView(AdBidsName, AdBidsExcludedFilter);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AdBidsMirrorName,
      AdBidsExcludedFilter,
      AllTimeParsedTestControls
    );

    // clear for Domain
    metricsExplorerStore.clearFilterForDimension(
      AdBidsName,
      AdBidsDomainDimension,
      true
    );
    assertMetricsView(AdBidsName, AdBidsClearedFilter);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AdBidsMirrorName,
      AdBidsClearedFilter,
      AllTimeParsedTestControls
    );

    // clear
    metricsExplorerStore.clearFilters(AdBidsName);
    assertMetricsView(AdBidsName);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(AdBidsMirrorName, undefined, AllTimeParsedTestControls);
  });

  it("Update time selections", () => {
    createAdBidsInStore();
    assertMetricsView(AdBidsName);

    // select a different time
    metricsExplorerStore.setSelectedTimeRange(
      AdBidsName,
      Last6HoursTestControls
    );
    assertMetricsView(AdBidsName, undefined, Last6HoursTestControls);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore();
    // start and end are not persisted
    assertMetricsView(
      AdBidsMirrorName,
      undefined,
      Last6HoursTestParsedControls
    );

    // select custom time
    metricsExplorerStore.setSelectedTimeRange(AdBidsName, CustomTestControls);
    assertMetricsView(AdBidsName, undefined, CustomTestControls);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore();
    // start and end are persisted for custom
    assertMetricsView(AdBidsMirrorName, undefined, CustomTestControls);
  });

  it("Select different measure", () => {
    createAdBidsInStore();
    assertMetricsView(AdBidsName);

    // select a different leaderboard measure
    metricsExplorerStore.setLeaderboardMeasureName(
      AdBidsName,
      AdBidsBidPriceMeasure
    );
    assertMetricsView(AdBidsName, undefined, undefined, AdBidsBidPriceMeasure);

    // create a mirror using the proto and assert that the leaderboard measure is persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AdBidsMirrorName,
      undefined,
      AllTimeParsedTestControls,
      AdBidsBidPriceMeasure
    );
  });
});
