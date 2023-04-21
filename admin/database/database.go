package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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
func Open(driver, dsn string) (DB, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	db, err := d.Open(dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Driver is the interface for DB drivers.
type Driver interface {
	Open(dsn string) (DB, error)
}

// DB is the interface for a database connection.
type DB interface {
	Close() error
	NewTx(ctx context.Context) (context.Context, Tx, error)

	Migrate(ctx context.Context) error
	FindMigrationVersion(ctx context.Context) (int, error)

	FindOrganizations(ctx context.Context) ([]*Organization, error)
	FindOrganizationByName(ctx context.Context, name string) (*Organization, error)
	FindOrganizationByID(ctx context.Context, id string) (*Organization, error)
	InsertOrganization(ctx context.Context, opts *InsertOrganizationOptions) (*Organization, error)
	UpdateOrganization(ctx context.Context, id string, opts *UpdateOrganizationOptions) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error

	FindProjects(ctx context.Context, orgName string) ([]*Project, error)
	FindProjectByID(ctx context.Context, id string) (*Project, error)
	FindProjectByName(ctx context.Context, orgName string, name string) (*Project, error)
	FindProjectByGithubURL(ctx context.Context, githubURL string) (*Project, error)
	InsertProject(ctx context.Context, opts *InsertProjectOptions) (*Project, error)
	UpdateProject(ctx context.Context, id string, opts *UpdateProjectOptions) (*Project, error)
	DeleteProject(ctx context.Context, id string) error

	FindUsers(ctx context.Context) ([]*User, error)
	FindUser(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	InsertUser(ctx context.Context, opts *InsertUserOptions) (*User, error)
	UpdateUser(ctx context.Context, id string, opts *UpdateUserOptions) (*User, error)
	DeleteUser(ctx context.Context, id string) error

	FindUserAuthTokens(ctx context.Context, userID string) ([]*UserAuthToken, error)
	FindUserAuthToken(ctx context.Context, id string) (*UserAuthToken, error)
	InsertUserAuthToken(ctx context.Context, opts *InsertUserAuthTokenOptions) (*UserAuthToken, error)
	DeleteUserAuthToken(ctx context.Context, id string) error

	// InsertAuthCode inserts the authorization code data into the store.
	InsertAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*AuthCode, error)
	// FindAuthCodeByDeviceCode retrieves the authorization code data from the store
	FindAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*AuthCode, error)
	// FindAuthCodeByUserCode retrieves the authorization code data from the store
	FindAuthCodeByUserCode(ctx context.Context, userCode string) (*AuthCode, error)
	// UpdateAuthCode updates the authorization code data in the store
	UpdateAuthCode(ctx context.Context, userCode, userID string, approvalState AuthCodeState) error
	// DeleteAuthCode deletes the authorization code data from the store
	DeleteAuthCode(ctx context.Context, deviceCode string) error

	FindDeployments(ctx context.Context, projectID string) ([]*Deployment, error)
	FindDeployment(ctx context.Context, id string) (*Deployment, error)
	InsertDeployment(ctx context.Context, opts *InsertDeploymentOptions) (*Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id string, status DeploymentStatus, logs string) (*Deployment, error)
	DeleteDeployment(ctx context.Context, id string) error

	QueryRuntimeSlotsUsed(ctx context.Context) ([]*RuntimeSlotsUsed, error)

	FindOrganizationMemberUsers(ctx context.Context, orgID string) ([]*Member, error)
	FindOrganizationMemberUsersByRole(ctx context.Context, orgID, roleID string) ([]*User, error)
	InsertOrganizationMemberUser(ctx context.Context, orgID, userID, roleID string) error
	DeleteOrganizationMemberUser(ctx context.Context, orgID, userID string) error
	UpdateOrganizationMemberUserRole(ctx context.Context, orgID, userID, roleID string) error

	FindProjectMemberUsers(ctx context.Context, projectID string) ([]*Member, error)
	InsertProjectMemberUser(ctx context.Context, projectID, userID, roleID string) error
	DeleteProjectMemberUser(ctx context.Context, projectID, userID string) error
	UpdateProjectMemberUserRole(ctx context.Context, projectID, userID, roleID string) error

	FindOrganizationRole(ctx context.Context, name string) (*OrganizationRole, error)
	FindProjectRole(ctx context.Context, name string) (*ProjectRole, error)

	// ResolveOrganizationMemberUserRoles resolves the direct and group roles of a user in an organization
	ResolveOrganizationMemberUserRoles(ctx context.Context, userID, orgID string) ([]*OrganizationRole, error)
	// ResolveProjectMemberUserRoles resolves the direct and group roles of a user in a project
	ResolveProjectMemberUserRoles(ctx context.Context, userID, projectID string) ([]*ProjectRole, error)

	InsertUsergroup(ctx context.Context, opts *InsertUsergroupOptions) (*Usergroup, error)
	UpdateOrganizationAllUsergroup(ctx context.Context, orgID, groupID string) (*Organization, error)
	InsertUserInUsergroup(ctx context.Context, userID, groupID string) error
	DeleteUserFromUsergroup(ctx context.Context, userID, groupID string) error
	InsertProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error
	FindUsersUsergroups(ctx context.Context, userID, orgID string) ([]*Usergroup, error)

	FindOrganizationsForUser(ctx context.Context, userID string) ([]*Organization, error)
	FindProjectsForUser(ctx context.Context, userID string) ([]*Project, error)
	FindProjectsForOrganization(ctx context.Context, orgID string) ([]*Project, error)

	FindProjectsForProjectMemberUser(ctx context.Context, orgID, userID string) ([]*Project, error)
	FindPublicProjectsInOrganization(ctx context.Context, orgID string) ([]*Project, error)

	CheckOrganizationProjectsHasMemberUser(ctx context.Context, orgID, userID string) (bool, error)
	CheckOrganizationHasPublicProjects(ctx context.Context, orgID string) (bool, error)

	InsertOrganizationMemberUserInvitation(ctx context.Context, email, invitedByID, orgID, roleID string) error
	InsertProjectMemberUserInvitation(ctx context.Context, email, invitedByID, projectID, roleID string) error
	FindOrganizationMemberInvitations(ctx context.Context, orgID string) ([]*UserInvite, error)
	FindOrganizationMemberUserInvitations(ctx context.Context, userEmail string) ([]*OrganizationMemberUserInvitation, error)
	FindOrganizationMemberUserInvitation(ctx context.Context, orgID, userEmail string) (*OrganizationMemberUserInvitation, error)
	FindProjectMemberInvitations(ctx context.Context, projectID string) ([]*UserInvite, error)
	FindProjectMemberUserInvitations(ctx context.Context, userEmail string) ([]*ProjectMemberUserInvitation, error)
	FindProjectMemberUserInvitation(ctx context.Context, projectID, userEmail string) (*ProjectMemberUserInvitation, error)
	DeleteOrganizationMemberUserInvitation(ctx context.Context, id string) error
	DeleteProjectMemberUserInvitation(ctx context.Context, id string) error
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

// ErrNotFound is returned for single row queries that return no values.
var ErrNotFound = errors.New("database: not found")

// ErrNotUnique is returned when a unique constraint is violated
var ErrNotUnique = errors.New("database: violates unique constraint")

// Entity is an enum representing the entities in this package.
type Entity string

const (
	EntityOrganization  Entity = "Organization"
	EntityProject       Entity = "Project"
	EntityUser          Entity = "User"
	EntityUserAuthToken Entity = "UserAuthToken"
	EntityClient        Entity = "Client"
)

// Organization represents a tenant.
type Organization struct {
	ID             string
	Name           string
	Description    string
	CreatedOn      time.Time `db:"created_on"`
	UpdatedOn      time.Time `db:"updated_on"`
	AllUsergroupID *string   `db:"all_usergroup_id"`
}

// InsertOrganizationOptions defines options for inserting a new org
type InsertOrganizationOptions struct {
	Name        string `validate:"slug"`
	Description string
}

// UpdateOrganizationOptions defines options for updating an existing org
type UpdateOrganizationOptions struct {
	Name        string `validate:"slug"`
	Description string
}

// Project represents one Git connection.
// Projects belong to an organization.
type Project struct {
	ID                     string
	OrganizationID         string `db:"org_id"`
	Name                   string
	Description            string
	Public                 bool
	Region                 string
	ProductionSlots        int       `db:"production_slots"`
	ProductionOLAPDriver   string    `db:"production_olap_driver"`
	ProductionOLAPDSN      string    `db:"production_olap_dsn"`
	ProductionBranch       string    `db:"production_branch"`
	ProductionVariables    Variables `db:"production_variables"`
	GithubURL              *string   `db:"github_url"`
	GithubInstallationID   *int64    `db:"github_installation_id"`
	ProductionDeploymentID *string   `db:"production_deployment_id"`
	CreatedOn              time.Time `db:"created_on"`
	UpdatedOn              time.Time `db:"updated_on"`
}

// Variables implements JSON SQL encoding of variables in Project.
type Variables map[string]string

func (e *Variables) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed type assertion to []byte")
	}
	return json.Unmarshal(b, &e)
}

