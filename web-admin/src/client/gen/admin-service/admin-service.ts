/**
 * Generated by orval v6.10.1 🍺
 * Do not edit manually.
 * rill/admin/v1/api.proto
 * OpenAPI spec version: version not set
 */
import { useQuery, useMutation } from "@sveltestack/svelte-query";
import type {
  UseQueryOptions,
  UseMutationOptions,
  QueryFunction,
  MutationFunction,
  UseQueryStoreResult,
  QueryKey,
} from "@sveltestack/svelte-query";
import type {
  V1GetGithubRepoStatusResponse,
  RpcStatus,
  AdminServiceGetGithubRepoStatusParams,
  V1ListOrganizationsResponse,
  AdminServiceListOrganizationsParams,
  V1CreateOrganizationResponse,
  V1CreateOrganizationRequest,
  V1GetOrganizationResponse,
  V1DeleteOrganizationResponse,
  V1UpdateOrganizationResponse,
  AdminServiceUpdateOrganizationBody,
  V1ListProjectsResponse,
  AdminServiceListProjectsParams,
  V1CreateProjectResponse,
  AdminServiceCreateProjectBody,
  V1GetProjectResponse,
  V1DeleteProjectResponse,
  V1UpdateProjectResponse,
  AdminServiceUpdateProjectBody,
  V1PingResponse,
  V1RevokeCurrentAuthTokenResponse,
  V1GetCurrentUserResponse,
} from "../index.schemas";
import { httpClient } from "../../http-client";

/**
 * @summary GetGithubRepoRequest returns info about a Github repo based on the caller's installations.
If the caller has not granted access to the repository, instructions for granting access are returned.
 */
export const adminServiceGetGithubRepoStatus = (
  params?: AdminServiceGetGithubRepoStatusParams,
  signal?: AbortSignal
) => {
  return httpClient<V1GetGithubRepoStatusResponse>({
    url: `/v1/github/repositories`,
    method: "get",
    params,
    signal,
  });
};

export const getAdminServiceGetGithubRepoStatusQueryKey = (
  params?: AdminServiceGetGithubRepoStatusParams
) => [`/v1/github/repositories`, ...(params ? [params] : [])];

export type AdminServiceGetGithubRepoStatusQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>
>;
export type AdminServiceGetGithubRepoStatusQueryError = RpcStatus;

export const useAdminServiceGetGithubRepoStatus = <
  TData = Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>,
  TError = RpcStatus
