/**
 * Generated by orval v6.16.0 🍺
 * Do not edit manually.
 * rill/admin/v1/api.proto
 * OpenAPI spec version: version not set
 */
export type AdminServiceSearchUsersParams = {
  emailPattern?: string;
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceListBookmarksParams = {
  projectId?: string;
};

export type AdminServiceGetUserParams = {
  email?: string;
};

export type AdminServiceSudoGetResourceParams = {
  userId?: string;
  orgId?: string;
  projectId?: string;
  deploymentId?: string;
  instanceId?: string;
};

export type AdminServiceSearchProjectNamesParams = {
  namePattern?: string;
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceGetReportMetaParams = {
  branch?: string;
  report?: string;
  /**
   * This is a request variable of the map type. The query format is "map_name[key]=value", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age["bob"]=18
   */
  annotations?: string;
};

export type AdminServicePullVirtualRepoParams = {
  branch?: string;
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceGetRepoMetaParams = {
  branch?: string;
};

export type AdminServicePingParams = {
  event?: string;
};

export type AdminServiceUpdateServiceBody = {
  newName?: string;
};

export type AdminServiceCreateServiceParams = {
  name?: string;
};

export type AdminServiceUpdateProjectVariablesBodyVariables = {
  [key: string]: string;
};

export type AdminServiceUpdateProjectVariablesBody = {
  variables?: AdminServiceUpdateProjectVariablesBodyVariables;
};

export type AdminServiceUpdateProjectBody = {
  description?: string;
  public?: boolean;
  prodBranch?: string;
  githubUrl?: string;
  prodSlots?: string;
  region?: string;
  newName?: string;
  prodTtlSeconds?: string;
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

export type AdminServiceCreateWhitelistedDomainBody = {
  domain?: string;
  role?: string;
};

export type AdminServiceSearchProjectUsersParams = {
  emailQuery?: string;
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

export type AdminServiceGetDeploymentCredentialsBodyAttrs = {
  [key: string]: any;
};

export type AdminServiceGetDeploymentCredentialsBody = {
  branch?: string;
  userId?: string;
  attrs?: AdminServiceGetDeploymentCredentialsBodyAttrs;
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

export type AdminServiceUpdateOrganizationBody = {
  description?: string;
  newName?: string;
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

export type AdminServiceCreateReportBodyBody = {
  options?: V1ReportOptions;
};

export type AdminServiceAddOrganizationMemberBodyBody = {
  email?: string;
  role?: string;
};

export type AdminServiceSetOrganizationMemberRoleBodyBody = {
  role?: string;
};

export type AdminServiceTriggerReconcileBodyBody = { [key: string]: any };

export interface V1WhitelistedDomain {
  domain?: string;
  role?: string;
}

export interface V1VirtualFile {
  path?: string;
  data?: string;
  deleted?: boolean;
  updatedOn?: string;
}

export interface V1UserQuotas {
  singleuserOrgs?: number;
}

export interface V1UserPreferences {
  timeZone?: string;
}

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
  quotas?: V1UserQuotas;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1UpdateUserPreferencesResponse {
  preferences?: V1UserPreferences;
}

export interface V1UpdateUserPreferencesRequest {
  preferences?: V1UserPreferences;
}

export interface V1UpdateServiceResponse {
  service?: V1Service;
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

export interface V1UnsubscribeReportResponse {
  [key: string]: any;
}

export interface V1TriggerReportResponse {
  [key: string]: any;
}

export interface V1TriggerRefreshSourcesResponse {
  [key: string]: any;
}

export interface V1TriggerRedeployResponse {
  [key: string]: any;
}

export interface V1TriggerRedeployRequest {
  organization?: string;
  project?: string;
  deploymentId?: string;
}

export interface V1TriggerReconcileResponse {
  [key: string]: any;
}

export interface V1TrackResponse {
  [key: string]: any;
}

export interface V1TrackRequest {
  [key: string]: any;
}

export interface V1SudoUpdateUserQuotasResponse {
  user?: V1User;
}

export interface V1SudoUpdateUserQuotasRequest {
  email?: string;
  singleuserOrgs?: number;
}

export interface V1SudoUpdateOrganizationQuotasResponse {
  organization?: V1Organization;
}

export interface V1SudoUpdateOrganizationQuotasRequest {
  orgName?: string;
  projects?: number;
  deployments?: number;
  slotsTotal?: number;
  slotsPerDeployment?: number;
  outstandingInvites?: number;
}

export interface V1SudoGetResourceResponse {
  user?: V1User;
  org?: V1Organization;
  project?: V1Project;
  deployment?: V1Deployment;
  instance?: V1Deployment;
}

export interface V1SetSuperuserResponse {
  [key: string]: any;
}

export interface V1SetSuperuserRequest {
  email?: string;
  superuser?: boolean;
}

export interface V1SetProjectMemberRoleResponse {
  [key: string]: any;
}

export interface V1SetOrganizationMemberRoleResponse {
  [key: string]: any;
}

export interface V1ServiceToken {
  id?: string;
  createdOn?: string;
  expiresOn?: string;
}

export interface V1Service {
  id?: string;
  name?: string;
  orgId?: string;
  orgName?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1SearchUsersResponse {
  users?: V1User[];
  nextPageToken?: string;
}

export interface V1SearchProjectUsersResponse {
  users?: V1User[];
  nextPageToken?: string;
}

export interface V1SearchProjectNamesResponse {
  names?: string[];
  nextPageToken?: string;
}

export interface V1RevokeServiceAuthTokenResponse {
  [key: string]: any;
}

export interface V1RevokeCurrentAuthTokenResponse {
  tokenId?: string;
}

export interface V1ReportOptions {
  title?: string;
  refreshCron?: string;
  refreshTimeZone?: string;
  queryName?: string;
  queryArgsJson?: string;
  exportLimit?: string;
  exportFormat?: V1ExportFormat;
  openProjectSubpath?: string;
  recipients?: string[];
}

export interface V1RemoveWhitelistedDomainResponse {
  [key: string]: any;
}

export interface V1RemoveProjectMemberResponse {
  [key: string]: any;
}

export interface V1RemoveOrganizationMemberResponse {
  [key: string]: any;
}

export interface V1RemoveBookmarkResponse {
  [key: string]: any;
}

export interface V1PullVirtualRepoResponse {
  files?: V1VirtualFile[];
  nextPageToken?: string;
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
  createReports?: boolean;
  manageReports?: boolean;
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
  prodTtlSeconds?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1PingResponse {
  version?: string;
  time?: string;
}

export interface V1OrganizationQuotas {
  projects?: number;
  deployments?: number;
  slotsTotal?: number;
  slotsPerDeployment?: number;
  outstandingInvites?: number;
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
  quotas?: V1OrganizationQuotas;
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

export interface V1ListWhitelistedDomainsResponse {
  domains?: V1WhitelistedDomain[];
}

export interface V1ListSuperusersResponse {
  users?: V1User[];
}

export interface V1ListServicesResponse {
  services?: V1Service[];
}

export interface V1ListServiceAuthTokensResponse {
  tokens?: V1ServiceToken[];
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

export interface V1ListBookmarksResponse {
  bookmarks?: V1Bookmark[];
}

export interface V1LeaveOrganizationResponse {
  [key: string]: any;
}

export interface V1IssueServiceAuthTokenResponse {
  token?: string;
}

export interface V1IssueRepresentativeAuthTokenResponse {
  token?: string;
}

export interface V1IssueRepresentativeAuthTokenRequest {
  email?: string;
  ttlMinutes?: string;
}

export interface V1GetUserResponse {
  user?: V1User;
}

export interface V1GetReportMetaResponse {
  openUrl?: string;
  exportUrl?: string;
  editUrl?: string;
}

export interface V1GetRepoMetaResponse {
  gitUrl?: string;
  gitUrlExpiresOn?: string;
  gitSubpath?: string;
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

export interface V1GetGitCredentialsResponse {
  repoUrl?: string;
  username?: string;
  password?: string;
  subpath?: string;
  prodBranch?: string;
}

export interface V1GetDeploymentCredentialsResponse {
  runtimeHost?: string;
  runtimeInstanceId?: string;
  jwt?: string;
}

export interface V1GetCurrentUserResponse {
  user?: V1User;
  preferences?: V1UserPreferences;
}

export interface V1GetBookmarkResponse {
  bookmark?: V1Bookmark;
}

export interface V1GenerateReportYAMLResponse {
  yaml?: string;
}

export type V1ExportFormat =
  (typeof V1ExportFormat)[keyof typeof V1ExportFormat];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1ExportFormat = {
  EXPORT_FORMAT_UNSPECIFIED: "EXPORT_FORMAT_UNSPECIFIED",
  EXPORT_FORMAT_CSV: "EXPORT_FORMAT_CSV",
  EXPORT_FORMAT_XLSX: "EXPORT_FORMAT_XLSX",
  EXPORT_FORMAT_PARQUET: "EXPORT_FORMAT_PARQUET",
} as const;

export interface V1EditReportResponse {
  [key: string]: any;
}

export type V1DeploymentStatus =
  (typeof V1DeploymentStatus)[keyof typeof V1DeploymentStatus];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1DeploymentStatus = {
  DEPLOYMENT_STATUS_UNSPECIFIED: "DEPLOYMENT_STATUS_UNSPECIFIED",
  DEPLOYMENT_STATUS_PENDING: "DEPLOYMENT_STATUS_PENDING",
  DEPLOYMENT_STATUS_OK: "DEPLOYMENT_STATUS_OK",
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
  statusMessage?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1DeleteServiceResponse {
  service?: V1Service;
}

export interface V1DeleteReportResponse {
  [key: string]: any;
}

export interface V1DeleteProjectResponse {
  [key: string]: any;
}

export interface V1DeleteOrganizationResponse {
  [key: string]: any;
}

export interface V1CreateWhitelistedDomainResponse {
  [key: string]: any;
}

export interface V1CreateServiceResponse {
  service?: V1Service;
}

export interface V1CreateReportResponse {
  name?: string;
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

export interface V1CreateBookmarkResponse {
  bookmark?: V1Bookmark;
}

export interface V1CreateBookmarkRequest {
  displayName?: string;
  data?: string;
  dashboardName?: string;
  projectId?: string;
}

export interface V1Bookmark {
  id?: string;
  displayName?: string;
  data?: string;
  dashboardName?: string;
  projectId?: string;
  userId?: string;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1AddProjectMemberResponse {
  pendingSignup?: boolean;
}

export interface V1AddOrganizationMemberResponse {
  pendingSignup?: boolean;
}

export interface RpcStatus {
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}

/**
 * `NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

 - NULL_VALUE: Null value.
 */
export type ProtobufNullValue =
  (typeof ProtobufNullValue)[keyof typeof ProtobufNullValue];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ProtobufNullValue = {
  NULL_VALUE: "NULL_VALUE",
} as const;

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: unknown;
}