// InsertProjectOptions defines options for inserting a new Project.
type InsertProjectOptions struct {
	OrganizationID       string `validate:"required"`
	Name                 string `validate:"slug"`
	UserID               string
	Description          string
	Public               bool
	Region               string
	ProductionOLAPDriver string
	ProductionOLAPDSN    string
	ProductionSlots      int
	ProductionBranch     string
	GithubURL            *string `validate:"omitempty,http_url"`
	GithubInstallationID *int64  `validate:"omitempty,ne=0"`
	ProductionVariables  map[string]string
}

// UpdateProjectOptions defines options for updating a Project.
type UpdateProjectOptions struct {
	Name                   string
	Description            string
	Public                 bool
	ProductionBranch       string
	ProductionVariables    map[string]string
	GithubURL              *string `validate:"omitempty,http_url"`
	GithubInstallationID   *int64  `validate:"omitempty,ne=0"`
	ProductionDeploymentID *string
}

// User is a person registered in Rill.
// Users may belong to multiple organizations and projects.
type User struct {
	ID             string
	Email          string
	DisplayName    string    `db:"display_name"`
	PhotoURL       string    `db:"photo_url"`
	GithubUsername string    `db:"github_username"`
	CreatedOn      time.Time `db:"created_on"`
	UpdatedOn      time.Time `db:"updated_on"`
}

