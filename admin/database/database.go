package database

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Drivers is a registry of drivers
var Drivers = make(map[string]Driver)

// Register registers a new driver.
func Register(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("already registered database driver with name '%s'", name))
	}
	Drivers[name] = driver
}

// Open opens a new database connection.
// See ParseEncryptionKeyring for the expected format for encKeyringConfig.
func Open(driver, dsn, encKeyringConfig string) (DB, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	encKeyring, err := ParseEncryptionKeyring(encKeyringConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing encryption keyring: %w", err)
	}

	db, err := d.Open(dsn, encKeyring)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Driver is the interface for DB drivers.
type Driver interface {
	Open(dsn string, encKeyring []*EncryptionKey) (DB, error)
}

// DB is the interface for a database connection.
type DB interface {
	Close() error
	NewTx(ctx context.Context) (context.Context, Tx, error)

	Migrate(ctx context.Context) error
	FindMigrationVersion(ctx context.Context) (int, error)

	FindOrganizations(ctx context.Context, afterName string, limit int) ([]*Organization, error)
	FindOrganizationsForUser(ctx context.Context, userID string, afterName string, limit int) ([]*Organization, error)
	FindOrganization(ctx context.Context, id string) (*Organization, error)
	FindOrganizationByName(ctx context.Context, name string) (*Organization, error)
	FindOrganizationByCustomDomain(ctx context.Context, domain string) (*Organization, error)
	CheckOrganizationHasOutsideUser(ctx context.Context, orgID, userID string) (bool, error)
	CheckOrganizationHasPublicProjects(ctx context.Context, orgID string) (bool, error)
	InsertOrganization(ctx context.Context, opts *InsertOrganizationOptions) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error
	UpdateOrganization(ctx context.Context, id string, opts *UpdateOrganizationOptions) (*Organization, error)
	UpdateOrganizationAllUsergroup(ctx context.Context, orgID, groupID string) (*Organization, error)

	FindOrganizationWhitelistedDomain(ctx context.Context, orgID string, domain string) (*OrganizationWhitelistedDomain, error)
	FindOrganizationWhitelistedDomainForOrganizationWithJoinedRoleNames(ctx context.Context, orgID string) ([]*OrganizationWhitelistedDomainWithJoinedRoleNames, error)
	FindOrganizationWhitelistedDomainsForDomain(ctx context.Context, domain string) ([]*OrganizationWhitelistedDomain, error)
	InsertOrganizationWhitelistedDomain(ctx context.Context, opts *InsertOrganizationWhitelistedDomainOptions) (*OrganizationWhitelistedDomain, error)
	DeleteOrganizationWhitelistedDomain(ctx context.Context, id string) error

	FindProjects(ctx context.Context, afterName string, limit int) ([]*Project, error)
	FindProjectsByVersion(ctx context.Context, version, afterName string, limit int) ([]*Project, error)
	FindProjectPathsByPattern(ctx context.Context, namePattern, afterName string, limit int) ([]string, error)
	FindProjectPathsByPatternAndAnnotations(ctx context.Context, namePattern, afterName string, annotationKeys []string, annotationPairs map[string]string, limit int) ([]string, error)
	FindProjectsForUser(ctx context.Context, userID string) ([]*Project, error)
	FindProjectsForOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*Project, error)
	// FindProjectsForOrgAndUser lists the public projects in the org and the projects where user is added as an external user
	FindProjectsForOrgAndUser(ctx context.Context, orgID, userID, afterProjectName string, limit int) ([]*Project, error)
	FindPublicProjectsInOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*Project, error)
	FindProjectsByGithubURL(ctx context.Context, githubURL string) ([]*Project, error)
	FindProjectsByGithubInstallationID(ctx context.Context, id int64) ([]*Project, error)
	FindProject(ctx context.Context, id string) (*Project, error)
	FindProjectByName(ctx context.Context, orgName string, name string) (*Project, error)
	InsertProject(ctx context.Context, opts *InsertProjectOptions) (*Project, error)
	DeleteProject(ctx context.Context, id string) error
	UpdateProject(ctx context.Context, id string, opts *UpdateProjectOptions) (*Project, error)
	CountProjectsForOrganization(ctx context.Context, orgID string) (int, error)
	CountProjectsQuotaUsage(ctx context.Context, orgID string) (*ProjectsQuotaUsage, error)
	FindProjectWhitelistedDomain(ctx context.Context, projectID, domain string) (*ProjectWhitelistedDomain, error)
	FindProjectWhitelistedDomainForProjectWithJoinedRoleNames(ctx context.Context, projectID string) ([]*ProjectWhitelistedDomainWithJoinedRoleNames, error)
	FindProjectWhitelistedDomainsForDomain(ctx context.Context, domain string) ([]*ProjectWhitelistedDomain, error)
	InsertProjectWhitelistedDomain(ctx context.Context, opts *InsertProjectWhitelistedDomainOptions) (*ProjectWhitelistedDomain, error)
	DeleteProjectWhitelistedDomain(ctx context.Context, id string) error

	FindDeployments(ctx context.Context, afterID string, limit int) ([]*Deployment, error)
	FindExpiredDeployments(ctx context.Context) ([]*Deployment, error)
	FindDeploymentsForProject(ctx context.Context, projectID string) ([]*Deployment, error)
	FindDeployment(ctx context.Context, id string) (*Deployment, error)
	FindDeploymentByInstanceID(ctx context.Context, instanceID string) (*Deployment, error)
	InsertDeployment(ctx context.Context, opts *InsertDeploymentOptions) (*Deployment, error)
	DeleteDeployment(ctx context.Context, id string) error
	UpdateDeployment(ctx context.Context, id string, opts *UpdateDeploymentOptions) (*Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id string, status DeploymentStatus, msg string) (*Deployment, error)
	UpdateDeploymentUsedOn(ctx context.Context, ids []string) error

	// UpsertStaticRuntimeAssignment tracks the host and slots registered for a provisioner resource.
	// It is used by the "static" runtime provisioner to track slot usage on each host.
	UpsertStaticRuntimeAssignment(ctx context.Context, id string, host string, slots int) error
	// DeleteStaticRuntimeAssignment removes the host and slots assignment for a provisioner resource.
	// The implementation should be idempotent.
	DeleteStaticRuntimeAssignment(ctx context.Context, id string) error
	// ResolveStaticRuntimeSlotsUsed returns the current slot usage for each runtime host as tracked by UpsertStaticRuntimeAssignment.
	ResolveStaticRuntimeSlotsUsed(ctx context.Context) ([]*StaticRuntimeSlotsUsed, error)

	FindUsers(ctx context.Context) ([]*User, error)
	FindUsersByEmailPattern(ctx context.Context, emailPattern, afterEmail string, limit int) ([]*User, error)
	FindUser(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	InsertUser(ctx context.Context, opts *InsertUserOptions) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, id string, opts *UpdateUserOptions) (*User, error)
	UpdateUserActiveOn(ctx context.Context, ids []string) error
	CheckUsersEmpty(ctx context.Context) (bool, error)
	FindSuperusers(ctx context.Context) ([]*User, error)
	UpdateSuperuser(ctx context.Context, userID string, superuser bool) error
	CheckUserIsAnOrganizationMember(ctx context.Context, userID, orgID string) (bool, error)
	CheckUserIsAProjectMember(ctx context.Context, userID, projectID string) (bool, error)
	GetCurrentTrialOrgCount(ctx context.Context, userID string) (int, error)
	IncrementCurrentTrialOrgCount(ctx context.Context, userID string) error

	InsertUsergroup(ctx context.Context, opts *InsertUsergroupOptions) (*Usergroup, error)
	UpdateUsergroupName(ctx context.Context, name, groupID string) (*Usergroup, error)
	UpdateUsergroupDescription(ctx context.Context, description, groupID string) (*Usergroup, error)
	DeleteUsergroup(ctx context.Context, groupID string) error
	FindUsergroupByName(ctx context.Context, orgName, name string) (*Usergroup, error)
	FindUsergroupsForUser(ctx context.Context, userID, orgID string) ([]*Usergroup, error)
	InsertUsergroupMemberUser(ctx context.Context, groupID, userID string) error
	FindUsergroupMemberUsers(ctx context.Context, groupID, afterEmail string, limit int) ([]*MemberUser, error)
	DeleteUsergroupMemberUser(ctx context.Context, groupID, userID string) error
	DeleteUsergroupsMemberUser(ctx context.Context, orgID, userID string) error
	CheckUsergroupExists(ctx context.Context, groupID string) (bool, error)

	FindUserAuthTokens(ctx context.Context, userID string) ([]*UserAuthToken, error)
	FindUserAuthToken(ctx context.Context, id string) (*UserAuthToken, error)
	InsertUserAuthToken(ctx context.Context, opts *InsertUserAuthTokenOptions) (*UserAuthToken, error)
	UpdateUserAuthTokenUsedOn(ctx context.Context, ids []string) error
	DeleteUserAuthToken(ctx context.Context, id string) error
	DeleteExpiredUserAuthTokens(ctx context.Context, retention time.Duration) error

	FindServicesByOrgID(ctx context.Context, orgID string) ([]*Service, error)
	FindService(ctx context.Context, id string) (*Service, error)
	FindServiceByName(ctx context.Context, orgID, name string) (*Service, error)
	InsertService(ctx context.Context, opts *InsertServiceOptions) (*Service, error)
	DeleteService(ctx context.Context, id string) error
	UpdateService(ctx context.Context, id string, opts *UpdateServiceOptions) (*Service, error)
	UpdateServiceActiveOn(ctx context.Context, ids []string) error

	FindServiceAuthTokens(ctx context.Context, serviceID string) ([]*ServiceAuthToken, error)
	FindServiceAuthToken(ctx context.Context, id string) (*ServiceAuthToken, error)
	InsertServiceAuthToken(ctx context.Context, opts *InsertServiceAuthTokenOptions) (*ServiceAuthToken, error)
	UpdateServiceAuthTokenUsedOn(ctx context.Context, ids []string) error
	DeleteServiceAuthToken(ctx context.Context, id string) error
	DeleteExpiredServiceAuthTokens(ctx context.Context, retention time.Duration) error

	FindDeploymentAuthToken(ctx context.Context, id string) (*DeploymentAuthToken, error)
	InsertDeploymentAuthToken(ctx context.Context, opts *InsertDeploymentAuthTokenOptions) (*DeploymentAuthToken, error)
	UpdateDeploymentAuthTokenUsedOn(ctx context.Context, ids []string) error
	DeleteExpiredDeploymentAuthTokens(ctx context.Context, retention time.Duration) error

	FindMagicAuthTokensWithUser(ctx context.Context, projectID string, createdByUserID *string, afterID string, limit int) ([]*MagicAuthTokenWithUser, error)
	FindMagicAuthToken(ctx context.Context, id string, withSecret bool) (*MagicAuthToken, error)
	FindMagicAuthTokenWithUser(ctx context.Context, id string) (*MagicAuthTokenWithUser, error)
	InsertMagicAuthToken(ctx context.Context, opts *InsertMagicAuthTokenOptions) (*MagicAuthToken, error)
	UpdateMagicAuthTokenUsedOn(ctx context.Context, ids []string) error
	DeleteMagicAuthToken(ctx context.Context, id string) error
	DeleteMagicAuthTokens(ctx context.Context, ids []string) error
	DeleteExpiredMagicAuthTokens(ctx context.Context, retention time.Duration) error

	FindReportTokens(ctx context.Context, reportName string) ([]*ReportToken, error)
	FindReportTokensWithSecret(ctx context.Context, reportName string) ([]*ReportTokenWithSecret, error)
	FindReportTokenForMagicAuthToken(ctx context.Context, magicAuthTokenID string) (*ReportToken, error)
	InsertReportToken(ctx context.Context, opts *InsertReportTokenOptions) (*ReportToken, error)

	FindDeviceAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*DeviceAuthCode, error)
	FindPendingDeviceAuthCodeByUserCode(ctx context.Context, userCode string) (*DeviceAuthCode, error)
	InsertDeviceAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*DeviceAuthCode, error)
	DeleteDeviceAuthCode(ctx context.Context, deviceCode string) error
	UpdateDeviceAuthCode(ctx context.Context, id, userID string, state DeviceAuthCodeState) error
	DeleteExpiredDeviceAuthCodes(ctx context.Context, retention time.Duration) error

	FindAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error)
	InsertAuthorizationCode(ctx context.Context, code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod string, expiration time.Time) (*AuthorizationCode, error)
	DeleteAuthorizationCode(ctx context.Context, code string) error
	DeleteExpiredAuthorizationCodes(ctx context.Context, retention time.Duration) error

	FindOrganizationRole(ctx context.Context, name string) (*OrganizationRole, error)
	FindProjectRole(ctx context.Context, name string) (*ProjectRole, error)
	ResolveOrganizationRolesForUser(ctx context.Context, userID, orgID string) ([]*OrganizationRole, error)
	ResolveProjectRolesForUser(ctx context.Context, userID, projectID string) ([]*ProjectRole, error)

	FindOrganizationMemberUsers(ctx context.Context, orgID, afterEmail string, limit int) ([]*MemberUser, error)
	FindOrganizationMemberUsersByRole(ctx context.Context, orgID, roleID string) ([]*User, error)
	InsertOrganizationMemberUser(ctx context.Context, orgID, userID, roleID string) error
	DeleteOrganizationMemberUser(ctx context.Context, orgID, userID string) error
	UpdateOrganizationMemberUserRole(ctx context.Context, orgID, userID, roleID string) error
	CountSingleuserOrganizationsForMemberUser(ctx context.Context, userID string) (int, error)
	FindOrganizationMembersWithManageUsersRole(ctx context.Context, orgID string) ([]*MemberUser, error)

	FindProjectMemberUsers(ctx context.Context, projectID, afterEmail string, limit int) ([]*MemberUser, error)
	InsertProjectMemberUser(ctx context.Context, projectID, userID, roleID string) error
	DeleteProjectMemberUser(ctx context.Context, projectID, userID string) error
	DeleteAllProjectMemberUserForOrganization(ctx context.Context, orgID, userID string) error
	UpdateProjectMemberUserRole(ctx context.Context, projectID, userID, roleID string) error

	FindOrganizationMemberUsergroups(ctx context.Context, orgID, afterName string, limit int) ([]*MemberUsergroup, error)
	InsertOrganizationMemberUsergroup(ctx context.Context, groupID, orgID, roleID string) error
	UpdateOrganizationMemberUsergroup(ctx context.Context, groupID, orgID, roleID string) error
	DeleteOrganizationMemberUsergroup(ctx context.Context, groupID, orgID string) error

	FindProjectMemberUsergroups(ctx context.Context, projectID, afterName string, limit int) ([]*MemberUsergroup, error)
	InsertProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error
	UpdateProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error
	DeleteProjectMemberUsergroup(ctx context.Context, groupID, projectID string) error

	FindOrganizationInvites(ctx context.Context, orgID, afterEmail string, limit int) ([]*Invite, error)
	FindOrganizationInvitesByEmail(ctx context.Context, userEmail string) ([]*OrganizationInvite, error)
	FindOrganizationInvite(ctx context.Context, orgID, userEmail string) (*OrganizationInvite, error)
	InsertOrganizationInvite(ctx context.Context, opts *InsertOrganizationInviteOptions) error
	UpdateOrganizationInviteUsergroups(ctx context.Context, id string, groupIDs []string) error
	DeleteOrganizationInvite(ctx context.Context, id string) error
	CountInvitesForOrganization(ctx context.Context, orgID string) (int, error)
	UpdateOrganizationInviteRole(ctx context.Context, id, roleID string) error

	FindProjectInvites(ctx context.Context, projectID, afterEmail string, limit int) ([]*Invite, error)
	FindProjectInvitesByEmail(ctx context.Context, userEmail string) ([]*ProjectInvite, error)
	FindProjectInvite(ctx context.Context, projectID, userEmail string) (*ProjectInvite, error)
	InsertProjectInvite(ctx context.Context, opts *InsertProjectInviteOptions) error
	DeleteProjectInvite(ctx context.Context, id string) error
	UpdateProjectInviteRole(ctx context.Context, id, roleID string) error

	FindProjectAccessRequests(ctx context.Context, projectID, afterID string, limit int) ([]*ProjectAccessRequest, error)
	FindProjectAccessRequest(ctx context.Context, projectID, userID string) (*ProjectAccessRequest, error)
	FindProjectAccessRequestByID(ctx context.Context, id string) (*ProjectAccessRequest, error)
	InsertProjectAccessRequest(ctx context.Context, opts *InsertProjectAccessRequestOptions) (*ProjectAccessRequest, error)
	DeleteProjectAccessRequest(ctx context.Context, id string) error

	FindBookmarks(ctx context.Context, projectID, resourceKind, resourceName, userID string) ([]*Bookmark, error)
	FindBookmark(ctx context.Context, bookmarkID string) (*Bookmark, error)
	FindDefaultBookmark(ctx context.Context, projectID, resourceKind, resourceName string) (*Bookmark, error)
	InsertBookmark(ctx context.Context, opts *InsertBookmarkOptions) (*Bookmark, error)
	UpdateBookmark(ctx context.Context, opts *UpdateBookmarkOptions) error
	DeleteBookmark(ctx context.Context, bookmarkID string) error

	SearchProjectUsers(ctx context.Context, projectID, emailQuery string, afterEmail string, limit int) ([]*User, error)

	FindVirtualFiles(ctx context.Context, projectID, branch string, afterUpdatedOn time.Time, afterPath string, limit int) ([]*VirtualFile, error)
	FindVirtualFile(ctx context.Context, projectID, branch, path string) (*VirtualFile, error)
	UpsertVirtualFile(ctx context.Context, opts *InsertVirtualFileOptions) error
	UpdateVirtualFileDeleted(ctx context.Context, projectID, branch, path string) error
	DeleteExpiredVirtualFiles(ctx context.Context, retention time.Duration) error

	FindAsset(ctx context.Context, id string) (*Asset, error)
	FindUnusedAssets(ctx context.Context, limit int) ([]*Asset, error)
	InsertAsset(ctx context.Context, id string, organizationID, path, ownerID string, public bool) (*Asset, error)
	DeleteAssets(ctx context.Context, ids []string) error

	FindOrganizationIDsWithBilling(ctx context.Context) ([]string, error)
	FindOrganizationIDsWithoutBilling(ctx context.Context) ([]string, error)

	// CountBillingProjectsForOrganization counts the projects which are not hibernated and created before the given time
	CountBillingProjectsForOrganization(ctx context.Context, orgID string, createdBefore time.Time) (int, error)
	FindBillingUsageReportedOn(ctx context.Context) (time.Time, error)
	UpdateBillingUsageReportedOn(ctx context.Context, usageReportedOn time.Time) error

	FindOrganizationForPaymentCustomerID(ctx context.Context, customerID string) (*Organization, error)
	FindOrganizationForBillingCustomerID(ctx context.Context, customerID string) (*Organization, error)

	FindBillingIssuesForOrg(ctx context.Context, orgID string) ([]*BillingIssue, error)
	FindBillingIssueByTypeForOrg(ctx context.Context, orgID string, errorType BillingIssueType) (*BillingIssue, error)
	FindBillingIssueByType(ctx context.Context, errorType BillingIssueType) ([]*BillingIssue, error)
	FindBillingIssueByTypeAndOverdueProcessed(ctx context.Context, errorType BillingIssueType, overdueProcessed bool) ([]*BillingIssue, error)
	UpsertBillingIssue(ctx context.Context, opts *UpsertBillingIssueOptions) (*BillingIssue, error)
	UpdateBillingIssueOverdueAsProcessed(ctx context.Context, id string) error
	DeleteBillingIssue(ctx context.Context, id string) error
	DeleteBillingIssueByTypeForOrg(ctx context.Context, orgID string, errorType BillingIssueType) error

	FindProjectVariables(ctx context.Context, projectID string, environment *string) ([]*ProjectVariable, error)
	UpsertProjectVariable(ctx context.Context, projectID, environment string, vars map[string]string, userID string) ([]*ProjectVariable, error)
	DeleteProjectVariables(ctx context.Context, projectID, environment string, vars []string) error

	FindProvisionerResourcesForDeployment(ctx context.Context, deploymentID string) ([]*ProvisionerResource, error)
	FindProvisionerResourceByTypeAndName(ctx context.Context, deploymentID, typ, name string) (*ProvisionerResource, error)
	InsertProvisionerResource(ctx context.Context, opts *InsertProvisionerResourceOptions) (*ProvisionerResource, error)
	UpdateProvisionerResource(ctx context.Context, id string, opts *UpdateProvisionerResourceOptions) (*ProvisionerResource, error)
	DeleteProvisionerResource(ctx context.Context, id string) error
}

