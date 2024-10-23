import {
  adminServiceListGithubUserRepos,
  createAdminServiceGetGithubUserStatus,
  createAdminServiceListGithubUserRepos,
  getAdminServiceGetGithubUserStatusQueryKey,
  getAdminServiceListGithubUserReposQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { openPopupWindow } from "@rilldata/web-common/lib/openPopupWindow";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/utils";
import { getContext, setContext } from "svelte";
import { derived, get, writable } from "svelte/store";

/**
 * Contains information about user connection to github and list of repos currently connected.
 */
export class GithubData {
  public readonly repoSelectionOpen = writable(false);
  public readonly githubConnectionFailed = writable(false);

  public readonly userStatus = createAdminServiceGetGithubUserStatus({
    query: {
      queryClient,
    },
  });
  public readonly userRepos = derived(
    [this.userStatus, this.repoSelectionOpen],
    ([userStatus, repoSelectionOpen], set) =>
      createAdminServiceListGithubUserRepos({
        query: {
          // do not run it when user gets to status page, only when repo selection is open
          enabled: !!userStatus.data?.hasAccess && repoSelectionOpen,
          queryClient,
        },
      }).subscribe(set),
  ) as ReturnType<
    typeof createAdminServiceListGithubUserRepos<
      Awaited<ReturnType<typeof adminServiceListGithubUserRepos>>,
      RpcStatus
    >
  >;

  private userPromptWindow: Window | null = null;
  private windowCheckTimer: ReturnType<typeof setInterval> | null = null;

  public readonly status = derived(
    [this.userStatus, this.userRepos],
    ([userStatus, userRepos]) => {
      if (userStatus.isFetching || userRepos.isFetching) {
        return {
          isFetching: true,
          error: undefined,
        };
      }

      return {
        isFetching: false,
        error: userStatus.error ?? userRepos.error,
      };
    },
  );

  /**
   * Marks the repo selection dialog to be opened.
   * If user doesn't have access, opens grant access page.
   */
  public async startRepoSelection() {
    this.repoSelectionOpen.set(true);

    return this.ensureGithubAccess();
  }

  /**
   * Used to reselect connected repos.
   * Opens the grantAccessUrl page to achieve this.
   */
  public async reselectRepos() {
    await waitUntil(() => !get(this.userStatus).isFetching);
    const userStatus = get(this.userStatus).data;
    if (!userStatus?.grantAccessUrl) {
      return;
    }

    this.openUserGithubConnectWindow(userStatus.grantAccessUrl);
  }

  public async ensureGithubAccess() {
    await waitUntil(() => !get(this.userStatus).isFetching);
    const userStatus = get(this.userStatus).data;
    if (!userStatus || userStatus?.hasAccess || !userStatus.grantAccessUrl) {
      return;
    }

    this.openUserGithubConnectWindow(userStatus.grantAccessUrl);
  }

  private openUserGithubConnectWindow(url: string) {
    try {
      // safeguard try catch
      this.userPromptWindow?.close();
    } catch {
      // no-op
    }
    if (this.windowCheckTimer) clearInterval(this.windowCheckTimer);
    this.userPromptWindow = openPopupWindow(
      // add `remote` to indicate the callback success dialog should auto close
      `${url}?remote=autoclose`,
      "githubWindow",
    );

    // periodically check if the new window was closed
    this.windowCheckTimer = setInterval(() => {
      if (!this.userPromptWindow?.closed) return;
      if (this.windowCheckTimer) clearInterval(this.windowCheckTimer);
      this.windowCheckTimer = null;
      this.userPromptWindow = null;
      void this.refetch();
    }, 200);
  }

  private async refetch() {
    const userStatus = get(this.userStatus).data;
    if (!userStatus?.hasAccess) {
      // refetch status if had no access
      await queryClient.refetchQueries(
        getAdminServiceGetGithubUserStatusQueryKey(),
      );

      await waitUntil(() => !get(this.userStatus).isFetching);
      if (!get(this.userStatus).data?.hasAccess) {
        this.githubConnectionFailed.set(true);
      } else {
        this.githubConnectionFailed.set(false);
      }
    } else {
      this.githubConnectionFailed.set(false);

      // else refetch the list of repos
      await queryClient.refetchQueries(
        getAdminServiceListGithubUserReposQueryKey(),
      );
    }
  }
}

export function setGithubData(githubData: GithubData) {
  setContext("rill:app:github", githubData);
}

export function getGithubData() {
  return getContext<GithubData>("rill:app:github");
}
