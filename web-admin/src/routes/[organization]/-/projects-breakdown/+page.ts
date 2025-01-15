import {
  adminServiceGetBillingProjectCredentials,
  getAdminServiceGetBillingProjectCredentialsQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error } from "@sveltejs/kit";

export const load = async ({ params, parent }) => {
  const { organizationPermissions } = await parent();
  if (!organizationPermissions.manageProjects) {
    throw error(404, "Page not found");
  }

  try {
    const { organization } = params;
    const billingProjectCredsResp = await queryClient.fetchQuery({
      queryKey: getAdminServiceGetBillingProjectCredentialsQueryKey({
        organization,
      }),
      queryFn: () => adminServiceGetBillingProjectCredentials({ organization }),
    });
    const runtime: Runtime = {
      host: fixLocalhostRuntimePort(billingProjectCredsResp.runtimeHost ?? ""),
      instanceId: billingProjectCredsResp.instanceId ?? "",
      jwt: {
        token: billingProjectCredsResp.accessToken ?? "",
        authContext: "embed",
        receivedAt: Date.now(),
      },
    };

    return {
      runtime,
    };
  } catch (err) {
    const statusCode = err?.response?.status || 500;
    throw error(statusCode, "Project breakdown is not available right now");
  }
};
