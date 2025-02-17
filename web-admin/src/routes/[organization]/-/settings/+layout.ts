import {
  adminServiceGetBillingSubscription,
  getAdminServiceGetBillingSubscriptionQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { getNeverSubscribedIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent, params }) => {
  const { issues } = await parent();
  const neverSubscribed = !!getNeverSubscribedIssue(issues);

  const queryKey = getAdminServiceGetBillingSubscriptionQueryKey(
    params.organization,
  );
  const queryFn = () => adminServiceGetBillingSubscription(params.organization);

  try {
    const billingSubscription = await queryClient.fetchQuery({
      queryKey,
      queryFn,
    });
    return {
      subscription: billingSubscription.subscription,
      billingPortalUrl: billingSubscription.billingPortalUrl,
      neverSubscribed,
    };
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Failed to fetch billing subscription");
    }

    throw error(e.response.status, e.response.data.message);
  }
};
