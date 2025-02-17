import type { RpcStatus } from "@rilldata/web-admin/client";
import { hasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
import { fetchAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
import { error, redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async ({ params: { organization }, parent }) => {
  const { organizationPermissions, issues } = await parent();

  if (!organizationPermissions.manageOrg) return;

  let projectHibernating = false;
  try {
    projectHibernating = await fetchAllProjectsHibernating(organization);
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Error fetching projects for the organization");
    }

    throw error(e.response.status, e.response.data.message);
  }

  // if all projects were hibernated due to a blocker issue on org then take the user to projects page
  if (hasBlockerIssues(issues) && projectHibernating) {
    throw redirect(307, `/${organization}`);
  }
};
