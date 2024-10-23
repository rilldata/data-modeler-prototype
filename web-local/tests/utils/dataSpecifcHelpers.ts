import type { Page } from "playwright";
import { assertLeaderboards } from "./metricsViewHelpers";
import {
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "./commonHelpers";
import { createModel } from "./modelHelpers";
import { uploadFile, waitForSource } from "./sourceHelpers";

export const AD_BIDS_METRICS_PATH = "/metrics/AdBids_model_metrics.yaml";
export const AD_BIDS_EXPLORE_PATH =
  "/dashboards/AdBids_model_metrics_explore.yaml";

export async function createAdBidsModel(page: Page) {
  await Promise.all([
    waitForSource(page, "/sources/AdBids.yaml", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    uploadFile(page, "AdBids.csv"),
  ]);

  await createModel(page, "AdBids_model.sql");
  await Promise.all([
    waitForProfiling(page, "AdBids_model", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    updateCodeEditor(page, `select * from "AdBids"`),
  ]);
}

export async function assertAdBidsDashboard(page: Page) {
  await wrapRetryAssertion(() =>
    assertLeaderboards(page, [
      {
        label: "Publisher",
        values: ["null", "Facebook", "Google", "Yahoo", "Microsoft"],
      },
      {
        label: "Domain",
        values: [
          "facebook.com",
          "msn.com",
          "google.com",
          "news.yahoo.com",
          "instagram.com",
          "sports.yahoo.com",
          "news.google.com",
        ],
      },
    ]),
  );
  // TODO: how do we assert timeseries?
}