>(
  params?: AdminServiceGetGithubRepoStatusParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getAdminServiceGetGithubRepoStatusQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>
  > = ({ signal }) => adminServiceGetGithubRepoStatus(params, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>,
    TError,
    TData
  >(queryKey, queryFn, queryOptions) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServiceGetGithubRepoStatus>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary ListOrganizations lists all the organizations currently managed by the admin
 */
export const adminServiceListOrganizations = (
  params?: AdminServiceListOrganizationsParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ListOrganizationsResponse>({
    url: `/v1/organizations`,
    method: "get",
    params,
    signal,
  });
};

export const getAdminServiceListOrganizationsQueryKey = (
  params?: AdminServiceListOrganizationsParams
) => [`/v1/organizations`, ...(params ? [params] : [])];

export type AdminServiceListOrganizationsQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceListOrganizations>>
>;
export type AdminServiceListOrganizationsQueryError = RpcStatus;

export const useAdminServiceListOrganizations = <
  TData = Awaited<ReturnType<typeof adminServiceListOrganizations>>,
  TError = RpcStatus
>(
  params?: AdminServiceListOrganizationsParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof adminServiceListOrganizations>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServiceListOrganizations>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getAdminServiceListOrganizationsQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceListOrganizations>>
  > = ({ signal }) => adminServiceListOrganizations(params, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServiceListOrganizations>>,
    TError,
    TData
  >(queryKey, queryFn, queryOptions) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServiceListOrganizations>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary CreateOrganization creates a new organization
 */
export const adminServiceCreateOrganization = (
  v1CreateOrganizationRequest: V1CreateOrganizationRequest
) => {
  return httpClient<V1CreateOrganizationResponse>({
    url: `/v1/organizations`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: v1CreateOrganizationRequest,
  });
};

export type AdminServiceCreateOrganizationMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceCreateOrganization>>
>;
export type AdminServiceCreateOrganizationMutationBody =
  V1CreateOrganizationRequest;
export type AdminServiceCreateOrganizationMutationError = RpcStatus;

export const useAdminServiceCreateOrganization = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceCreateOrganization>>,
    TError,
    { data: V1CreateOrganizationRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceCreateOrganization>>,
    { data: V1CreateOrganizationRequest }
  > = (props) => {
    const { data } = props ?? {};

    return adminServiceCreateOrganization(data);
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceCreateOrganization>>,
    TError,
    { data: V1CreateOrganizationRequest },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary GetOrganization returns information about a specific organization
 */
export const adminServiceGetOrganization = (
  name: string,
  signal?: AbortSignal
) => {
  return httpClient<V1GetOrganizationResponse>({
    url: `/v1/organizations/${name}`,
    method: "get",
    signal,
  });
};

export const getAdminServiceGetOrganizationQueryKey = (name: string) => [
  `/v1/organizations/${name}`,
];

export type AdminServiceGetOrganizationQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceGetOrganization>>
>;
export type AdminServiceGetOrganizationQueryError = RpcStatus;

export const useAdminServiceGetOrganization = <
  TData = Awaited<ReturnType<typeof adminServiceGetOrganization>>,
  TError = RpcStatus
>(
  name: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof adminServiceGetOrganization>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServiceGetOrganization>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getAdminServiceGetOrganizationQueryKey(name);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetOrganization>>
  > = ({ signal }) => adminServiceGetOrganization(name, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServiceGetOrganization>>,
    TError,
    TData
  >(queryKey, queryFn, {
    enabled: !!name,
    ...queryOptions,
  }) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServiceGetOrganization>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary DeleteOrganization deletes an organizations
 */
export const adminServiceDeleteOrganization = (name: string) => {
  return httpClient<V1DeleteOrganizationResponse>({
    url: `/v1/organizations/${name}`,
    method: "delete",
  });
};

export type AdminServiceDeleteOrganizationMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceDeleteOrganization>>
>;

export type AdminServiceDeleteOrganizationMutationError = RpcStatus;

export const useAdminServiceDeleteOrganization = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceDeleteOrganization>>,
    TError,
    { name: string },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceDeleteOrganization>>,
    { name: string }
  > = (props) => {
    const { name } = props ?? {};

    return adminServiceDeleteOrganization(name);
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceDeleteOrganization>>,
    TError,
    { name: string },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary UpdateOrganization deletes an organizations
 */
export const adminServiceUpdateOrganization = (
  name: string,
  adminServiceUpdateOrganizationBody: AdminServiceUpdateOrganizationBody
) => {
  return httpClient<V1UpdateOrganizationResponse>({
    url: `/v1/organizations/${name}`,
    method: "put",
    headers: { "Content-Type": "application/json" },
    data: adminServiceUpdateOrganizationBody,
  });
};

export type AdminServiceUpdateOrganizationMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceUpdateOrganization>>
>;
export type AdminServiceUpdateOrganizationMutationBody =
  AdminServiceUpdateOrganizationBody;
export type AdminServiceUpdateOrganizationMutationError = RpcStatus;

export const useAdminServiceUpdateOrganization = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceUpdateOrganization>>,
    TError,
    { name: string; data: AdminServiceUpdateOrganizationBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceUpdateOrganization>>,
    { name: string; data: AdminServiceUpdateOrganizationBody }
  > = (props) => {
    const { name, data } = props ?? {};

    return adminServiceUpdateOrganization(name, data);
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceUpdateOrganization>>,
    TError,
    { name: string; data: AdminServiceUpdateOrganizationBody },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary ListProjects lists all the projects currently available for given organizations
 */
export const adminServiceListProjects = (
  organizationName: string,
  params?: AdminServiceListProjectsParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ListProjectsResponse>({
    url: `/v1/organizations/${organizationName}/projects`,
    method: "get",
    params,
    signal,
  });
};

export const getAdminServiceListProjectsQueryKey = (
  organizationName: string,
  params?: AdminServiceListProjectsParams
) => [
  `/v1/organizations/${organizationName}/projects`,
  ...(params ? [params] : []),
];

export type AdminServiceListProjectsQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceListProjects>>
>;
export type AdminServiceListProjectsQueryError = RpcStatus;

export const useAdminServiceListProjects = <
  TData = Awaited<ReturnType<typeof adminServiceListProjects>>,
  TError = RpcStatus
>(
  organizationName: string,
  params?: AdminServiceListProjectsParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof adminServiceListProjects>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServiceListProjects>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getAdminServiceListProjectsQueryKey(organizationName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceListProjects>>
  > = ({ signal }) =>
    adminServiceListProjects(organizationName, params, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServiceListProjects>>,
    TError,
    TData
  >(queryKey, queryFn, {
    enabled: !!organizationName,
    ...queryOptions,
  }) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServiceListProjects>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary CreateProject creates a new project
 */
export const adminServiceCreateProject = (
  organizationName: string,
  adminServiceCreateProjectBody: AdminServiceCreateProjectBody
) => {
  return httpClient<V1CreateProjectResponse>({
    url: `/v1/organizations/${organizationName}/projects`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: adminServiceCreateProjectBody,
  });
};

export type AdminServiceCreateProjectMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceCreateProject>>
>;
export type AdminServiceCreateProjectMutationBody =
  AdminServiceCreateProjectBody;
export type AdminServiceCreateProjectMutationError = RpcStatus;

export const useAdminServiceCreateProject = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceCreateProject>>,
    TError,
    { organizationName: string; data: AdminServiceCreateProjectBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceCreateProject>>,
    { organizationName: string; data: AdminServiceCreateProjectBody }
  > = (props) => {
    const { organizationName, data } = props ?? {};

    return adminServiceCreateProject(organizationName, data);
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceCreateProject>>,
    TError,
    { organizationName: string; data: AdminServiceCreateProjectBody },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary GetProject returns information about a specific project
 */
export const adminServiceGetProject = (
  organizationName: string,
  name: string,
  signal?: AbortSignal
) => {
  return httpClient<V1GetProjectResponse>({
    url: `/v1/organizations/${organizationName}/projects/${name}`,
    method: "get",
    signal,
  });
};

export const getAdminServiceGetProjectQueryKey = (
  organizationName: string,
  name: string
) => [`/v1/organizations/${organizationName}/projects/${name}`];

export type AdminServiceGetProjectQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceGetProject>>
>;
export type AdminServiceGetProjectQueryError = RpcStatus;

export const useAdminServiceGetProject = <
  TData = Awaited<ReturnType<typeof adminServiceGetProject>>,
  TError = RpcStatus
>(
  organizationName: string,
  name: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof adminServiceGetProject>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServiceGetProject>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getAdminServiceGetProjectQueryKey(organizationName, name);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProject>>
  > = ({ signal }) => adminServiceGetProject(organizationName, name, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServiceGetProject>>,
    TError,
    TData
  >(queryKey, queryFn, {
    enabled: !!(organizationName && name),
    ...queryOptions,
  }) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServiceGetProject>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary DeleteProject deletes an project
 */
export const adminServiceDeleteProject = (
  organizationName: string,
  name: string
) => {
  return httpClient<V1DeleteProjectResponse>({
    url: `/v1/organizations/${organizationName}/projects/${name}`,
    method: "delete",
  });
};

export type AdminServiceDeleteProjectMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceDeleteProject>>
>;

export type AdminServiceDeleteProjectMutationError = RpcStatus;

export const useAdminServiceDeleteProject = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceDeleteProject>>,
    TError,
    { organizationName: string; name: string },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceDeleteProject>>,
    { organizationName: string; name: string }
  > = (props) => {
    const { organizationName, name } = props ?? {};

    return adminServiceDeleteProject(organizationName, name);
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceDeleteProject>>,
    TError,
    { organizationName: string; name: string },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary UpdateProject updates a project
 */
export const adminServiceUpdateProject = (
  organizationName: string,
  name: string,
  adminServiceUpdateProjectBody: AdminServiceUpdateProjectBody
) => {
  return httpClient<V1UpdateProjectResponse>({
    url: `/v1/organizations/${organizationName}/projects/${name}`,
    method: "put",
    headers: { "Content-Type": "application/json" },
    data: adminServiceUpdateProjectBody,
  });
};

export type AdminServiceUpdateProjectMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceUpdateProject>>
>;
export type AdminServiceUpdateProjectMutationBody =
  AdminServiceUpdateProjectBody;
export type AdminServiceUpdateProjectMutationError = RpcStatus;

export const useAdminServiceUpdateProject = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceUpdateProject>>,
    TError,
    {
      organizationName: string;
      name: string;
      data: AdminServiceUpdateProjectBody;
    },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceUpdateProject>>,
    {
      organizationName: string;
      name: string;
      data: AdminServiceUpdateProjectBody;
    }
  > = (props) => {
    const { organizationName, name, data } = props ?? {};

    return adminServiceUpdateProject(organizationName, name, data);
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceUpdateProject>>,
    TError,
    {
      organizationName: string;
      name: string;
      data: AdminServiceUpdateProjectBody;
    },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary Ping returns information about the server
 */
export const adminServicePing = (signal?: AbortSignal) => {
  return httpClient<V1PingResponse>({ url: `/v1/ping`, method: "get", signal });
};

export const getAdminServicePingQueryKey = () => [`/v1/ping`];

export type AdminServicePingQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServicePing>>
>;
export type AdminServicePingQueryError = RpcStatus;

export const useAdminServicePing = <
  TData = Awaited<ReturnType<typeof adminServicePing>>,
  TError = RpcStatus
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof adminServicePing>>,
    TError,
    TData
  >;
}): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServicePing>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getAdminServicePingQueryKey();

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServicePing>>
  > = ({ signal }) => adminServicePing(signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServicePing>>,
    TError,
    TData
  >(queryKey, queryFn, queryOptions) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServicePing>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary RevokeCurrentAuthToken revoke the current auth token
 */
export const adminServiceRevokeCurrentAuthToken = () => {
  return httpClient<V1RevokeCurrentAuthTokenResponse>({
    url: `/v1/tokens/current`,
    method: "delete",
  });
};

export type AdminServiceRevokeCurrentAuthTokenMutationResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceRevokeCurrentAuthToken>>
>;

export type AdminServiceRevokeCurrentAuthTokenMutationError = RpcStatus;

export const useAdminServiceRevokeCurrentAuthToken = <
  TError = RpcStatus,
  TVariables = void,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof adminServiceRevokeCurrentAuthToken>>,
    TError,
    TVariables,
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof adminServiceRevokeCurrentAuthToken>>,
    TVariables
  > = () => {
    return adminServiceRevokeCurrentAuthToken();
  };

  return useMutation<
    Awaited<ReturnType<typeof adminServiceRevokeCurrentAuthToken>>,
    TError,
    TVariables,
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary GetCurrentUser returns the currently authenticated user (if any)
 */
export const adminServiceGetCurrentUser = (signal?: AbortSignal) => {
  return httpClient<V1GetCurrentUserResponse>({
    url: `/v1/users/current`,
    method: "get",
    signal,
  });
};

export const getAdminServiceGetCurrentUserQueryKey = () => [
  `/v1/users/current`,
];

export type AdminServiceGetCurrentUserQueryResult = NonNullable<
  Awaited<ReturnType<typeof adminServiceGetCurrentUser>>
>;
export type AdminServiceGetCurrentUserQueryError = RpcStatus;

export const useAdminServiceGetCurrentUser = <
  TData = Awaited<ReturnType<typeof adminServiceGetCurrentUser>>,
  TError = RpcStatus
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof adminServiceGetCurrentUser>>,
    TError,
    TData
  >;
}): UseQueryStoreResult<
  Awaited<ReturnType<typeof adminServiceGetCurrentUser>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getAdminServiceGetCurrentUserQueryKey();

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetCurrentUser>>
  > = ({ signal }) => adminServiceGetCurrentUser(signal);

  const query = useQuery<
    Awaited<ReturnType<typeof adminServiceGetCurrentUser>>,
    TError,
    TData
  >(queryKey, queryFn, queryOptions) as UseQueryStoreResult<
    Awaited<ReturnType<typeof adminServiceGetCurrentUser>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