// Tx represents a database transaction. It can only be used to commit and rollback transactions.
// Actual database calls should be made by passing the ctx returned from DB.NewTx to functions on the DB.
type Tx interface {
	// Commit commits the transaction
	Commit() error
	// Rollback discards the transaction *unless* it has already been committed.
	// It does nothing if Commit has already been called.
	// This means that a call to Rollback should almost always be defer'ed right after a call to NewTx.
	Rollback() error
}

// Organization represents a tenant.
type Organization struct {
	ID                                  string
	Name                                string
	DisplayName                         string `db:"display_name"`
	Description                         string
	LogoAssetID                         *string   `db:"logo_asset_id"`
	FaviconAssetID                      *string   `db:"favicon_asset_id"`
	CustomDomain                        string    `db:"custom_domain"`
	AllUsergroupID                      *string   `db:"all_usergroup_id"`
	CreatedOn                           time.Time `db:"created_on"`
	UpdatedOn                           time.Time `db:"updated_on"`
	QuotaProjects                       int       `db:"quota_projects"`
	QuotaDeployments                    int       `db:"quota_deployments"`
	QuotaSlotsTotal                     int       `db:"quota_slots_total"`
	QuotaSlotsPerDeployment             int       `db:"quota_slots_per_deployment"`
	QuotaOutstandingInvites             int       `db:"quota_outstanding_invites"`
	QuotaStorageLimitBytesPerDeployment int64     `db:"quota_storage_limit_bytes_per_deployment"`
	BillingCustomerID                   string    `db:"billing_customer_id"`
	PaymentCustomerID                   string    `db:"payment_customer_id"`
	BillingEmail                        string    `db:"billing_email"`
	BillingPlanName                     *string   `db:"billing_plan_name"`
	BillingPlanDisplayName              *string   `db:"billing_plan_display_name"`
	CreatedByUserID                     *string   `db:"created_by_user_id"`
}

