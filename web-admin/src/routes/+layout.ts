/* 
In svelte.config.js, the "adapter-static" option makes the application a single-page
app in production. Here, we are setting server-side rendering (SSR) to false to 
ensure the same single-page app behavior in development.
*/
export const ssr = false;

import { dev } from "$app/environment";
import {
  adminServiceGetCurrentUser,
  getAdminServiceGetCurrentUserQueryKey,
  type RpcStatus,
  type V1GetCurrentUserResponse,
  type V1OrganizationPermissions,
  type V1ProjectPermissions,
  type V1User,
} from "@rilldata/web-admin/client";
import { redirectToLoginOrRequestAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { getFetchOrganizationQueryOptions } from "@rilldata/web-admin/features/organizations/selectors";
import { fetchProjectDeploymentDetails } from "@rilldata/web-admin/features/projects/selectors";
import { initPosthog } from "@rilldata/web-common/lib/analytics/posthog";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, redirect, type Page } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async ({ params, url, route, depends }) => {
  depends("root");
  // Route params
  const { organization, project, token: routeToken } = params;
  const pageState = {
    url,
    route,
    params,
  } as Page;

  let searchParamToken: string | undefined;
  if (url.searchParams.has("token")) {
    searchParamToken = url.searchParams.get("token") ?? undefined;
  }
  const token = searchParamToken ?? routeToken;

  // Initialize analytics
  const shouldSendAnalytics = !import.meta.env.VITE_PLAYWRIGHT_TEST && !dev;
  if (shouldSendAnalytics) {
    const rillVersion = import.meta.env.RILL_UI_VERSION;
    const posthogSessionId = url.searchParams.get("ph_session_id") as
      | string
      | null;
    initPosthog(rillVersion, posthogSessionId);
    if (posthogSessionId) {
      // Remove the PostHog sessionID from the url
      url.searchParams.delete("ph_session_id");
      throw redirect(307, url.toString());
    }
  }

  let user: V1User | undefined;
  try {
    const userQuery = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
      queryKey: getAdminServiceGetCurrentUserQueryKey(),
      queryFn: () => adminServiceGetCurrentUser(),
    });
    user = userQuery.user;
  } catch {
    // no-op
  }

  // If no organization or project, return empty permissions
  if (!organization) {
    return {
      user,
      organizationPermissions: <V1OrganizationPermissions>{},
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  // Get organization permissions
  let organizationPermissions: V1OrganizationPermissions = {};
  let organizationLogoUrl: string | undefined = undefined;
  let organizationFaviconUrl: string | undefined = undefined;
  if (organization && !token) {
    try {
      const organizationResp = await queryClient.fetchQuery(
        getFetchOrganizationQueryOptions(organization),
      );
      organizationPermissions = organizationResp.permissions ?? {};
      organizationLogoUrl = organizationResp.organization?.logoUrl;
      organizationFaviconUrl = organizationResp.organization?.faviconUrl;
    } catch (e: unknown) {
      if (!isAxiosError<RpcStatus>(e) || !e.response) {
        throw error(500, "Error fetching organization");
      }

      const shouldRedirectToRequestAccess =
        e.response.status === 403 && !!project;

      if (shouldRedirectToRequestAccess) {
        // The redirect is handled below after the call to `GetProject`
      } else {
        throw error(e.response.status, e.response.data.message);
      }
    }
  }

  if (!project) {
    return {
      user,
      organizationPermissions,
      organizationLogoUrl,
      organizationFaviconUrl,
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  try {
    const {
      projectPermissions,
      project: proj,
      runtime: runtimeData,
    } = await fetchProjectDeploymentDetails(organization, project, token);

    await runtime.setRuntime(
      queryClient,
      fixLocalhostRuntimePort(runtimeData.host ?? ""),
      runtimeData.instanceId,
      runtimeData.jwt?.token,
      runtimeData.jwt?.authContext,
    );

    return {
      user,
      organizationPermissions,
      organizationLogoUrl,
      organizationFaviconUrl,
      projectPermissions,
      project: proj,
      runtime: runtimeData,
    };
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Error fetching project");
    }

    const shouldRedirectToRequestAccess =
      e.response.status === 403 && !!project;

    if (shouldRedirectToRequestAccess) {
      const didRedirect = await redirectToLoginOrRequestAccess(pageState);
      if (didRedirect) return;
    }

    throw error(e.response.status, e.response.data.message);
  }
};
