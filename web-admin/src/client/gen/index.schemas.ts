/**
 * Generated by orval v6.13.1 🍺
 * Do not edit manually.
 * rill/admin/v1/api.proto
 * OpenAPI spec version: version not set
 */
export type AdminServiceUpdateProjectVariablesBodyVariables = {
  [key: string]: string;
};

export type AdminServiceUpdateProjectVariablesBody = {
  variables?: AdminServiceUpdateProjectVariablesBodyVariables;
};

export type AdminServiceUpdateProjectBody = {
  id?: string;
  description?: string;
  public?: boolean;
  prodBranch?: string;
  githubUrl?: string;
};

export type AdminServiceCreateProjectBodyVariables = { [key: string]: string };

export type AdminServiceCreateProjectBody = {
  name?: string;
  description?: string;
  public?: boolean;
  region?: string;
  prodOlapDriver?: string;
  prodOlapDsn?: string;
  prodSlots?: string;
  subpath?: string;
  prodBranch?: string;
  githubUrl?: string;
  variables?: AdminServiceCreateProjectBodyVariables;
};

export type AdminServiceListProjectsForOrganizationParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceListProjectMembersParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceListProjectInvitesParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceRemoveOrganizationMemberParams = {
  keepProjectRoles?: boolean;
};

export type AdminServiceListOrganizationMembersParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceListOrganizationInvitesParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceCreateAutoinviteDomainBody = {
  domain?: string;
  role?: string;
};

export type AdminServiceUpdateOrganizationBody = {
  id?: string;
  description?: string;
};

export type AdminServiceListOrganizationsParams = {
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceGetGithubRepoStatusParams = {
  githubUrl?: string;
};

export type AdminServiceTriggerRefreshSourcesBody = {
  sources?: string[];
};

export type AdminServiceAddOrganizationMemberBodyBody = {
  email?: string;
  role?: string;
};

export type AdminServiceSetOrganizationMemberRoleBodyBody = {
  role?: string;
};

export type AdminServiceTriggerReconcileBodyBody = { [key: string]: any };

export interface V1UserInvite {
  email?: string;
  role?: string;
  invitedBy?: string;
}

export interface V1User {
  id?: string;
  email?: string;
  displayName?: string;
  photoUrl?: string;
  createdOn?: string;
  updatedOn?: string;
}

export type V1UpdateProjectVariablesResponseVariables = {
  [key: string]: string;
};

export interface V1UpdateProjectVariablesResponse {
  variables?: V1UpdateProjectVariablesResponseVariables;
}

export interface V1UpdateProjectResponse {
  project?: V1Project;
}

export interface V1UpdateOrganizationResponse {
  organization?: V1Organization;
}

export interface V1TriggerRefreshSourcesResponse {
  [key: string]: any;
}

export interface V1TriggerRedeployResponse {
  [key: string]: any;
}

export interface V1TriggerReconcileResponse {
  [key: string]: any;
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

export interface V1RemoveAutoinviteDomainResponse {
  [key: string]: any;
}

export interface V1ProjectPermissions {
  readProject?: boolean;
  manageProject?: boolean;
  readProd?: boolean;
  readProdStatus?: boolean;
  manageProd?: boolean;
  readDev?: boolean;
  readDevStatus?: boolean;
  manageDev?: boolean;
  readProjectMembers?: boolean;
  manageProjectMembers?: boolean;
}

export interface V1Project {
  id?: string;
  name?: string;
  orgId?: string;
  orgName?: string;
  description?: string;
  public?: boolean;
  region?: string;
  githubUrl?: string;
  subpath?: string;
  prodBranch?: string;
  prodOlapDriver?: string;
  prodOlapDsn?: string;
  prodSlots?: string;
  prodDeploymentId?: string;
  frontendUrl?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1PingResponse {
  version?: string;
  time?: string;
}

export interface V1OrganizationPermissions {
  readOrg?: boolean;
  manageOrg?: boolean;
  readProjects?: boolean;
  createProjects?: boolean;
  manageProjects?: boolean;
  readOrgMembers?: boolean;
  manageOrgMembers?: boolean;
}

export interface V1Organization {
  id?: string;
  name?: string;
  description?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1Member {
  userId?: string;
  userEmail?: string;
  userName?: string;
  roleName?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1ListProjectsForOrganizationResponse {
  projects?: V1Project[];
  nextPageToken?: string;
}

export interface V1ListProjectMembersResponse {
  members?: V1Member[];
  nextPageToken?: string;
}

export interface V1ListProjectInvitesResponse {
  invites?: V1UserInvite[];
  nextPageToken?: string;
}

export interface V1ListOrganizationsResponse {
  organizations?: V1Organization[];
  nextPageToken?: string;
}

export interface V1ListOrganizationMembersResponse {
  members?: V1Member[];
  nextPageToken?: string;
}

export interface V1ListOrganizationInvitesResponse {
  invites?: V1UserInvite[];
  nextPageToken?: string;
}

export interface V1LeaveOrganizationResponse {
  [key: string]: any;
}

export type V1GetProjectVariablesResponseVariables = { [key: string]: string };

export interface V1GetProjectVariablesResponse {
  variables?: V1GetProjectVariablesResponseVariables;
}

export interface V1GetProjectResponse {
  project?: V1Project;
  prodDeployment?: V1Deployment;
  jwt?: string;
  projectPermissions?: V1ProjectPermissions;
}

export interface V1GetOrganizationResponse {
  organization?: V1Organization;
  permissions?: V1OrganizationPermissions;
}

export interface V1GetGithubRepoStatusResponse {
  hasAccess?: boolean;
  grantAccessUrl?: string;
  defaultBranch?: string;
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
  id?: string;
  projectId?: string;
  slots?: string;
  branch?: string;
  runtimeHost?: string;
  runtimeInstanceId?: string;
  status?: V1DeploymentStatus;
  logs?: string;
  createdOn?: string;
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
}

export interface V1CreateOrganizationResponse {
  organization?: V1Organization;
}

export interface V1CreateOrganizationRequest {
  name?: string;
  description?: string;
}

export interface V1CreateAutoinviteDomainResponse {
  [key: string]: any;
}

export interface V1AddProjectMemberResponse {
  pendingSignup?: boolean;
}

export interface V1AddOrganizationMemberResponse {
  pendingSignup?: boolean;
}

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: unknown;
}

export interface RpcStatus {
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}