// InsertOrganizationOptions defines options for inserting a new org
type InsertOrganizationOptions struct {
	Name                                string `validate:"slug"`
	DisplayName                         string
	Description                         string
	LogoAssetID                         *string
	FaviconAssetID                      *string
	CustomDomain                        string `validate:"omitempty,fqdn"`
	QuotaProjects                       int
	QuotaDeployments                    int
	QuotaSlotsTotal                     int
	QuotaSlotsPerDeployment             int
	QuotaOutstandingInvites             int
	QuotaStorageLimitBytesPerDeployment int64
	BillingCustomerID                   string
	PaymentCustomerID                   string
	BillingEmail                        string
	CreatedByUserID                     *string
}

// UpdateOrganizationOptions defines options for updating an existing org
type UpdateOrganizationOptions struct {
	Name                                string `validate:"slug"`
	DisplayName                         string
	Description                         string
	LogoAssetID                         *string
	FaviconAssetID                      *string
	CustomDomain                        string `validate:"omitempty,fqdn"`
	QuotaProjects                       int
	QuotaDeployments                    int
	QuotaSlotsTotal                     int
	QuotaSlotsPerDeployment             int
	QuotaOutstandingInvites             int
	QuotaStorageLimitBytesPerDeployment int64
	BillingCustomerID                   string
	PaymentCustomerID                   string
	BillingEmail                        string
	BillingPlanName                     *string
	BillingPlanDisplayName              *string
	CreatedByUserID                     *string
}

