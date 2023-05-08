package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/admin/database"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	database.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (database.DB, error) {
	db, err := otelsql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(db, "pgx")
	return &connection{db: dbx}, nil
}

type connection struct {
	db *sqlx.DB
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) FindOrganizations(ctx context.Context) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM orgs ORDER BY lower(name)")
	if err != nil {
		return nil, parseErr("orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationsForUser(ctx context.Context, userID string) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT u.* FROM (SELECT o.* FROM orgs o JOIN users_orgs_roles uor ON o.id = uor.org_id
		WHERE uor.user_id = $1
		UNION
		SELECT o.* FROM orgs o JOIN usergroups_orgs_roles ugor ON o.id = ugor.org_id
		JOIN usergroups_users uug ON ugor.usergroup_id = uug.usergroup_id
		WHERE uug.user_id = $1
		UNION
		SELECT o.* FROM orgs o JOIN projects p ON o.id = p.org_id
		JOIN users_projects_roles upr ON p.id = upr.project_id
		WHERE upr.user_id = $1) u order by u.name
	`, userID)
	if err != nil {
		return nil, parseErr("orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganization(ctx context.Context, orgID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs WHERE id = $1", orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationByName(ctx context.Context, name string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs WHERE lower(name)=lower($1)", name).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) CheckOrganizationHasOutsideUser(ctx context.Context, orgID, userID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM projects p JOIN users_projects_roles upr ON p.id = upr.project_id WHERE p.org_id = $1 AND upr.user_id = $2 limit 1)", orgID, userID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) CheckOrganizationHasPublicProjects(ctx context.Context, orgID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM projects p WHERE p.org_id = $1 AND p.public = true limit 1)", orgID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) InsertOrganization(ctx context.Context, opts *database.InsertOrganizationOptions) (*database.Organization, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO orgs(name, description, quota_projects, quota_deployments, quota_slots_total, quota_slots_per_deployment, quota_outstanding_invites) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`,
		opts.Name, opts.Description, opts.QuotaProjects, opts.QuotaDeployments, opts.QuotaSlotsTotal, opts.QuotaSlotsPerDeployment, opts.QuotaOutstandingInvites).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) DeleteOrganization(ctx context.Context, name string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM orgs WHERE lower(name)=lower($1)", name)
	return checkDeleteRow("org", res, err)
}

func (c *connection) UpdateOrganization(ctx context.Context, id string, opts *database.UpdateOrganizationOptions) (*database.Organization, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE orgs SET name=$1, description=$2, updated_on=now() WHERE id=$3 RETURNING *", opts.Name, opts.Description, id).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) UpdateOrganizationAllUsergroup(ctx context.Context, orgID, groupID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `UPDATE orgs SET all_usergroup_id = $1 WHERE id = $2 RETURNING *`, groupID, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("org", err)
	}
	return res, nil
}

func (c *connection) FindProjects(ctx context.Context, orgName string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p JOIN orgs o ON p.org_id = o.id WHERE lower(o.name)=lower($1) ORDER BY lower(p.name)", orgName)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectsForUser(ctx context.Context, userID string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p JOIN users_projects_roles upr ON p.id = upr.project_id
		WHERE upr.user_id = $1
		UNION
		SELECT p.* FROM projects p JOIN usergroups_projects_roles upgr ON p.id = upgr.project_id
		JOIN usergroups_users uug ON upgr.usergroup_id = uug.usergroup_id
		WHERE uug.user_id = $1
	`, userID)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectsForOrganization(ctx context.Context, orgID string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE p.org_id=$1 ORDER BY lower(p.name)", orgID)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectsForOrgAndUser(ctx context.Context, orgID, userID string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p
		WHERE p.org_id = $1 AND p.id IN (
			SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = $2
			UNION
			SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users uug ON ugpr.usergroup_id = uug.usergroup_id WHERE uug.user_id = $2
		)
	`, orgID, userID)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindPublicProjectsInOrganization(ctx context.Context, orgID string) ([]*database.Project, error) {
	var res []*database.Project
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE p.org_id = $1 AND p.public = true", orgID)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectsByGithubURL(ctx context.Context, githubURL string) ([]*database.Project, error) {
	result := make([]*database.Project, 0)
	err := c.getDB(ctx).SelectContext(ctx, &result, "SELECT p.* FROM projects p WHERE lower(p.github_url)=lower($1) ", githubURL)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return result, nil
}

