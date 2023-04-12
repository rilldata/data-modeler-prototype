/**
 * Generated by orval v6.13.1 🍺
 * Do not edit manually.
 * rill/admin/v1/api.proto
 * OpenAPI spec version: version not set
 */
export type AdminServiceUpdateProjectBodyVariables = { [key: string]: string };

export type AdminServiceUpdateProjectBody = {
  description?: string;
  githubUrl?: string;
  productionBranch?: string;
  public?: boolean;
  variables?: AdminServiceUpdateProjectBodyVariables;
};

export type AdminServiceCreateProjectBodyVariables = { [key: string]: string };

export type AdminServiceCreateProjectBody = {
  description?: string;
  githubUrl?: string;
  name?: string;
  productionBranch?: string;
  productionOlapDriver?: string;
  productionOlapDsn?: string;
  productionSlots?: string;
  public?: boolean;
  region?: string;
  variables?: AdminServiceCreateProjectBodyVariables;
};

export type AdminServiceListProjectsForOrganizationParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceSetProjectMemberRoleBody = {
  role?: string;
};

export type AdminServiceAddProjectMemberBody = {
  email?: string;
  role?: string;
};

export type AdminServiceListProjectMembersParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceSetOrganizationMemberRoleBody = {
  role?: string;
};

export type AdminServiceAddOrganizationMemberBody = {
  email?: string;
  role?: string;
};

export type AdminServiceListOrganizationMembersParams = {
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

export type AdminServiceGetGithubRepoStatusParams = {
  githubUrl?: string;
};

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

export interface V1SetProjectMemberRoleResponse {
  [key: string]: any;
}

export interface V1SetOrganizationMemberRoleResponse {
  [key: string]: any;
}

export interface V1RevokeCurrentAuthTokenResponse {
  tokenId?: string;
}

export interface V1RemoveProjectMemberResponse {
  [key: string]: any;
}

export interface V1RemoveOrganizationMemberResponse {
  [key: string]: any;
}

export type V1ProjectVariables = { [key: string]: string };

export interface V1Project {
  createdOn?: string;
  description?: string;
  githubUrl?: string;
  id?: string;
  name?: string;
  orgId?: string;
  orgName?: string;
  productionBranch?: string;
  productionDeploymentId?: string;
  productionOlapDriver?: string;
  productionOlapDsn?: string;
  productionSlots?: string;
  public?: boolean;
  region?: string;
  updatedOn?: string;
  variables?: V1ProjectVariables;
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

export interface V1Member {
  createdOn?: string;
  roleName?: string;
  updatedOn?: string;
  userEmail?: string;
  userId?: string;
  userName?: string;
}

export interface V1ListProjectsForOrganizationResponse {
  nextPageToken?: string;
  projects?: V1Project[];
}

export interface V1ListProjectMembersResponse {
  members?: V1Member[];
  nextPageToken?: string;
}

export interface V1ListOrganizationsResponse {
  nextPageToken?: string;
  organizations?: V1Organization[];
}

export interface V1ListOrganizationMembersResponse {
  members?: V1Member[];
  nextPageToken?: string;
}

export interface V1LeaveOrganizationResponse {
  [key: string]: any;
}

export interface V1GetProjectResponse {
  jwt?: string;
  productionDeployment?: V1Deployment;
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

export type V1DeploymentStatus =
  typeof V1DeploymentStatus[keyof typeof V1DeploymentStatus];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1DeploymentStatus = {
  DEPLOYMENT_STATUS_UNSPECIFIED: "DEPLOYMENT_STATUS_UNSPECIFIED",
  DEPLOYMENT_STATUS_PENDING: "DEPLOYMENT_STATUS_PENDING",
  DEPLOYMENT_STATUS_OK: "DEPLOYMENT_STATUS_OK",
  DEPLOYMENT_STATUS_RECONCILING: "DEPLOYMENT_STATUS_RECONCILING",
  DEPLOYMENT_STATUS_ERROR: "DEPLOYMENT_STATUS_ERROR",
} as const;

export interface V1Deployment {
  branch?: string;
  createdOn?: string;
  id?: string;
  logs?: string;
  projectId?: string;
  runtimeHost?: string;
  runtimeInstanceId?: string;
  slots?: string;
  status?: V1DeploymentStatus;
  updatedOn?: string;
}

export interface V1DeleteProjectResponse {
  [key: string]: any;
}

export interface V1DeleteOrganizationResponse {
  [key: string]: any;
}

export interface V1CreateProjectResponse {
  project?: V1Project;
  projectUrl?: string;
}

export interface V1CreateOrganizationResponse {
  organization?: V1Organization;
}

export interface V1CreateOrganizationRequest {
  description?: string;
  name?: string;
}

export interface V1AddProjectMemberResponse {
  [key: string]: any;
}

export interface V1AddOrganizationMemberResponse {
  [key: string]: any;
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
