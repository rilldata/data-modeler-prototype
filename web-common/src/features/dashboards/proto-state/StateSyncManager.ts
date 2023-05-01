import { goto } from "$app/navigation";
import { page } from "$app/stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { get } from "svelte/store";

export class StateSyncManager {
  private protoState: string;
  private urlState: string;
  private updating = false;

  public constructor(private readonly metricViewName: string) {}

  public handleStateChange(metricsExplorer: MetricsExplorerEntity) {
    const pageUrl = get(page).url;
    if (this.protoState === metricsExplorer.proto) return;
    this.protoState = metricsExplorer.proto;

    // if state didn't change do not call goto. this avoids adding unnecessary urls to history stack
    if (this.protoState !== this.urlState) {
      if (this.protoState === metricsExplorer.defaultProto) {
        goto(`${pageUrl.pathname}`);
      } else {
        goto(`${pageUrl.pathname}?state=${this.protoState}`);
      }
      this.updating = true;
    }
  }

  public handleUrlChange() {
    const pageUrl = get(page).url;
    const newUrlState = pageUrl.searchParams.get("state");
    if (this.urlState === newUrlState) return;
    this.urlState = newUrlState;

    // run sync if we didn't change the url through a state change
    // this can happen when url is updated directly by the user
    if (!this.updating && this.urlState && this.urlState !== this.protoState) {
      metricsExplorerStore.syncFromUrl(this.metricViewName, pageUrl);
    }
    this.updating = false;
  }
}
