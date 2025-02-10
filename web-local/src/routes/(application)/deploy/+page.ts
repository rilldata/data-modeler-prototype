import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getLocalServiceListOrganizationsAndBillingMetadataRequestQueryKey,
  localServiceListOrganizationsAndBillingMetadataRequest,
} from "@rilldata/web-common/runtime-client/local-service";

export const load = async ({ url }) => {
  const orgMetadata = await queryClient.fetchQuery({
    queryKey:
      getLocalServiceListOrganizationsAndBillingMetadataRequestQueryKey(),
    queryFn: () => localServiceListOrganizationsAndBillingMetadataRequest(),
    staleTime: Infinity,
  });
  return {
    orgParam: url.searchParams.get("org") ?? "",
    orgMetadata,
  };
};