// InsertUserOptions defines options for inserting a new user
type InsertUserOptions struct {
	Email       string `validate:"email"`
	DisplayName string
	PhotoURL    string
}

// UpdateUserOptions defines options for updating an existing user
type UpdateUserOptions struct {
	DisplayName    string
	PhotoURL       string
	GithubUsername string
}

// InsertUsergroupOptions defines options for inserting a new usergroup
type InsertUsergroupOptions struct {
	OrgID string
	Name  string `validate:"slug"`
}

// UserAuthToken is a persistent API token for a user.
type UserAuthToken struct {
	ID           string
	SecretHash   []byte    `db:"secret_hash"`
	UserID       string    `db:"user_id"`
	DisplayName  string    `db:"display_name"`
	AuthClientID *string   `db:"auth_client_id"`
	CreatedOn    time.Time `db:"created_on"`
}

// InsertUserAuthTokenOptions defines options for creating a UserAuthToken.
type InsertUserAuthTokenOptions struct {
	ID           string
	SecretHash   []byte
	UserID       string
	DisplayName  string
	AuthClientID *string
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
	AuthClientIDRillWeb = "12345678-0000-0000-0000-000000000001"
	AuthClientIDRillCLI = "12345678-0000-0000-0000-000000000002"
)

// AuthCodeState is an enum representing the approval state of an AuthCode
type AuthCodeState int

const (
	AuthCodeStatePending  AuthCodeState = 0
	AuthCodeStateApproved AuthCodeState = 1
	AuthCodeStateRejected AuthCodeState = 2
)