func (c *connection) FindProjectsByOrgAndGithubURL(ctx context.Context, orgID, githubURL string) ([]*database.Project, error) {
	result := make([]*database.Project, 0)
	err := c.getDB(ctx).SelectContext(ctx, &result, "SELECT p.* FROM projects p WHERE lower(p.github_url)=lower($1) AND org_id=$2", githubURL, orgID)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return result, nil
}

func (c *connection) FindProject(ctx context.Context, id string) (*database.Project, error) {
	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM projects WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return res, nil
}

func (c *connection) FindProjectByName(ctx context.Context, orgName, name string) (*database.Project, error) {
	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT p.* FROM projects p JOIN orgs o ON p.org_id = o.id WHERE lower(p.name)=lower($1) AND lower(o.name)=lower($2)", name, orgName).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return res, nil
}

func (c *connection) InsertProject(ctx context.Context, opts *database.InsertProjectOptions) (*database.Project, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO projects (org_id, name, description, public, region, prod_olap_driver, prod_olap_dsn, prod_slots, prod_branch, prod_variables, github_url, github_installation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *`,
		opts.OrganizationID, opts.Name, opts.Description, opts.Public, opts.Region, opts.ProdOLAPDriver, opts.ProdOLAPDSN, opts.ProdSlots, opts.ProdBranch, database.Variables(opts.ProdVariables), opts.GithubURL, opts.GithubInstallationID,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return res, nil
}

func (c *connection) DeleteProject(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM projects WHERE id=$1", id)
	return checkDeleteRow("project", res, err)
}

func (c *connection) UpdateProject(ctx context.Context, id string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Project{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE projects SET name=$1, description=$2, public=$3, prod_branch=$4, prod_variables=$5, github_url=$6, github_installation_id=$7, prod_deployment_id=$8, updated_on=now()
		WHERE id=$9 RETURNING *`,
		opts.Name, opts.Description, opts.Public, opts.ProdBranch, database.Variables(opts.ProdVariables), opts.GithubURL, opts.GithubInstallationID, opts.ProdDeploymentID, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return res, nil
}

func (c *connection) CountProjectsForOrganization(ctx context.Context, orgID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT COUNT(*) FROM projects WHERE org_id = $1", orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("project count", err)
	}
	return count, nil
}

func (c *connection) FindDeployments(ctx context.Context, projectID string) ([]*database.Deployment, error) {
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM deployments d WHERE d.project_id=$1", projectID)
	if err != nil {
		return nil, parseErr("deployments", err)
	}
	return res, nil
}

func (c *connection) FindDeployment(ctx context.Context, id string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT d.* FROM deployments d WHERE d.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) InsertDeployment(ctx context.Context, opts *database.InsertDeploymentOptions) (*database.Deployment, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO deployments (project_id, slots, branch, runtime_host, runtime_instance_id, runtime_audience, status, logs)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`,
		opts.ProjectID, opts.Slots, opts.Branch, opts.RuntimeHost, opts.RuntimeInstanceID, opts.RuntimeAudience, opts.Status, opts.Logs,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) DeleteDeployment(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM deployments WHERE id=$1", id)
	return checkDeleteRow("deployment", res, err)
}

func (c *connection) UpdateDeploymentStatus(ctx context.Context, id string, status database.DeploymentStatus, logs string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET status=$1, logs=$2, updated_on=now() WHERE id=$3 RETURNING *", status, logs, id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentTS(ctx context.Context, ids []string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET updated_on=now() WHERE id = any($1) RETURNING *", ids).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentBranch(ctx context.Context, id, branch string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET branch=$1, updated_on=now() WHERE id=$2 RETURNING *", branch, id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) CountDeploymentsForOrganization(ctx context.Context, orgID string) (*database.DeploymentsCount, error) {
	res := &database.DeploymentsCount{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT COUNT(*) as deployments, COALESCE(SUM(slots), 0) as slots FROM deployments WHERE project_id IN (SELECT id FROM projects WHERE org_id = $1)`, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("deployments count", err)
	}
	return res, nil
}

func (c *connection) ResolveRuntimeSlotsUsed(ctx context.Context) ([]*database.RuntimeSlotsUsed, error) {
	var res []*database.RuntimeSlotsUsed
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT d.runtime_host, SUM(d.slots) AS slots_used FROM deployments d GROUP BY d.runtime_host")
	if err != nil {
		return nil, parseErr("slots used", err)
	}
	return res, nil
}

