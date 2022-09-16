/**
 * Generated by orval v6.10.0 🍺
 * Do not edit manually.
 * Rill Cloud
 * OpenAPI spec version: 1.0.0
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
  Organization,
  ErrorResponseResponse,
  CreateOrganizationBody,
  UpdateOrganizationBody,
  Project,
  CreateProjectBody,
  UpdateProjectBody,
} from "../index.schemas";
import { httpClient } from "../../http-client";

export const findOrganizations = (signal?: AbortSignal) => {
  return httpClient<Organization[]>({
    url: `/v1/organizations`,
    method: "get",
    signal,
  });
};

export const getFindOrganizationsQueryKey = () => [`/v1/organizations`];

export type FindOrganizationsQueryResult = NonNullable<
  Awaited<ReturnType<typeof findOrganizations>>
>;
export type FindOrganizationsQueryError = ErrorResponseResponse;

export const useFindOrganizations = <
  TData = Awaited<ReturnType<typeof findOrganizations>>,
  TError = ErrorResponseResponse
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof findOrganizations>>,
    TError,
    TData
  >;
}): UseQueryStoreResult<
  Awaited<ReturnType<typeof findOrganizations>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getFindOrganizationsQueryKey();

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof findOrganizations>>
  > = ({ signal }) => findOrganizations(signal);

  const query = useQuery<
    Awaited<ReturnType<typeof findOrganizations>>,
    TError,
    TData
  >(queryKey, queryFn, queryOptions) as UseQueryStoreResult<
    Awaited<ReturnType<typeof findOrganizations>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

export const createOrganization = (
  createOrganizationBody: CreateOrganizationBody
) => {
  return httpClient<Organization>({
    url: `/v1/organizations`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: createOrganizationBody,
  });
};

export type CreateOrganizationMutationResult = NonNullable<
  Awaited<ReturnType<typeof createOrganization>>
>;
export type CreateOrganizationMutationBody = CreateOrganizationBody;
export type CreateOrganizationMutationError = ErrorResponseResponse;

export const useCreateOrganization = <
  TError = ErrorResponseResponse,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof createOrganization>>,
    TError,
    { data: CreateOrganizationBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof createOrganization>>,
    { data: CreateOrganizationBody }
  > = (props) => {
    const { data } = props ?? {};

    return createOrganization(data);
  };

  return useMutation<
    Awaited<ReturnType<typeof createOrganization>>,
    TError,
    { data: CreateOrganizationBody },
    TContext
  >(mutationFn, mutationOptions);
};
export const findOrganization = (name: string, signal?: AbortSignal) => {
  return httpClient<Organization>({
    url: `/v1/organizations/${name}`,
    method: "get",
    signal,
  });
};

export const getFindOrganizationQueryKey = (name: string) => [
  `/v1/organizations/${name}`,
];

export type FindOrganizationQueryResult = NonNullable<
  Awaited<ReturnType<typeof findOrganization>>
>;
export type FindOrganizationQueryError = ErrorResponseResponse;

export const useFindOrganization = <
  TData = Awaited<ReturnType<typeof findOrganization>>,
  TError = ErrorResponseResponse
>(
  name: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof findOrganization>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof findOrganization>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getFindOrganizationQueryKey(name);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof findOrganization>>
  > = ({ signal }) => findOrganization(name, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof findOrganization>>,
    TError,
    TData
  >(queryKey, queryFn, {
    enabled: !!name,
    ...queryOptions,
  }) as UseQueryStoreResult<
    Awaited<ReturnType<typeof findOrganization>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

export const updateOrganization = (
  name: string,
  updateOrganizationBody: UpdateOrganizationBody
) => {
  return httpClient<Organization>({
    url: `/v1/organizations/${name}`,
    method: "put",
    headers: { "Content-Type": "application/json" },
    data: updateOrganizationBody,
  });
};

export type UpdateOrganizationMutationResult = NonNullable<
  Awaited<ReturnType<typeof updateOrganization>>
>;
export type UpdateOrganizationMutationBody = UpdateOrganizationBody;
export type UpdateOrganizationMutationError = ErrorResponseResponse;

export const useUpdateOrganization = <
  TError = ErrorResponseResponse,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof updateOrganization>>,
    TError,
    { name: string; data: UpdateOrganizationBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof updateOrganization>>,
    { name: string; data: UpdateOrganizationBody }
  > = (props) => {
    const { name, data } = props ?? {};

    return updateOrganization(name, data);
  };

  return useMutation<
    Awaited<ReturnType<typeof updateOrganization>>,
    TError,
    { name: string; data: UpdateOrganizationBody },
    TContext
  >(mutationFn, mutationOptions);
};
export const deleteOrganization = (name: string) => {
  return httpClient<void>({
    url: `/v1/organizations/${name}`,
    method: "delete",
  });
};

export type DeleteOrganizationMutationResult = NonNullable<
  Awaited<ReturnType<typeof deleteOrganization>>
>;

export type DeleteOrganizationMutationError = ErrorResponseResponse;

export const useDeleteOrganization = <
  TError = ErrorResponseResponse,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof deleteOrganization>>,
    TError,
    { name: string },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof deleteOrganization>>,
    { name: string }
  > = (props) => {
    const { name } = props ?? {};

    return deleteOrganization(name);
  };

  return useMutation<
    Awaited<ReturnType<typeof deleteOrganization>>,
    TError,
    { name: string },
    TContext
  >(mutationFn, mutationOptions);
};
export const findProjects = (organization: string, signal?: AbortSignal) => {
  return httpClient<Project[]>({
    url: `/v1/organizations/${organization}/projects`,
    method: "get",
    signal,
  });
};

export const getFindProjectsQueryKey = (organization: string) => [
  `/v1/organizations/${organization}/projects`,
];

export type FindProjectsQueryResult = NonNullable<
  Awaited<ReturnType<typeof findProjects>>
>;
export type FindProjectsQueryError = ErrorResponseResponse;

export const useFindProjects = <
  TData = Awaited<ReturnType<typeof findProjects>>,
  TError = ErrorResponseResponse
>(
  organization: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof findProjects>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof findProjects>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getFindProjectsQueryKey(organization);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof findProjects>>> = ({
    signal,
  }) => findProjects(organization, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof findProjects>>,
    TError,
    TData
  >(queryKey, queryFn, {
    enabled: !!organization,
    ...queryOptions,
  }) as UseQueryStoreResult<
    Awaited<ReturnType<typeof findProjects>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

export const createProject = (
  organization: string,
  createProjectBody: CreateProjectBody
) => {
  return httpClient<Project>({
    url: `/v1/organizations/${organization}/projects`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: createProjectBody,
  });
};

export type CreateProjectMutationResult = NonNullable<
  Awaited<ReturnType<typeof createProject>>
>;
export type CreateProjectMutationBody = CreateProjectBody;
export type CreateProjectMutationError = ErrorResponseResponse;

export const useCreateProject = <
  TError = ErrorResponseResponse,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof createProject>>,
    TError,
    { organization: string; data: CreateProjectBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof createProject>>,
    { organization: string; data: CreateProjectBody }
  > = (props) => {
    const { organization, data } = props ?? {};

    return createProject(organization, data);
  };

  return useMutation<
    Awaited<ReturnType<typeof createProject>>,
    TError,
    { organization: string; data: CreateProjectBody },
    TContext
  >(mutationFn, mutationOptions);
};
export const findProject = (
  organization: string,
  name: string,
  signal?: AbortSignal
) => {
  return httpClient<Project>({
    url: `/v1/organizations/${organization}/project/${name}`,
    method: "get",
    signal,
  });
};

export const getFindProjectQueryKey = (organization: string, name: string) => [
  `/v1/organizations/${organization}/project/${name}`,
];

export type FindProjectQueryResult = NonNullable<
  Awaited<ReturnType<typeof findProject>>
>;
export type FindProjectQueryError = ErrorResponseResponse;

export const useFindProject = <
  TData = Awaited<ReturnType<typeof findProject>>,
  TError = ErrorResponseResponse
>(
  organization: string,
  name: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof findProject>>,
      TError,
      TData
    >;
  }
): UseQueryStoreResult<
  Awaited<ReturnType<typeof findProject>>,
  TError,
  TData,
  QueryKey
> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getFindProjectQueryKey(organization, name);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof findProject>>> = ({
    signal,
  }) => findProject(organization, name, signal);

  const query = useQuery<
    Awaited<ReturnType<typeof findProject>>,
    TError,
    TData
  >(queryKey, queryFn, {
    enabled: !!(organization && name),
    ...queryOptions,
  }) as UseQueryStoreResult<
    Awaited<ReturnType<typeof findProject>>,
    TError,
    TData,
    QueryKey
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

export const updateProject = (
  organization: string,
  name: string,
  updateProjectBody: UpdateProjectBody
) => {
  return httpClient<Project>({
    url: `/v1/organizations/${organization}/project/${name}`,
    method: "put",
    headers: { "Content-Type": "application/json" },
    data: updateProjectBody,
  });
};

export type UpdateProjectMutationResult = NonNullable<
  Awaited<ReturnType<typeof updateProject>>
>;
export type UpdateProjectMutationBody = UpdateProjectBody;
export type UpdateProjectMutationError = ErrorResponseResponse;

export const useUpdateProject = <
  TError = ErrorResponseResponse,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof updateProject>>,
    TError,
    { organization: string; name: string; data: UpdateProjectBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof updateProject>>,
    { organization: string; name: string; data: UpdateProjectBody }
  > = (props) => {
    const { organization, name, data } = props ?? {};

    return updateProject(organization, name, data);
  };

  return useMutation<
    Awaited<ReturnType<typeof updateProject>>,
    TError,
    { organization: string; name: string; data: UpdateProjectBody },
    TContext
  >(mutationFn, mutationOptions);
};
export const deleteProject = (organization: string, name: string) => {
  return httpClient<void>({
    url: `/v1/organizations/${organization}/project/${name}`,
    method: "delete",
  });
};

export type DeleteProjectMutationResult = NonNullable<
  Awaited<ReturnType<typeof deleteProject>>
>;

export type DeleteProjectMutationError = ErrorResponseResponse;

export const useDeleteProject = <
  TError = ErrorResponseResponse,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof deleteProject>>,
    TError,
    { organization: string; name: string },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof deleteProject>>,
    { organization: string; name: string }
  > = (props) => {
    const { organization, name } = props ?? {};

    return deleteProject(organization, name);
  };

  return useMutation<
    Awaited<ReturnType<typeof deleteProject>>,
    TError,
    { organization: string; name: string },
    TContext
  >(mutationFn, mutationOptions);
};