// Project represents one Git connection.
// Projects belong to an organization.
type Project struct {
	ID              string
	OrganizationID  string `db:"org_id"`
	Name            string
	Description     string
	Public          bool
	CreatedByUserID *string `db:"created_by_user_id"`
	Provisioner     string
	// ArchiveAssetID is set when project files are managed by Rill instead of maintained in Git.
	// If ArchiveAssetID is set all git related fields will be empty.
	ArchiveAssetID               *string           `db:"archive_asset_id"`
	GithubURL                    *string           `db:"github_url"`
	GithubInstallationID         *int64            `db:"github_installation_id"`
	Subpath                      string            `db:"subpath"`
	ProdVersion                  string            `db:"prod_version"`
	ProdBranch                   string            `db:"prod_branch"`
	ProdVariables                map[string]string `db:"prod_variables"`
	ProdVariablesEncryptionKeyID string            `db:"prod_variables_encryption_key_id"`
	ProdOLAPDriver               string            `db:"prod_olap_driver"`
	ProdOLAPDSN                  string            `db:"prod_olap_dsn"`
	ProdSlots                    int               `db:"prod_slots"`
	ProdTTLSeconds               *int64            `db:"prod_ttl_seconds"`
	ProdDeploymentID             *string           `db:"prod_deployment_id"`
	Annotations                  map[string]string `db:"annotations"`
	CreatedOn                    time.Time         `db:"created_on"`
	UpdatedOn                    time.Time         `db:"updated_on"`
}

// InsertProjectOptions defines options for inserting a new Project.
type InsertProjectOptions struct {
	OrganizationID       string `validate:"required"`
	Name                 string `validate:"slug"`
	Description          string
	Public               bool
	CreatedByUserID      *string
	Provisioner          string
	ArchiveAssetID       *string
	GithubURL            *string `validate:"omitempty,http_url"`
	GithubInstallationID *int64  `validate:"omitempty,ne=0"`
	Subpath              string
	ProdVersion          string
	ProdBranch           string
	ProdOLAPDriver       string
	ProdOLAPDSN          string
	ProdSlots            int
	ProdTTLSeconds       *int64
}