func (c *connection) FindUsers(ctx context.Context) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT u.* FROM users u")
	if err != nil {
		return nil, parseErr("users", err)
	}
	return res, nil
}

func (c *connection) FindUser(ctx context.Context, id string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE u.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) FindUserByEmail(ctx context.Context, email string) (*database.User, error) {
	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT u.* FROM users u WHERE lower(u.email)=lower($1)", email).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) InsertUser(ctx context.Context, opts *database.InsertUserOptions) (*database.User, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO users (email, display_name, photo_url, quota_singleuser_orgs) VALUES ($1, $2, $3, $4) RETURNING *", opts.Email, opts.DisplayName, opts.PhotoURL, opts.QuotaSingleuserOrgs).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) DeleteUser(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	return checkDeleteRow("user", res, err)
}

func (c *connection) UpdateUser(ctx context.Context, id string, opts *database.UpdateUserOptions) (*database.User, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE users SET display_name=$2, photo_url=$3, github_username=$4, updated_on=now() WHERE id=$1 RETURNING *",
		id,
		opts.DisplayName,
		opts.PhotoURL,
		opts.GithubUsername).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) InsertUsergroup(ctx context.Context, opts *database.InsertUsergroupOptions) (*database.Usergroup, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO usergroups (org_id, name) VALUES ($1, $2) RETURNING *
	`, opts.OrgID, opts.Name).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) InsertUsergroupMember(ctx context.Context, groupID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO usergroups_users (user_id, usergroup_id) VALUES ($1, $2)", userID, groupID)
	if err != nil {
		return parseErr("usergroup member", err)
	}
	return nil
}

func (c *connection) DeleteUsergroupMember(ctx context.Context, groupID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_users WHERE user_id = $1 AND usergroup_id = $2", userID, groupID)
	return checkDeleteRow("usergroup member", res, err)
}

func (c *connection) FindUserAuthTokens(ctx context.Context, userID string) ([]*database.UserAuthToken, error) {
	var res []*database.UserAuthToken
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT t.* FROM user_auth_tokens t WHERE t.user_id=$1", userID)
	if err != nil {
		return nil, parseErr("auth tokens", err)
	}
	return res, nil
}

func (c *connection) FindUserAuthToken(ctx context.Context, id string) (*database.UserAuthToken, error) {
	res := &database.UserAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM user_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("auth token", err)
	}
	return res, nil
}

func (c *connection) InsertUserAuthToken(ctx context.Context, opts *database.InsertUserAuthTokenOptions) (*database.UserAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.UserAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO user_auth_tokens (id, secret_hash, user_id, display_name, auth_client_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		opts.ID, opts.SecretHash, opts.UserID, opts.DisplayName, opts.AuthClientID,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("auth token", err)
	}
	return res, nil
}

func (c *connection) DeleteUserAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("auth token", res, err)
}

func (c *connection) FindDeviceAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*database.DeviceAuthCode, error) {
	authCode := &database.DeviceAuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM device_auth_codes WHERE device_code = $1", deviceCode).StructScan(authCode)
	if err != nil {
		return nil, parseErr("device auth code", err)
	}
	return authCode, nil
}

func (c *connection) FindPendingDeviceAuthCodeByUserCode(ctx context.Context, userCode string) (*database.DeviceAuthCode, error) {
	authCode := &database.DeviceAuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM device_auth_codes WHERE user_code = $1 AND expires_on > now() AND approval_state = 0", userCode).StructScan(authCode)
	if err != nil {
		return nil, parseErr("device auth code", err)
	}
	return authCode, nil
}

func (c *connection) InsertDeviceAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*database.DeviceAuthCode, error) {
	res := &database.DeviceAuthCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`INSERT INTO device_auth_codes (device_code, user_code, expires_on, approval_state, client_id)
		VALUES ($1, $2, $3, $4, $5)  RETURNING *`, deviceCode, userCode, expiresOn, database.DeviceAuthCodeStatePending, clientID).StructScan(res)
	if err != nil {
		return nil, parseErr("device auth code", err)
	}
	return res, nil
}

func (c *connection) DeleteDeviceAuthCode(ctx context.Context, deviceCode string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM device_auth_codes WHERE device_code=$1", deviceCode)
	return checkDeleteRow("device auth code", res, err)
}

func (c *connection) UpdateDeviceAuthCode(ctx context.Context, id, userID string, approvalState database.DeviceAuthCodeState) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "UPDATE device_auth_codes SET approval_state=$1, user_id=$2, updated_on=now() WHERE id=$3", approvalState, userID, id)
	return checkUpdateRow("device auth code", res, err)
}

func (c *connection) FindOrganizationRole(ctx context.Context, name string) (*database.OrganizationRole, error) {
	role := &database.OrganizationRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM org_roles WHERE lower(name)=lower($1)", name).StructScan(role)
	if err != nil {
		return nil, parseErr("org role", err)
	}
	return role, nil
}

func (c *connection) FindProjectRole(ctx context.Context, name string) (*database.ProjectRole, error) {
	role := &database.ProjectRole{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_roles WHERE lower(name)=lower($1)", name).StructScan(role)
	if err != nil {
		return nil, parseErr("project role", err)
	}
	return role, nil
}

func (c *connection) ResolveOrganizationRolesForUser(ctx context.Context, userID, orgID string) ([]*database.OrganizationRole, error) {
	var res []*database.OrganizationRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT r.* FROM users_orgs_roles uor
		JOIN org_roles r ON uor.org_role_id = r.id
		WHERE uor.user_id = $1 AND uor.org_id = $2
		UNION
		SELECT * FROM org_roles WHERE id IN (
			SELECT org_role_id FROM usergroups_orgs_roles uor JOIN usergroups_users uug 
			ON uor.usergroup_id = uug.usergroup_id WHERE uug.user_id = $1 AND uor.org_id = $2
		)`, userID, orgID)
	if err != nil {
		return nil, parseErr("org roles", err)
	}
	return res, nil
}

