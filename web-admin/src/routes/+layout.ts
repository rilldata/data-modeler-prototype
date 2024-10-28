/* 
In svelte.config.js, the "adapter-static" option makes the application a single-page
app in production. Here, we are setting server-side rendering (SSR) to false to 
ensure the same single-page app behavior in development.
*/
export const ssr = false;

import {
  adminServiceGetProject,
  getAdminServiceGetProjectQueryKey,
  type V1OrganizationPermissions,
  type V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import { redirectToLoginIfNotLoggedIn } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import {
  isProjectRequestAccessPage,
  withinProject,
} from "@rilldata/web-admin/features/navigation/nav-utils";
import { fetchOrganizationPermissions } from "@rilldata/web-admin/features/organizations/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error, type Page, redirect } from "@sveltejs/kit";
import type { QueryFunction, QueryKey } from "@tanstack/svelte-query";
import {
  adminServiceGetProjectWithBearerToken,
  getAdminServiceGetProjectWithBearerTokenQueryKey,
} from "../features/public-urls/get-project-with-bearer-token.js";

export const load = async ({ params, url, route }) => {
  const { organization, project, token: routeToken } = params;

  let searchParamToken: string | undefined;
  if (url.searchParams.has("token")) {
    searchParamToken = url.searchParams.get("token");
  }
  const token = searchParamToken ?? routeToken;

  let organizationPermissions: V1OrganizationPermissions = {};
  if (organization && !token) {
    try {
      organizationPermissions =
        await fetchOrganizationPermissions(organization);
    } catch (e) {
      if (
        e.response?.status !== 403 ||
        (await redirectToLoginIfNotLoggedIn())
      ) {
        if (
          withinProject({ url, route } as Page) &&
          !isProjectRequestAccessPage({ url, route } as Page)
        ) {
          // if not in request access page (approve or deny routes) then go to a page to get access
          throw redirect(
            307,
            `/-/request-project-access/?organization=${organization}&project=${project}`,
          );
        }
        throw error(e.response.status, "Error fetching organization");
      }
    }
  }

  if (!organization || !project) {
    return {
      organizationPermissions,
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  let queryKey: QueryKey;
  let queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProject>>
  >;

  if (token) {
    queryKey = getAdminServiceGetProjectWithBearerTokenQueryKey(
      organization,
      project,
      token,
      {},
    );

    queryFn = ({ signal }) =>
      adminServiceGetProjectWithBearerToken(
        organization,
        project,
        token,
        {},
        signal,
      );
  } else {
    queryKey = getAdminServiceGetProjectQueryKey(organization, project);

    queryFn = ({ signal }) =>
      adminServiceGetProject(organization, project, {}, signal);
  }

  try {
    const response = await queryClient.fetchQuery({
      queryFn,
      queryKey,
    });

    const { projectPermissions } = response;

    return {
      organizationPermissions,
      projectPermissions,
    };
  } catch (e) {
    if (e.response?.status !== 403 || (await redirectToLoginIfNotLoggedIn())) {
      if (
        withinProject({ url, route } as Page) &&
        !isProjectRequestAccessPage({ url, route } as Page)
      ) {
        // if not in request access page (approve or deny routes) then go to a page to get access
        throw redirect(
          307,
          `/-/request-project-access/?organization=${organization}&project=${project}`,
        );
      }
      throw error(e.response.status, "Error fetching deployment");
    }
  }
};