// UpdateProjectOptions defines options for updating a Project.
type UpdateProjectOptions struct {
	Name                 string `validate:"slug"`
	Description          string
	Public               bool
	Provisioner          string
	ArchiveAssetID       *string
	GithubURL            *string `validate:"omitempty,http_url"`
	GithubInstallationID *int64  `validate:"omitempty,ne=0"`
	Subpath              string
	ProdVersion          string
	ProdBranch           string
	ProdDeploymentID     *string
	ProdSlots            int
	ProdTTLSeconds       *int64
	Annotations          map[string]string
}

// DeploymentStatus is an enum representing the state of a deployment
type DeploymentStatus int

const (
	DeploymentStatusUnspecified DeploymentStatus = 0
	DeploymentStatusPending     DeploymentStatus = 1
	DeploymentStatusOK          DeploymentStatus = 2
	DeploymentStatusError       DeploymentStatus = 4
)

func (d DeploymentStatus) String() string {
	switch d {
	case DeploymentStatusPending:
		return "Pending"
	case DeploymentStatusOK:
		return "OK"
	case DeploymentStatusError:
		return "Error"
	default:
		return "Unspecified"
	}
}

// Deployment is a single deployment of a git branch.
// Deployments belong to a project.
type Deployment struct {
	ID                string           `db:"id"`
	ProjectID         string           `db:"project_id"`
	Branch            string           `db:"branch"`
	RuntimeHost       string           `db:"runtime_host"`
	RuntimeInstanceID string           `db:"runtime_instance_id"`
	RuntimeAudience   string           `db:"runtime_audience"`
	Status            DeploymentStatus `db:"status"`
	StatusMessage     string           `db:"status_message"`
	CreatedOn         time.Time        `db:"created_on"`
	UpdatedOn         time.Time        `db:"updated_on"`
	UsedOn            time.Time        `db:"used_on"`
}

// InsertDeploymentOptions defines options for inserting a new Deployment.
type InsertDeploymentOptions struct {
	ProjectID         string
	Branch            string
	RuntimeHost       string
	RuntimeInstanceID string
	RuntimeAudience   string
	Status            DeploymentStatus
	StatusMessage     string
}

// UpdateDeploymentOptions defines options for updating a Deployment.
type UpdateDeploymentOptions struct {
	Branch            string
	RuntimeHost       string
	RuntimeInstanceID string
	RuntimeAudience   string
	Status            DeploymentStatus
	StatusMessage     string
}

// StaticRuntimeSlotsUsed is the number of slots currently assigned to a runtime host.
type StaticRuntimeSlotsUsed struct {
	Host  string `db:"host"`
	Slots int    `db:"slots"`
}

// User is a person registered in Rill.
// Users may belong to multiple organizations and projects.
type User struct {
	ID                    string
	Email                 string
	DisplayName           string    `db:"display_name"`
	PhotoURL              string    `db:"photo_url"`
	GithubUsername        string    `db:"github_username"`
	GithubRefreshToken    string    `db:"github_refresh_token"`
	CreatedOn             time.Time `db:"created_on"`
	UpdatedOn             time.Time `db:"updated_on"`
	ActiveOn              time.Time `db:"active_on"`
	QuotaSingleuserOrgs   int       `db:"quota_singleuser_orgs"`
	QuotaTrialOrgs        int       `db:"quota_trial_orgs"`
	CurrentTrialOrgsCount int       `db:"current_trial_orgs_count"`
	PreferenceTimeZone    string    `db:"preference_time_zone"`
	Superuser             bool      `db:"superuser"`
}

// InsertUserOptions defines options for inserting a new user
type InsertUserOptions struct {
	Email               string `validate:"email"`
	DisplayName         string
	PhotoURL            string
	QuotaSingleuserOrgs int
	QuotaTrialOrgs      int
	Superuser           bool
}

// UpdateUserOptions defines options for updating an existing user
type UpdateUserOptions struct {
	DisplayName         string
	PhotoURL            string
	GithubUsername      string
	GithubRefreshToken  string
	QuotaSingleuserOrgs int
	QuotaTrialOrgs      int
	PreferenceTimeZone  string
}

// Service represents a service account.
// Service accounts may belong to single organization
type Service struct {
	ID        string
	OrgID     string    `db:"org_id"`
	Name      string    `validate:"slug"`
	CreatedOn time.Time `db:"created_on"`
	UpdatedOn time.Time `db:"updated_on"`
	ActiveOn  time.Time `db:"active_on"`
}

// InsertServiceOptions defines options for inserting a new service
type InsertServiceOptions struct {
	OrgID string
	Name  string `validate:"slug"`
}

// UpdateServiceOptions defines options for updating an existing service
type UpdateServiceOptions struct {
	Name string `validate:"slug"`
}