func (c *connection) ResolveProjectRolesForUser(ctx context.Context, userID, projectID string) ([]*database.ProjectRole, error) {
	var res []*database.ProjectRole
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT r.* FROM users_projects_roles upr
		JOIN project_roles r ON upr.project_role_id = r.id
		WHERE upr.user_id = $1 AND upr.project_id = $2
		UNION
		SELECT * FROM project_roles WHERE id IN (
			SELECT project_role_id FROM usergroups_projects_roles upr JOIN usergroups_users uug 
			ON upr.usergroup_id = uug.usergroup_id WHERE uug.user_id = $1 AND upr.project_id = $2
		)`, userID, projectID)
	if err != nil {
		return nil, parseErr("project roles", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationMemberUsers(ctx context.Context, orgID string) ([]*database.Member, error) {
	var res []*database.Member
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT u.id, u.email, u.display_name, u.created_on, u.updated_on, r.name FROM users u 
    	JOIN users_orgs_roles uor ON u.id = uor.user_id
		JOIN org_roles r ON r.id = uor.org_role_id WHERE uor.org_id=$1`, orgID)
	if err != nil {
		return nil, parseErr("org members", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationMemberUsersByRole(ctx context.Context, orgID, roleID string) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(
		ctx, &res, "SELECT u.* FROM users u JOIN users_orgs_roles uor on u.id = uor.user_id WHERE uor.org_id=$1 AND uor.org_role_id=$2", orgID, roleID)
	if err != nil {
		return nil, parseErr("org members", err)
	}
	return res, nil
}

func (c *connection) InsertOrganizationMemberUser(ctx context.Context, orgID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_orgs_roles (user_id, org_id, org_role_id) VALUES ($1, $2, $3)", userID, orgID, roleID)
	if err != nil {
		return parseErr("org member", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no rows affected when adding user to organization")
	}
	return nil
}

func (c *connection) DeleteOrganizationMemberUser(ctx context.Context, orgID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_orgs_roles WHERE user_id = $1 AND org_id = $2", userID, orgID)
	return checkDeleteRow("org member", res, err)
}

func (c *connection) UpdateOrganizationMemberUserRole(ctx context.Context, orgID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users_orgs_roles SET org_role_id = $1 WHERE user_id = $2 AND org_id = $3`, roleID, userID, orgID)
	return checkUpdateRow("org member", res, err)
}

func (c *connection) CountSingleuserOrganizationsForMemberUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT COALESCE(SUM(total_count), 0) as total_count FROM (
			SELECT CASE WHEN COUNT(*) = 1 THEN 1 ELSE 0 END as total_count FROM users_orgs_roles WHERE org_id IN (
				SELECT org_id FROM users_orgs_roles WHERE user_id = $1
			) GROUP BY org_id
		) as subquery
	`, userID).Scan(&count)
	if err != nil {
		return 0, parseErr("singleuser orgs count", err)
	}
	return count, nil
}

func (c *connection) FindProjectMemberUsers(ctx context.Context, projectID string) ([]*database.Member, error) {
	var res []*database.Member
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT u.id, u.email, u.display_name, u.created_on, u.updated_on, r.name FROM users u 
    	JOIN users_projects_roles upr ON u.id = upr.user_id
		JOIN project_roles r ON r.id = upr.project_role_id WHERE upr.project_id=$1`, projectID)
	if err != nil {
		return nil, parseErr("project members", err)
	}
	return res, nil
}

func (c *connection) InsertProjectMemberUser(ctx context.Context, projectID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO users_projects_roles (user_id, project_id, project_role_id) VALUES ($1, $2, $3)", userID, projectID, roleID)
	if err != nil {
		return parseErr("project member", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no rows affected when adding user to project")
	}
	return nil
}

func (c *connection) InsertProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO usergroups_projects_roles (usergroup_id, project_id, project_role_id) VALUES ($1, $2, $3)", groupID, projectID, roleID)
	if err != nil {
		return parseErr("project group member", err)
	}
	return nil
}

func (c *connection) DeleteProjectMemberUser(ctx context.Context, projectID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_projects_roles WHERE user_id = $1 AND project_id = $2", userID, projectID)
	return checkDeleteRow("project member", res, err)
}

func (c *connection) DeleteAllProjectMemberUserForOrganization(ctx context.Context, orgID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM users_projects_roles upr WHERE upr.user_id = $1 AND upr.project_id IN (SELECT p.id FROM projects p WHERE p.org_id = $2)", userID, orgID)
	if err != nil {
		return parseErr("project member", err)
	}
	return nil
}

func (c *connection) UpdateProjectMemberUserRole(ctx context.Context, projectID, userID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users_projects_roles SET project_role_id = $1 WHERE user_id = $2 AND project_id = $3`, roleID, userID, projectID)
	return checkUpdateRow("project member", res, err)
}

