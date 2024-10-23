import { waitUntil } from "@rilldata/utils";
import type { Page } from "playwright";
import { updateCodeEditor } from "./commonHelpers";

const ResourceWatcherLogRegex = /^\[(.*)] rill\.runtime\.v1\.(.*)\/(.*)$/;

export class ResourceWatcher {
  private statuses = new Map<string, string>();

  public constructor(private readonly page: Page) {
    page.on("console", (e) => {
      const matches = ResourceWatcherLogRegex.exec(e.text());
      if (!matches || matches.length < 4) return;
      const [, status, type, name] = matches;
      this.statuses.set(`${type}__${name}`, status);
    });
  }

  private async waitForResource(
    type: string,
    name = "AdBids_model_metrics_explore",
  ) {
    return waitUntil(
      () => this.statuses.get(`${type}__${name}`) === "RECONCILE_STATUS_IDLE",
      2000,
    );
  }

  public async updateAndWaitForExplore(
    code: string,
    name = "AdBids_model_metrics_explore",
  ) {
    this.statuses.delete(name); // clear older state
    return Promise.all([
      updateCodeEditor(this.page, code),
      this.waitForResource("rill.runtime.v1.Explore", name),
    ]);
  }

  public async updateAndWaitForDashboard(
    code: string,
    name = "AdBids_model_metrics",
  ) {
    this.statuses.delete(name); // clear older state
    return Promise.all([
      updateCodeEditor(this.page, code),
      this.waitForResource("rill.runtime.v1.MetricsView", name),
    ]);
  }
}
