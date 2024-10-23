import { page } from "$app/stores";
import type { ConnectError } from "@connectrpc/connect";
import { sanitizeName, waitUntil } from "@rilldata/utils";
import { extractDeployError } from "@rilldata/web-common/features/project/deploy-errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  GetCurrentProjectResponse,
  GetCurrentUserResponse,
  GetMetadataResponse,
} from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  createLocalServiceDeploy,
  createLocalServiceGetCurrentProject,
  createLocalServiceGetCurrentUser,
  createLocalServiceGetMetadata,
  createLocalServiceRedeploy,
  getLocalServiceGetCurrentUserQueryKey,
  localServiceGetCurrentUser,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, writable } from "svelte/store";

export class ProjectDeployer {
  public readonly metadata = createLocalServiceGetMetadata();
  public readonly user = createLocalServiceGetCurrentUser();
  public readonly project = createLocalServiceGetCurrentProject();
  public readonly promptOrgSelection = writable(false);

  private readonly deployMutation = createLocalServiceDeploy();
  private readonly redeployMutation = createLocalServiceRedeploy();

  public get isDeployed() {
    const projectResp = get(this.project).data as GetCurrentProjectResponse;
    return !!projectResp?.project;
  }

  public getStatus() {
    return derived(
      [
        this.metadata,
        this.user,
        this.project,
        this.deployMutation,
        this.redeployMutation,
      ],
      ([metadata, user, project, deployMutation, redeployMutation]) => {
        if (
          metadata.error ||
          user.error ||
          project.error ||
          deployMutation.error ||
          redeployMutation.error
        ) {
          return {
            isLoading: false,
            error: extractDeployError(
              (metadata.error as ConnectError) ??
                (user.error as ConnectError) ??
                (project.error as ConnectError) ??
                (deployMutation.error as ConnectError) ??
                (redeployMutation.error as ConnectError),
            ).message,
          };
        }

        // we can have periods where no mutation is firing, so it can flash to non-loading state
        // to avoid this we always have isLoading to true and only set to false on error
        // since after successful deploy the same page is redirected to cloud we wont have an issue of the final loaded state
        return {
          isLoading: true,
          error: undefined,
        };
      },
    );
  }

  public async loginOrDeploy() {
    await waitUntil(
      () => !get(this.metadata).isLoading && !get(this.user).isLoading,
    );

    const metadata = get(this.metadata).data as GetMetadataResponse;
    const userResp = get(this.user).data as GetCurrentUserResponse;
    if (!userResp.user) {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      const u = new URL(metadata.loginUrl);
      u.searchParams.set("redirect", get(page).url.toString());
      window.open(u.toString(), "_self");
    } else {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);
    }

    return this.deploy();
  }

  public async deploy(org?: string) {
    await waitUntil(() => !get(this.project).isLoading);

    const projectResp = get(this.project).data as GetCurrentProjectResponse;
    if (projectResp.project) {
      if (projectResp.project.githubUrl) {
        // we do not support pushing to a project already connected to github
        return;
      }

      const resp = await get(this.redeployMutation).mutateAsync({
        projectId: projectResp.project.id,
        reupload: true,
      });
      window.open(resp.frontendUrl, "_self");
      return;
    }

    let checkNextOrg = false;
    if (!org) {
      const { org: inferredOrg, checkNextOrg: inferredCheckNextOrg } =
        await this.inferOrg(get(this.user).data?.rillUserOrgs ?? []);
      // no org was inferred. right now this is because we have prompted the user for an org
      if (!inferredOrg) return;
      org = inferredOrg;
      checkNextOrg = inferredCheckNextOrg;
    }

    // hardcoded to upload for now
    const frontendUrl = await this.tryDeployWithOrg(
      org,
      projectResp.localProjectName,
      checkNextOrg,
    );
    window.open(frontendUrl + "/-/invite", "_self");
  }

  private async inferOrg(rillUserOrgs: string[]) {
    let org: string | undefined;
    let checkNextOrg = false;
    if (rillUserOrgs.length === 1) {
      org = rillUserOrgs[0];
    } else if (rillUserOrgs.length > 1) {
      this.promptOrgSelection.set(true);
    } else {
      const userResp = await queryClient.fetchQuery({
        queryKey: getLocalServiceGetCurrentUserQueryKey(),
        queryFn: localServiceGetCurrentUser,
      });
      org = this.getOrgNameFromEmail(userResp.user?.email ?? "");
      checkNextOrg = true;
    }
    return { org, checkNextOrg };
  }

  private async tryDeployWithOrg(
    org: string,
    projectName: string,
    checkNextOrg: boolean,
  ) {
    let i = 0;

    // eslint-disable-next-line no-constant-condition
    while (true) {
      try {
        const resp = await get(this.deployMutation).mutateAsync({
          projectName,
          org: `${org}${i === 0 ? "" : "-" + i}`,
          upload: true,
        });
        // wait for the telemetry to finish since the page will be redirected after a deploy success
        await behaviourEvent?.fireDeployEvent(
          BehaviourEventAction.DeploySuccess,
        );
        return resp.frontendUrl;
      } catch (e) {
        const err = extractDeployError(e);
        if (err.noAccess && checkNextOrg) {
          i++;
        } else {
          throw e;
        }
      }
    }
  }

  private getOrgNameFromEmail(email: string): string {
    return sanitizeName(email.split("@")[0] ?? "");
  }
}