func (c *connection) FindOrganizationInvites(ctx context.Context, orgID string) ([]*database.Invite, error) {
	var res []*database.Invite
	err := c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT uoi.email, ur.name as role, u.email as invited_by 
			FROM org_invites uoi JOIN org_roles ur ON uoi.org_role_id = ur.id JOIN users u ON uoi.invited_by_user_id = u.id WHERE uoi.org_id = $1`, orgID)
	if err != nil {
		return nil, parseErr("org invites", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationInvitesByEmail(ctx context.Context, userEmail string) ([]*database.OrganizationInvite, error) {
	var res []*database.OrganizationInvite
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM org_invites WHERE lower(email) = lower($1)", userEmail)
	if err != nil {
		return nil, parseErr("org invites", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationInvite(ctx context.Context, orgID, userEmail string) (*database.OrganizationInvite, error) {
	res := &database.OrganizationInvite{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM org_invites WHERE lower(email) = lower($1) AND org_id = $2", userEmail, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("org invite", err)
	}
	return res, nil
}

func (c *connection) InsertOrganizationInvite(ctx context.Context, email, orgID, roleID, invitedByID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO org_invites (email, invited_by_user_id, org_id, org_role_id) VALUES ($1, $2, $3, $4)", email, invitedByID, orgID, roleID)
	if err != nil {
		return parseErr("org invite", err)
	}
	return nil
}

func (c *connection) DeleteOrganizationInvite(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM org_invites WHERE id = $1", id)
	return checkDeleteRow("org invite", res, err)
}

func (c *connection) CountInvitesForOrganization(ctx context.Context, orgID string) (int, error) {
	var count int
	// count outstanding org invites as well as project invites for this org
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT COALESCE(SUM(total_count), 0) as total_count FROM (
  			SELECT COUNT(*) as total_count FROM org_invites WHERE org_id = $1
  			UNION ALL
  			SELECT COUNT(*) as total_count FROM project_invites WHERE project_id IN (SELECT id FROM projects WHERE org_id = $1)
		) as subquery
		`, orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("invites count", err)
	}
	return count, nil
}

func (c *connection) UpdateOrganizationInviteRole(ctx context.Context, id, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE org_invites SET org_role_id = $1 WHERE id = $2`, roleID, id)
	return checkUpdateRow("org invite", res, err)
}

func (c *connection) FindProjectInvites(ctx context.Context, projectID string) ([]*database.Invite, error) {
	var res []*database.Invite
	err := c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT upi.email, ur.name as role, u.email as invited_by 
			FROM project_invites upi JOIN project_roles ur ON upi.project_role_id = ur.id JOIN users u ON upi.invited_by_user_id = u.id WHERE upi.project_id = $1`, projectID)
	if err != nil {
		return nil, parseErr("project invites", err)
	}
	return res, nil
}

