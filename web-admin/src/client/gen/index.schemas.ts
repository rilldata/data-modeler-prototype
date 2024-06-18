/**
 * Generated by orval v6.12.0 🍺
 * Do not edit manually.
 * rill/admin/v1/ai.proto
 * OpenAPI spec version: version not set
 */
export type AdminServiceSearchUsersParams = {
  emailPattern?: string;
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceListBookmarksParams = {
  projectId?: string;
  resourceKind?: string;
  resourceName?: string;
};

export type AdminServiceGetUserParams = { email?: string };

export type AdminServiceSudoGetResourceParams = {
  userId?: string;
  orgId?: string;
  projectId?: string;
  deploymentId?: string;
  instanceId?: string;
};

export type AdminServiceSearchProjectNamesParams = {
  namePattern?: string;
  annotations?: string;
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceGetReportMetaBodyAnnotations = {
  [key: string]: string;
};

export type AdminServiceGetReportMetaBody = {
  branch?: string;
  report?: string;
  annotations?: AdminServiceGetReportMetaBodyAnnotations;
  executionTime?: string;
};

export type AdminServicePullVirtualRepoParams = {
  branch?: string;
  pageSize?: number;
  pageToken?: string;
};

export type AdminServiceGetRepoMetaParams = { branch?: string };

export type AdminServiceGetAlertMetaBodyAnnotations = { [key: string]: string };

export type AdminServiceGetAlertMetaBody = {
  branch?: string;
  alert?: string;
  annotations?: AdminServiceGetAlertMetaBodyAnnotations;
  queryForUserId?: string;
  queryForUserEmail?: string;
};

export type AdminServiceUpdateServiceBody = {
  newName?: string;
};

export type AdminServiceCreateServiceParams = { name?: string };

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
  provisioner?: string;
  newName?: string;
  prodTtlSeconds?: string;
  prodVersion?: string;
};

export type AdminServiceGetProjectParams = { accessTokenTtlSeconds?: number };

export type AdminServiceCreateProjectBodyVariables = { [key: string]: string };

export type AdminServiceCreateProjectBody = {
  name?: string;
  description?: string;
  public?: boolean;
  provisioner?: string;
  prodOlapDriver?: string;
  prodOlapDsn?: string;
  prodSlots?: string;
  subpath?: string;
  prodBranch?: string;
  githubUrl?: string;
  variables?: AdminServiceCreateProjectBodyVariables;
  prodVersion?: string;
};

export type AdminServiceListProjectsForOrganizationParams = {
  pageSize?: number;
  pageToken?: string;
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

/**
 * DEPRECATED: Additional parameters to set outright in the generated URL query.
 */
export type AdminServiceGetIFrameBodyQuery = { [key: string]: string };

/**
 * If set, will use the provided attributes outright.
 */
export type AdminServiceGetIFrameBodyAttributes = { [key: string]: any };

/**
 * GetIFrameRequest is the request payload for AdminService.GetIFrame.
 */
export type AdminServiceGetIFrameBody = {
  /** Branch to embed. If not set, the production branch is used. */
  branch?: string;
  /** TTL for the iframe's access token. If not set, defaults to 24 hours. */
  ttlSeconds?: number;
  /** If set, will use the attributes of the user with this ID. */
  userId?: string;
  /** If set, will generate attributes corresponding to a user with this email. */
  userEmail?: string;
  /** If set, will use the provided attributes outright. */
  attributes?: AdminServiceGetIFrameBodyAttributes;
  /** Kind of resource to embed. If not set, defaults to "rill.runtime.v1.MetricsView". */
  kind?: string;
  /** Name of the resource to embed. This should identify a resource that is valid for embedding, such as a dashboard or component. */
  resource?: string;
  /** Theme to use for the embedded resource. */
  theme?: string;
  /** Navigation denotes whether navigation between different resources should be enabled in the embed. */
  navigation?: boolean;
  /** Blob containing UI state for rendering the initial embed. Not currently supported. */
  state?: string;
  /** DEPRECATED: Additional parameters to set outright in the generated URL query. */
  query?: AdminServiceGetIFrameBodyQuery;
};

export type AdminServiceGetDeploymentCredentialsBodyAttributes = {
  [key: string]: any;
};

export type AdminServiceGetDeploymentCredentialsBody = {
  branch?: string;
  ttlSeconds?: number;
  userId?: string;
  userEmail?: string;
  attributes?: AdminServiceGetDeploymentCredentialsBodyAttributes;
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

export type AdminServiceGetGithubRepoStatusParams = { githubUrl?: string };

export type AdminServiceTriggerRefreshSourcesBody = {
  sources?: string[];
};

export type AdminServiceCreateReportBodyBody = {
  options?: V1ReportOptions;
};

export type AdminServiceCreateAlertBodyBody = {
  options?: V1AlertOptions;
};

export type AdminServiceAddOrganizationMemberBodyBody = {
  email?: string;
  role?: string;
};

export type AdminServiceCreateProjectWhitelistedDomainBodyBody = {
  domain?: string;
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

export interface V1UpdateBookmarkResponse {
  [key: string]: any;
}

export interface V1UpdateBookmarkRequest {
  bookmarkId?: string;
  displayName?: string;
  description?: string;
  data?: string;
  default?: boolean;
  shared?: boolean;
}

export interface V1UnsubscribeReportResponse {
  [key: string]: any;
}

export interface V1UnsubscribeAlertResponse {
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

export interface V1SudoUpdateAnnotationsResponse {
  project?: V1Project;
}

export type V1SudoUpdateAnnotationsRequestAnnotations = {
  [key: string]: string;
};

export interface V1SudoUpdateAnnotationsRequest {
  organization?: string;
  project?: string;
  annotations?: V1SudoUpdateAnnotationsRequestAnnotations;
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
  intervalDuration?: string;
  queryName?: string;
  queryArgsJson?: string;
  exportLimit?: string;
  exportFormat?: V1ExportFormat;
  openProjectSubpath?: string;
  emailRecipients?: string[];
  slackUsers?: string[];
  slackChannels?: string[];
  slackWebhooks?: string[];
  webShowPage?: string;
}

export interface V1RemoveWhitelistedDomainResponse {
  [key: string]: any;
}

export interface V1RemoveProjectWhitelistedDomainResponse {
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

export interface V1RecordEventsResponse {
  [key: string]: any;
}

export type V1RecordEventsRequestEventsItem = { [key: string]: any };

export interface V1RecordEventsRequest {
  events?: V1RecordEventsRequestEventsItem[];
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
  createAlerts?: boolean;
  manageAlerts?: boolean;
}

export type V1ProjectAnnotations = { [key: string]: string };

export interface V1Project {
  id?: string;
  name?: string;
  orgId?: string;
  orgName?: string;
  description?: string;
  public?: boolean;
  createdByUserId?: string;
  provisioner?: string;
  githubUrl?: string;
  subpath?: string;
  prodBranch?: string;
  prodOlapDriver?: string;
  prodOlapDsn?: string;
  prodSlots?: string;
  prodDeploymentId?: string;
  frontendUrl?: string;
  prodTtlSeconds?: string;
  annotations?: V1ProjectAnnotations;
  prodVersion?: string;
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

export interface V1ListProjectWhitelistedDomainsResponse {
  domains?: V1WhitelistedDomain[];
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

export type V1GithubPermission =
  (typeof V1GithubPermission)[keyof typeof V1GithubPermission];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1GithubPermission = {
  GITHUB_PERMISSION_UNSPECIFIED: "GITHUB_PERMISSION_UNSPECIFIED",
  GITHUB_PERMISSION_READ: "GITHUB_PERMISSION_READ",
  GITHUB_PERMISSION_WRITE: "GITHUB_PERMISSION_WRITE",
} as const;

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

export interface V1GetIFrameResponse {
  iframeSrc?: string;
  runtimeHost?: string;
  instanceId?: string;
  accessToken?: string;
  ttlSeconds?: number;
}

export type V1GetGithubUserStatusResponseOrganizationInstallationPermissions = {
  [key: string]: V1GithubPermission;
};

export interface V1GetGithubUserStatusResponse {
  hasAccess?: boolean;
  grantAccessUrl?: string;
  accessToken?: string;
  account?: string;
  userInstallationPermission?: V1GithubPermission;
  organizationInstallationPermissions?: V1GetGithubUserStatusResponseOrganizationInstallationPermissions;
  /** DEPRECATED: Use organization_installation_permissions instead. */
  organizations?: string[];
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
  instanceId?: string;
  accessToken?: string;
  ttlSeconds?: number;
}

export interface V1GetCurrentUserResponse {
  user?: V1User;
  preferences?: V1UserPreferences;
}

export interface V1GetBookmarkResponse {
  bookmark?: V1Bookmark;
}

export interface V1GetAlertYAMLResponse {
  yaml?: string;
}

export type V1GetAlertMetaResponseQueryForAttributes = { [key: string]: any };

export interface V1GetAlertMetaResponse {
  openUrl?: string;
  editUrl?: string;
  queryForAttributes?: V1GetAlertMetaResponseQueryForAttributes;
}

export interface V1GenerateReportYAMLResponse {
  yaml?: string;
}

export interface V1GenerateAlertYAMLResponse {
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

export interface V1EditAlertResponse {
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

export interface V1DeleteAlertResponse {
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

export interface V1CreateProjectWhitelistedDomainResponse {
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

export interface V1CreateBookmarkResponse {
  bookmark?: V1Bookmark;
}

export interface V1CreateBookmarkRequest {
  displayName?: string;
  description?: string;
  data?: string;
  resourceKind?: string;
  resourceName?: string;
  projectId?: string;
  default?: boolean;
  shared?: boolean;
}

export interface V1CreateAlertResponse {
  name?: string;
}

export interface V1CompletionMessage {
  role?: string;
  data?: string;
}

export interface V1CompleteResponse {
  message?: V1CompletionMessage;
}

export interface V1CompleteRequest {
  messages?: V1CompletionMessage[];
}

export interface V1Bookmark {
  id?: string;
  displayName?: string;
  description?: string;
  data?: string;
  resourceKind?: string;
  resourceName?: string;
  projectId?: string;
  userId?: string;
  default?: boolean;
  shared?: boolean;
  createdOn?: string;
  updatedOn?: string;
}

export interface V1AlertOptions {
  title?: string;
  intervalDuration?: string;
  queryName?: string;
  queryArgsJson?: string;
  metricsViewName?: string;
  renotify?: boolean;
  renotifyAfterSeconds?: number;
  emailRecipients?: string[];
  slackUsers?: string[];
  slackChannels?: string[];
  slackWebhooks?: string[];
  metricsViewMeasureFilterIndices?: string;
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