// Usergroup represents a group of org members
type Usergroup struct {
	ID          string    `db:"id"`
	OrgID       string    `db:"org_id"`
	Name        string    `db:"name" validate:"slug"`
	Description string    `db:"description"`
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

// InsertUsergroupOptions defines options for inserting a new usergroup
type InsertUsergroupOptions struct {
	OrgID string
	Name  string `validate:"slug"`
}

// UserAuthToken is a persistent API token for a user.
type UserAuthToken struct {
	ID                 string
	SecretHash         []byte     `db:"secret_hash"`
	UserID             string     `db:"user_id"`
	DisplayName        string     `db:"display_name"`
	AuthClientID       *string    `db:"auth_client_id"`
	RepresentingUserID *string    `db:"representing_user_id"`
	CreatedOn          time.Time  `db:"created_on"`
	ExpiresOn          *time.Time `db:"expires_on"`
	UsedOn             time.Time  `db:"used_on"`
}

// InsertUserAuthTokenOptions defines options for creating a UserAuthToken.
type InsertUserAuthTokenOptions struct {
	ID                 string
	SecretHash         []byte
	UserID             string
	DisplayName        string
	AuthClientID       *string
	RepresentingUserID *string
	ExpiresOn          *time.Time
}

// ServiceAuthToken is a persistent API token for an external (tenant managed) service.
type ServiceAuthToken struct {
	ID         string
	SecretHash []byte     `db:"secret_hash"`
	ServiceID  string     `db:"service_id"`
	CreatedOn  time.Time  `db:"created_on"`
	ExpiresOn  *time.Time `db:"expires_on"`
	UsedOn     time.Time  `db:"used_on"`
}

// InsertServiceAuthTokenOptions defines options for creating a ServiceAuthToken.
type InsertServiceAuthTokenOptions struct {
	ID         string
	SecretHash []byte
	ServiceID  string
	ExpiresOn  *time.Time
}

// DeploymentAuthToken is a persistent API token for a deployment.
type DeploymentAuthToken struct {
	ID           string
	SecretHash   []byte     `db:"secret_hash"`
	DeploymentID string     `db:"deployment_id"`
	CreatedOn    time.Time  `db:"created_on"`
	ExpiresOn    *time.Time `db:"expires_on"`
	UsedOn       time.Time  `db:"used_on"`
}

// InsertDeploymentAuthTokenOptions defines options for creating a DeploymentAuthToken.
type InsertDeploymentAuthTokenOptions struct {
	ID           string
	SecretHash   []byte
	DeploymentID string
	ExpiresOn    *time.Time
}

// MagicAuthToken is a persistent API token for accessing a specific (filtered) resource in a project.
type MagicAuthToken struct {
	ID                    string
	SecretHash            []byte         `db:"secret_hash"`
	Secret                []byte         `db:"secret"`
	SecretEncryptionKeyID string         `db:"secret_encryption_key_id"`
	ProjectID             string         `db:"project_id"`
	CreatedOn             time.Time      `db:"created_on"`
	ExpiresOn             *time.Time     `db:"expires_on"`
	UsedOn                time.Time      `db:"used_on"`
	CreatedByUserID       *string        `db:"created_by_user_id"`
	Attributes            map[string]any `db:"attributes"`
	ResourceType          string         `db:"resource_type"`
	ResourceName          string         `db:"resource_name"`
	FilterJSON            string         `db:"filter_json"`
	Fields                []string       `db:"fields"`
	State                 string         `db:"state"`
	DisplayName           string         `db:"display_name"`
	Internal              bool           `db:"internal"`
}

// MagicAuthTokenWithUser is a MagicAuthToken with additional information about the user who created it.
type MagicAuthTokenWithUser struct {
	*MagicAuthToken
	CreatedByUserEmail string `db:"created_by_user_email"`
}

// InsertMagicAuthTokenOptions defines options for creating a MagicAuthToken.
type InsertMagicAuthTokenOptions struct {
	ID              string
	SecretHash      []byte
	Secret          []byte
	ProjectID       string `validate:"required"`
	ExpiresOn       *time.Time
	CreatedByUserID *string
	Attributes      map[string]any
	ResourceType    string `validate:"required"`
	ResourceName    string `validate:"required"`
	FilterJSON      string
	Fields          []string
	State           string
	DisplayName     string
	Internal        bool
}

type ReportToken struct {
	ID               string
	ReportName       string `db:"report_name"`
	RecipientEmail   string `db:"recipient_email"`
	MagicAuthTokenID string `db:"magic_auth_token_id"`
}

type ReportTokenWithSecret struct {
	ID                   string
	ReportName           string `db:"report_name"`
	RecipientEmail       string `db:"recipient_email"`
	MagicAuthTokenID     string `db:"magic_auth_token_id"`
	MagicAuthTokenSecret []byte `db:"magic_auth_token_secret"`
}

type InsertReportTokenOptions struct {
	ReportName       string
	RecipientEmail   string
	MagicAuthTokenID string
}

// AuthClient is a client that requests and consumes auth tokens.
type AuthClient struct {
	ID          string
	DisplayName string
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

// Hard-coded auth client IDs (created in the migrations).
const (
	AuthClientIDRillWeb      = "12345678-0000-0000-0000-000000000001"
	AuthClientIDRillCLI      = "12345678-0000-0000-0000-000000000002"
	AuthClientIDRillSupport  = "12345678-0000-0000-0000-000000000003"
	AuthClientIDRillWebLocal = "12345678-0000-0000-0000-000000000004"
)

// DeviceAuthCodeState is an enum representing the approval state of a DeviceAuthCode
type DeviceAuthCodeState int

const (
	DeviceAuthCodeStatePending  DeviceAuthCodeState = 0
	DeviceAuthCodeStateApproved DeviceAuthCodeState = 1
	DeviceAuthCodeStateRejected DeviceAuthCodeState = 2
)

// DeviceAuthCode represents a user authentication code as part of the OAuth2 Device Authorization flow.
// They're currently used for authenticating users in the CLI.
type DeviceAuthCode struct {
	ID            string              `db:"id"`
	DeviceCode    string              `db:"device_code"`
	UserCode      string              `db:"user_code"`
	Expiry        time.Time           `db:"expires_on"`
	ApprovalState DeviceAuthCodeState `db:"approval_state"`
	ClientID      string              `db:"client_id"`
	UserID        *string             `db:"user_id"`
	CreatedOn     time.Time           `db:"created_on"`
	UpdatedOn     time.Time           `db:"updated_on"`
}

// AuthorizationCode represents an authorization code used for OAuth2 PKCE auth flow.
type AuthorizationCode struct {
	ID                  string    `db:"id"`
	Code                string    `db:"code"`
	UserID              string    `db:"user_id"`
	ClientID            string    `db:"client_id"`
	RedirectURI         string    `db:"redirect_uri"`
	CodeChallenge       string    `db:"code_challenge"`
	CodeChallengeMethod string    `db:"code_challenge_method"`
	Expiration          time.Time `db:"expires_on"`
	CreatedOn           time.Time `db:"created_on"`
	UpdatedOn           time.Time `db:"updated_on"`
}

// Constants for known role names (created in migrations).
const (
	OrganizationRoleNameAdmin        = "admin"
	OrganizationRoleNameCollaborator = "collaborator"
	OrganizationRoleNameViewer       = "viewer"
	ProjectRoleNameAdmin             = "admin"
	ProjectRoleNameCollaborator      = "collaborator"
	ProjectRoleNameViewer            = "viewer"
)

// OrganizationRole represents roles for orgs.
type OrganizationRole struct {
	ID               string
	Name             string
	ReadOrg          bool `db:"read_org"`
	ManageOrg        bool `db:"manage_org"`
	ReadProjects     bool `db:"read_projects"`
	CreateProjects   bool `db:"create_projects"`
	ManageProjects   bool `db:"manage_projects"`
	ReadOrgMembers   bool `db:"read_org_members"`
	ManageOrgMembers bool `db:"manage_org_members"`
}

// ProjectRole represents roles for projects.
type ProjectRole struct {
	ID                         string
	Name                       string
	ReadProject                bool `db:"read_project"`
	ManageProject              bool `db:"manage_project"`
	ReadProd                   bool `db:"read_prod"`
	ReadProdStatus             bool `db:"read_prod_status"`
	ManageProd                 bool `db:"manage_prod"`
	ReadDev                    bool `db:"read_dev"`
	ReadDevStatus              bool `db:"read_dev_status"`
	ManageDev                  bool `db:"manage_dev"`
	ReadProvisionerResources   bool `db:"read_provisioner_resources"`
	ManageProvisionerResources bool `db:"manage_provisioner_resources"`
	ReadProjectMembers         bool `db:"read_project_members"`
	ManageProjectMembers       bool `db:"manage_project_members"`
	CreateMagicAuthTokens      bool `db:"create_magic_auth_tokens"`
	ManageMagicAuthTokens      bool `db:"manage_magic_auth_tokens"`
	CreateReports              bool `db:"create_reports"`
	ManageReports              bool `db:"manage_reports"`
	CreateAlerts               bool `db:"create_alerts"`
	ManageAlerts               bool `db:"manage_alerts"`
	CreateBookmarks            bool `db:"create_bookmarks"`
	ManageBookmarks            bool `db:"manage_bookmarks"`
}

// MemberUser is a convenience type used for display-friendly representation of an org or project member.
type MemberUser struct {
	ID          string
	Email       string
	DisplayName string    `db:"display_name"`
	PhotoURL    string    `db:"photo_url"`
	RoleName    string    `db:"name"`
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

type MemberUsergroup struct {
	ID        string    `db:"id"`
	Name      string    `db:"name" validate:"slug"`
	RoleName  string    `db:"role_name"`
	CreatedOn time.Time `db:"created_on"`
	UpdatedOn time.Time `db:"updated_on"`
}

// OrganizationInvite represents an outstanding invitation to join an org.
type OrganizationInvite struct {
	ID              string
	Email           string
	OrgID           string    `db:"org_id"`
	OrgRoleID       string    `db:"org_role_id"`
	UsergroupIDs    []string  `db:"usergroup_ids"`
	InvitedByUserID string    `db:"invited_by_user_id"`
	CreatedOn       time.Time `db:"created_on"`
}

// ProjectInvite represents an outstanding invitation to join a project.
type ProjectInvite struct {
	ID              string
	Email           string
	ProjectID       string    `db:"project_id"`
	ProjectRoleID   string    `db:"project_role_id"`
	InvitedByUserID string    `db:"invited_by_user_id"`
	CreatedOn       time.Time `db:"created_on"`
}

// Invite is a convenience type used for display-friendly representation of an OrganizationInvite or ProjectInvite.
type Invite struct {
	Email     string
	Role      string
	InvitedBy string `db:"invited_by"`
}

type ProjectsQuotaUsage struct {
	Projects    int `db:"projects"`
	Deployments int `db:"deployments"`
	Slots       int `db:"slots"`
}

type OrganizationWhitelistedDomain struct {
	ID        string
	OrgID     string `db:"org_id"`
	OrgRoleID string `db:"org_role_id"`
	Domain    string
	CreatedOn time.Time `db:"created_on"`
	UpdatedOn time.Time `db:"updated_on"`
}

type InsertOrganizationWhitelistedDomainOptions struct {
	OrgID     string `validate:"required"`
	OrgRoleID string `validate:"required"`
	Domain    string `validate:"fqdn"`
}

// OrganizationWhitelistedDomainWithJoinedRoleNames convenience type used for display-friendly representation of an OrganizationWhitelistedDomain.
type OrganizationWhitelistedDomainWithJoinedRoleNames struct {
	Domain   string
	RoleName string `db:"name"`
}

type ProjectWhitelistedDomain struct {
	ID            string
	ProjectID     string `db:"project_id"`
	ProjectRoleID string `db:"project_role_id"`
	Domain        string
	CreatedOn     time.Time `db:"created_on"`
	UpdatedOn     time.Time `db:"updated_on"`
}

type InsertProjectWhitelistedDomainOptions struct {
	ProjectID     string `validate:"required"`
	ProjectRoleID string `validate:"required"`
	Domain        string `validate:"fqdn"`
}

type ProjectWhitelistedDomainWithJoinedRoleNames struct {
	Domain   string
	RoleName string `db:"name"`
}

const (
	DefaultQuotaProjects                       = 1
	DefaultQuotaDeployments                    = 2
	DefaultQuotaSlotsTotal                     = 4
	DefaultQuotaSlotsPerDeployment             = 2
	DefaultQuotaOutstandingInvites             = 200
	DefaultQuotaSingleuserOrgs                 = 100
	DefaultQuotaTrialOrgs                      = 2
	DefaultQuotaStorageLimitBytesPerDeployment = int64(10737418240) // 10GB
)

type InsertOrganizationInviteOptions struct {
	Email     string `validate:"email"`
	InviterID string
	OrgID     string `validate:"required"`
	RoleID    string `validate:"required"`
}

type InsertProjectInviteOptions struct {
	Email     string `validate:"email"`
	InviterID string
	ProjectID string `validate:"required"`
	RoleID    string `validate:"required"`
}

type ProjectAccessRequest struct {
	ID        string
	UserID    string    `db:"user_id"`
	ProjectID string    `db:"project_id"`
	CreatedOn time.Time `db:"created_on"`
}

type InsertProjectAccessRequestOptions struct {
	UserID    string `validate:"required"`
	ProjectID string `validate:"required"`
}

type Bookmark struct {
	ID           string
	DisplayName  string    `db:"display_name"`
	Description  string    `db:"description"`
	Data         []byte    `db:"data"`
	ResourceKind string    `db:"resource_kind"`
	ResourceName string    `db:"resource_name"`
	ProjectID    string    `db:"project_id"`
	UserID       string    `db:"user_id"`
	Default      bool      `db:"default"`
	Shared       bool      `db:"shared"`
	CreatedOn    time.Time `db:"created_on"`
	UpdatedOn    time.Time `db:"updated_on"`
}

// InsertBookmarkOptions defines options for inserting a new bookmark
type InsertBookmarkOptions struct {
	DisplayName  string `json:"display_name"`
	Data         []byte `json:"data"`
	ResourceKind string `json:"resource_kind"`
	ResourceName string `json:"resource_name"`
	Description  string `json:"description"`
	ProjectID    string `json:"project_id"`
	UserID       string `json:"user_id"`
	Default      bool   `json:"default"`
	Shared       bool   `json:"shared"`
}

// UpdateBookmarkOptions defines options for updating an existing bookmark
type UpdateBookmarkOptions struct {
	BookmarkID  string `json:"bookmark_id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Data        []byte `json:"data"`
	Shared      bool   `json:"shared"`
}

// VirtualFile represents an ad-hoc file for a project (not managed in Git)
type VirtualFile struct {
	Path      string    `db:"path"`
	Data      []byte    `db:"data"`
	Deleted   bool      `db:"deleted"`
	UpdatedOn time.Time `db:"updated_on"`
}

// InsertVirtualFileOptions defines options for inserting a VirtualFile
type InsertVirtualFileOptions struct {
	ProjectID string
	Branch    string
	Path      string `validate:"required"`
	Data      []byte `validate:"max=8192"` // 8kb
}

// Asset represents a user-uploaded file asset.
// For example, this can be an upload deploy of a project or a custom logo for an org.
type Asset struct {
	ID             string
	OrganizationID *string   `db:"org_id"`
	Path           string    `db:"path"`
	OwnerID        string    `db:"owner_id"`
	Public         bool      `db:"public"`
	CreatedOn      time.Time `db:"created_on"`
}

// ProjectVariable represents a key-value variable for a project, possible for a specific environment (e.g. production or development).
type ProjectVariable struct {
	ID                   string    `db:"id"`
	ProjectID            string    `db:"project_id"`
	Environment          string    `db:"environment"`
	Name                 string    `db:"name"`
	Value                string    `db:"value"`
	ValueEncryptionKeyID string    `db:"value_encryption_key_id"`
	UpdatedByUserID      *string   `db:"updated_by_user_id"`
	CreatedOn            time.Time `db:"created_on"`
	UpdatedOn            time.Time `db:"updated_on"`
}

// EncryptionKey represents an encryption key for column-level encryption/decryption.
// Column-level encryption provides an extra layer of security for highly sensitive columns in the database.
// It is implemented on the application side before writes to and after reads from the database.
type EncryptionKey struct {
	ID     string `json:"key_id"`
	Secret []byte `json:"key"`
}

// ParseEncryptionKeyring parses a JSON string containing an array of EncryptionKey objects.
// If the provided string is empty, an empty keyring is returned.
// When using an empty keyring, columns will be read and written without applying encryption/decryption.
func ParseEncryptionKeyring(keyring string) ([]*EncryptionKey, error) {
	if keyring == "" {
		return nil, nil
	}

	var encKeyring []*EncryptionKey
	err := json.Unmarshal([]byte(keyring), &encKeyring)
	if err != nil {
		return nil, err
	}

	return encKeyring, nil
}

func NewRandomKeyring() ([]*EncryptionKey, error) {
	secret := make([]byte, 32) // 32 bytes for AES-256
	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}

	encKeyRing := []*EncryptionKey{
		{ID: uuid.New().String(), Secret: secret},
	}

	return encKeyRing, nil
}

const BillingGracePeriodDays = 9

type BillingIssueType int

const (
	BillingIssueTypeUnspecified           BillingIssueType = iota
	BillingIssueTypeOnTrial                                = 1
	BillingIssueTypeTrialEnded                             = 2
	BillingIssueTypeNoPaymentMethod                        = 3
	BillingIssueTypeNoBillableAddress                      = 4
	BillingIssueTypePaymentFailed                          = 5
	BillingIssueTypeSubscriptionCancelled                  = 6
	BillingIssueTypeNeverSubscribed                        = 7
)

type BillingIssueLevel int

const (
	BillingIssueLevelUnspecified BillingIssueLevel = iota
	BillingIssueLevelWarning                       = 1
	BillingIssueLevelError                         = 2
)

type BillingIssue struct {
	ID        string
	OrgID     string
	Type      BillingIssueType
	Level     BillingIssueLevel
	Metadata  BillingIssueMetadata
	EventTime time.Time
	CreatedOn time.Time
}

type BillingIssueMetadata interface{}

type BillingIssueMetadataOnTrial struct {
	SubID              string    `json:"subscription_id"`
	PlanID             string    `json:"plan_id"`
	EndDate            time.Time `json:"end_date"`
	GracePeriodEndDate time.Time `json:"grace_period_end_date"`
}

type BillingIssueMetadataTrialEnded struct {
	SubID              string    `json:"subscription_id"`
	PlanID             string    `json:"plan_id"`
	EndDate            time.Time `json:"end_date"`
	GracePeriodEndDate time.Time `json:"grace_period_end_date"`
}

type BillingIssueMetadataNoPaymentMethod struct{}

type BillingIssueMetadataNoBillableAddress struct{}

type BillingIssueMetadataPaymentFailed struct {
	Invoices map[string]*BillingIssueMetadataPaymentFailedMeta `json:"invoices"`
}

type BillingIssueMetadataPaymentFailedMeta struct {
	ID                 string    `json:"id"`
	Number             string    `json:"invoice_number"`
	URL                string    `json:"invoice_url"`
	Amount             string    `json:"amount"`
	Currency           string    `json:"currency"`
	DueDate            time.Time `json:"due_date"`
	FailedOn           time.Time `json:"failed_on"`
	GracePeriodEndDate time.Time `json:"grace_period_end_date"`
}

type BillingIssueMetadataSubscriptionCancelled struct {
	EndDate time.Time `json:"end_date"`
}

type BillingIssueMetadataNeverSubscribed struct{}

type UpsertBillingIssueOptions struct {
	OrgID     string           `validate:"required"`
	Type      BillingIssueType `validate:"required"`
	Metadata  BillingIssueMetadata
	EventTime time.Time `validate:"required"`
}

// ProvisionerResourceStatus is an enum representing the state of a provisioner resource
type ProvisionerResourceStatus int

const (
	ProvisionerResourceStatusUnspecified ProvisionerResourceStatus = 0
	ProvisionerResourceStatusPending     ProvisionerResourceStatus = 1
	ProvisionerResourceStatusOK          ProvisionerResourceStatus = 2
	ProvisionerResourceStatusError       ProvisionerResourceStatus = 4
)

func (d ProvisionerResourceStatus) String() string {
	switch d {
	case ProvisionerResourceStatusPending:
		return "Pending"
	case ProvisionerResourceStatusOK:
		return "OK"
	case ProvisionerResourceStatusError:
		return "Error"
	default:
		return "Unspecified"
	}
}

// ProvisionerResource represents a resource created by a provisioner (see admin/provisioner/README.md for details about provisioners).
type ProvisionerResource struct {
	ID            string                    `db:"id"`
	DeploymentID  string                    `db:"deployment_id"`
	Type          string                    `db:"type"`
	Name          string                    `db:"name"`
	Status        ProvisionerResourceStatus `db:"status"`
	StatusMessage string                    `db:"status_message"`
	Provisioner   string                    `db:"provisioner"`
	Args          map[string]any            `db:"args_json"`
	State         map[string]any            `db:"state_json"`
	Config        map[string]any            `db:"config_json"`
	CreatedOn     time.Time                 `db:"created_on"`
	UpdatedOn     time.Time                 `db:"updated_on"`
}

// InsertProvisionerResourceOptions defines options for inserting a new ProvisionerResource.
type InsertProvisionerResourceOptions struct {
	ID            string
	DeploymentID  string
	Type          string
	Name          string
	Status        ProvisionerResourceStatus
	StatusMessage string
	Provisioner   string
	Args          map[string]any
	State         map[string]any
	Config        map[string]any
}

// UpdateProvisionerResourceOptions defines options for updating a ProvisionerResource.
type UpdateProvisionerResourceOptions struct {
	Status        ProvisionerResourceStatus
	StatusMessage string
	Args          map[string]any
	State         map[string]any
	Config        map[string]any
}