// AuthCode represents a user authentication code as part of the OAuth2 device flow.
// They're currently used for authenticating users in the CLI.
type AuthCode struct {
	ID            string        `db:"id"`
	DeviceCode    string        `db:"device_code"`
	UserCode      string        `db:"user_code"`
	Expiry        time.Time     `db:"expires_on"`
	ApprovalState AuthCodeState `db:"approval_state"`
	ClientID      string        `db:"client_id"`
	UserID        *string       `db:"user_id"`
	CreatedOn     time.Time     `db:"created_on"`
	UpdatedOn     time.Time     `db:"updated_on"`
}

// DeploymentStatus is an enum representing the state of a deployment
type DeploymentStatus int

const (
	DeploymentStatusUnspecified DeploymentStatus = 0
	DeploymentStatusPending     DeploymentStatus = 1
	DeploymentStatusOK          DeploymentStatus = 2
	DeploymentStatusReconciling DeploymentStatus = 3
	DeploymentStatusError       DeploymentStatus = 4
)

// Deployment is a single deployment of a git branch.
// Deployments belong to a project.
type Deployment struct {
	ID                string           `db:"id"`
	ProjectID         string           `db:"project_id"`
	Slots             int              `db:"slots"`
	Branch            string           `db:"branch"`
	RuntimeHost       string           `db:"runtime_host"`
	RuntimeInstanceID string           `db:"runtime_instance_id"`
	RuntimeAudience   string           `db:"runtime_audience"`
	Status            DeploymentStatus `db:"status"`
	Logs              string           `db:"logs"`
	CreatedOn         time.Time        `db:"created_on"`
	UpdatedOn         time.Time        `db:"updated_on"`
}

// InsertDeploymentOptions defines options for inserting a new Deployment.
type InsertDeploymentOptions struct {
	ProjectID         string
	Slots             int
	Branch            string `validate:"required"`
	RuntimeHost       string `validate:"required"`
	RuntimeInstanceID string `validate:"required"`
	RuntimeAudience   string
	Status            DeploymentStatus
	Logs              string
}

// RuntimeSlotsUsed is the result of a QueryRuntimeSlotsUsed query.
type RuntimeSlotsUsed struct {
	RuntimeHost string `db:"runtime_host"`
	SlotsUsed   int    `db:"slots_used"`
}

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

type ProjectRole struct {
	ID                   string
	Name                 string
	ReadProject          bool `db:"read_project"`
	ManageProject        bool `db:"manage_project"`
	ReadProd             bool `db:"read_prod"`
	ReadProdStatus       bool `db:"read_prod_status"`
	ManageProd           bool `db:"manage_prod"`
	ReadDev              bool `db:"read_dev"`
	ReadDevStatus        bool `db:"read_dev_status"`
	ManageDev            bool `db:"manage_dev"`
	ReadProjectMembers   bool `db:"read_project_members"`
	ManageProjectMembers bool `db:"manage_project_members"`
}

const (
	OrganizationAdminRoleName        = "admin"
	OrganizationCollaboratorRoleName = "collaborator"
	OrganizationViewerRoleName       = "viewer"
	ProjectAdminRoleName             = "admin"
	ProjectCollaboratorRoleName      = "collaborator"
	ProjectViewerRoleName            = "viewer"
)

type Usergroup struct {
	ID    string `db:"id"`
	OrgID string `db:"org_id"`
	Name  string `db:"name"`
}

type Member struct {
	ID          string
	Email       string
	DisplayName string    `db:"display_name"`
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
	RoleName    string    `db:"name"`
}

type OrganizationMemberUserInvitation struct {
	ID              string
	Email           string
	InvitedByUserID string    `db:"invited_by_user_id"`
	OrgID           string    `db:"org_id"`
	OrgRoleID       string    `db:"org_role_id"`
	CreatedOn       time.Time `db:"created_on"`
}

type ProjectMemberUserInvitation struct {
	ID              string
	Email           string
	InvitedByUserID string    `db:"invited_by_user_id"`
	ProjectID       string    `db:"project_id"`
	ProjectRoleID   string    `db:"project_role_id"`
	CreatedOn       time.Time `db:"created_on"`
}

type UserInvite struct {
	Email     string
	Role      string
	InvitedBy string `db:"invited_by"`
}
