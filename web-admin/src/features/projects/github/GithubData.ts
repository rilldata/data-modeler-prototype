import {
  adminServiceListGithubUserRepos,
  createAdminServiceGetGithubUserStatus,
  createAdminServiceListGithubUserRepos,
  getAdminServiceGetGithubUserStatusQueryKey,
  getAdminServiceListGithubUserReposQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
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

  private promptingUser: boolean;

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

    await waitUntil(() => !get(this.userStatus).isFetching);
    const userStatus = get(this.userStatus).data;
    if (userStatus?.hasAccess) {
      return;
    }

    this.promptingUser = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  /**
   * Used to reselect connected repos.
   * Opens the grantAccessUrl page to achieve this.
   */
  public async reselectRepos() {
    await waitUntil(() => !get(this.userStatus).isFetching);
    const userStatus = get(this.userStatus).data;

    this.promptingUser = true;
    window.open(userStatus.grantAccessUrl, "_blank");
  }

  /**
   * If prompting user to connect to github then check the status of the connection.
   *
   * If did not have access before, refetch the user status query. (list of repos is fetched since the enabled will be flipped)
   * Else refetch the list of queries.
   */
  public async refetch() {
    if (!this.promptingUser) return;
    this.promptingUser = false;

    const userStatus = get(this.userStatus).data;
    if (!userStatus.hasAccess) {
      // refetch status if had no access
      await queryClient.refetchQueries(
        getAdminServiceGetGithubUserStatusQueryKey(),
      );

      await waitUntil(() => !get(this.userStatus).isFetching);
      if (!get(this.userStatus).data.hasAccess) {
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
