/**
 * Generated by orval v6.10.1 🍺
 * Do not edit manually.
 * rill/admin/v1/api.proto
 * OpenAPI spec version: version not set
 */
export type AdminServiceUpdateProjectBody = {
  description?: string;
  githubUrl?: string;
  productionBranch?: string;
};

export type AdminServiceCreateProjectBody = {
  description?: string;
  githubUrl?: string;
  name?: string;
  productionBranch?: string;
};

export type AdminServiceListProjectsParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceUpdateOrganizationBody = {
  description?: string;
};

export type AdminServiceListOrganizationsParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceGetGithubRepoStatusParams = { githubUrl?: string };

export interface V1User {
  createdOn?: string;
  displayName?: string;
  email?: string;
  id?: string;
  photoUrl?: string;
  updatedOn?: string;
}

export interface V1UpdateOrganizationResponse {
  organization?: V1Organization;
}

export interface V1RevokeCurrentAuthTokenResponse {
  tokenId?: string;
}

export interface V1Project {
  createdOn?: string;
  description?: string;
  githubUrl?: string;
  id?: string;
  name?: string;
  productionBranch?: string;
  updatedOn?: string;
}

export interface V1UpdateProjectResponse {
  project?: V1Project;
}

export interface V1PingResponse {
  time?: string;
  version?: string;
}

export interface V1Organization {
  createdOn?: string;
  description?: string;
  id?: string;
  name?: string;
  updatedOn?: string;
}

export interface V1ListProjectsResponse {
  nextPageToken?: string;
  projects?: V1Project[];
}

export interface V1ListOrganizationsResponse {
  nextPageToken?: string;
  organization?: V1Organization[];
}

export interface V1GetProjectResponse {
  project?: V1Project;
}

export interface V1GetOrganizationResponse {
  organization?: V1Organization;
}

export interface V1GetGithubRepoStatusResponse {
  defaultBranch?: string;
  grantAccessUrl?: string;
  hasAccess?: boolean;
}

export interface V1GetCurrentUserResponse {
  user?: V1User;
}

export interface V1DeleteProjectResponse {
  name?: string;
}

export interface V1DeleteOrganizationResponse {
  name?: string;
}

export interface V1CreateProjectResponse {
  project?: V1Project;
}

export interface V1CreateOrganizationResponse {
  organization?: V1Organization;
}

export interface V1CreateOrganizationRequest {
  description?: string;
  id?: string;
  name?: string;
}

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: unknown;
}

export interface RpcStatus {
  code?: number;
  details?: ProtobufAny[];
  message?: string;
}