func (c *connection) FindProjectInvitesByEmail(ctx context.Context, userEmail string) ([]*database.ProjectInvite, error) {
	var res []*database.ProjectInvite
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM project_invites WHERE lower(email) = lower($1)", userEmail)
	if err != nil {
		return nil, parseErr("project invites", err)
	}
	return res, nil
}

func (c *connection) FindProjectInvite(ctx context.Context, projectID, userEmail string) (*database.ProjectInvite, error) {
	res := &database.ProjectInvite{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_invites WHERE lower(email) = lower($1) AND project_id = $2", userEmail, projectID).StructScan(res)
	if err != nil {
		return nil, parseErr("project invite", err)
	}
	return res, nil
}

func (c *connection) InsertProjectInvite(ctx context.Context, email, projectID, roleID, invitedByID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO project_invites (email, invited_by_user_id, project_id, project_role_id) VALUES ($1, $2, $3, $4)", email, invitedByID, projectID, roleID)
	if err != nil {
		return parseErr("project invite", err)
	}
	return nil
}

func (c *connection) DeleteProjectInvite(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM project_invites WHERE id = $1", id)
	return checkDeleteRow("project invite", res, err)
}

func (c *connection) UpdateProjectInviteRole(ctx context.Context, id, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE project_invites SET project_role_id = $1 WHERE id = $2`, roleID, id)
	return checkUpdateRow("project invite", res, err)
}

func checkUpdateRow(target string, res sql.Result, err error) error {
	if err != nil {
		return parseErr(target, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return parseErr(target, err)
	}
	if n == 0 {
		return parseErr(target, sql.ErrNoRows)
	}
	if n > 1 {
		// This should never happen
		panic(fmt.Errorf("expected to update 1 row, but updated %d", n))
	}
	return nil
}

func checkDeleteRow(target string, res sql.Result, err error) error {
	if err != nil {
		return parseErr(target, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return parseErr(target, err)
	}
	if n == 0 {
		return parseErr(target, sql.ErrNoRows)
	}
	if n > 1 {
		// This should never happen
		panic(fmt.Errorf("expected to delete 1 row, but deleted %d", n))
	}
	return nil
}

func parseErr(target string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		if target == "" {
			return database.ErrNotFound
		}
		return &wrappedError{
			msg: fmt.Sprintf("%s not found", target),
			// wrap database.ErrNotFound so checks with errors.Is(...) still work
			err: database.ErrNotFound,
		}
	}
	var pgerr *pgconn.PgError
	if !errors.As(err, &pgerr) {
		return err
	}
	if pgerr.Code == "23505" { // unique_violation
		switch pgerr.ConstraintName {
		case "orgs_name_idx":
			return newAlreadyExistsErr("an org with that name already exists")
		case "projects_name_idx":
			return newAlreadyExistsErr("a project with that name already exists in the org")
		case "users_email_idx":
			return newAlreadyExistsErr("a user with that email already exists")
		case "usergroups_name_idx":
			return newAlreadyExistsErr("a usergroup with that name already exists in the org")
		case "usergroups_users_pkey":
			return newAlreadyExistsErr("user is already a member of the usergroup")
		case "users_orgs_roles_pkey":
			return newAlreadyExistsErr("user is already a member of the org")
		case "users_projects_roles_pkey":
			return newAlreadyExistsErr("user is already a member of the project")
		case "usergroups_orgs_roles_pkey":
			return newAlreadyExistsErr("usergroup is already a member of the org")
		case "usergroups_projects_roles_pkey":
			return newAlreadyExistsErr("usergroup is already a member of the project")
		case "org_invites_email_org_idx":
			return newAlreadyExistsErr("email has already been invited to the org")
		case "project_invites_email_project_idx":
			return newAlreadyExistsErr("email has already been invited to the project")
		default:
			if target == "" {
				return database.ErrNotUnique
			}
			return newAlreadyExistsErr(fmt.Sprintf("%s already exists", target))
		}
	}
	return err
}

func newAlreadyExistsErr(msg string) error {
	// wrap database.ErrNotUnique so checks with errors.Is(...) still work
	return &wrappedError{msg: msg, err: database.ErrNotUnique}
}

type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string {
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.err
}
