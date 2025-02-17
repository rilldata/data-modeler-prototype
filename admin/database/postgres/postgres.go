package postgres

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/admin/database"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	database.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, encKeyring []*database.EncryptionKey) (database.DB, error) {
	db, err := otelsql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(db, "pgx")
	return &connection{db: dbx, encKeyring: encKeyring}, nil
}

type connection struct {
	db         *sqlx.DB
	encKeyring []*database.EncryptionKey
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) FindOrganizations(ctx context.Context, afterName string, limit int) ([]*database.Organization, error) {
	var res []*database.Organization
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM orgs WHERE lower(name) > lower($1) ORDER BY lower(name) LIMIT $2", afterName, limit)
	if err != nil {
		return nil, parseErr("orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationsForUser(ctx context.Context, userID, afterName string, limit int) ([]*database.Organization, error) {
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
		WHERE upr.user_id = $1) u
		WHERE lower(u.name) > lower($2) ORDER BY lower(u.name) LIMIT $3
	`, userID, afterName, limit)
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

func (c *connection) CountProjectsForOrganization(ctx context.Context, orgID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT COUNT(*) FROM projects WHERE org_id=$1", orgID).Scan(&count)
	if err != nil {
		return 0, parseErr("projects", err)
	}
	return count, nil
}

func (c *connection) FindOrganizationByCustomDomain(ctx context.Context, domain string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs WHERE lower(custom_domain)=lower($1)", domain).StructScan(res)
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
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO orgs(name, display_name, description, logo_asset_id, favicon_asset_id, custom_domain, quota_projects, quota_deployments, quota_slots_total, quota_slots_per_deployment, quota_outstanding_invites, quota_storage_limit_bytes_per_deployment, billing_customer_id, payment_customer_id, billing_email, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING *`,
		opts.Name, opts.DisplayName, opts.Description, opts.LogoAssetID, opts.FaviconAssetID, opts.CustomDomain, opts.QuotaProjects, opts.QuotaDeployments, opts.QuotaSlotsTotal, opts.QuotaSlotsPerDeployment, opts.QuotaOutstandingInvites, opts.QuotaStorageLimitBytesPerDeployment, opts.BillingCustomerID, opts.PaymentCustomerID, opts.BillingEmail, opts.CreatedByUserID).StructScan(res)
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
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`UPDATE orgs SET name=$1, display_name=$2, description=$3, logo_asset_id=$4, favicon_asset_id=$5, custom_domain=$6, quota_projects=$7, quota_deployments=$8, quota_slots_total=$9, quota_slots_per_deployment=$10, quota_outstanding_invites=$11, quota_storage_limit_bytes_per_deployment=$12, billing_customer_id=$13, payment_customer_id=$14, billing_email=$15, created_by_user_id=$16, billing_plan_name=$17, billing_plan_display_name=$18, updated_on=now() WHERE id=$19 RETURNING *`,
		opts.Name, opts.DisplayName, opts.Description, opts.LogoAssetID, opts.FaviconAssetID, opts.CustomDomain, opts.QuotaProjects, opts.QuotaDeployments, opts.QuotaSlotsTotal, opts.QuotaSlotsPerDeployment, opts.QuotaOutstandingInvites, opts.QuotaStorageLimitBytesPerDeployment, opts.BillingCustomerID, opts.PaymentCustomerID, opts.BillingEmail, opts.CreatedByUserID, opts.BillingPlanName, opts.BillingPlanDisplayName, id).StructScan(res)
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

func (c *connection) FindOrganizationWhitelistedDomainForOrganizationWithJoinedRoleNames(ctx context.Context, orgID string) ([]*database.OrganizationWhitelistedDomainWithJoinedRoleNames, error) {
	var res []*database.OrganizationWhitelistedDomainWithJoinedRoleNames
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT oad.domain, r.name FROM orgs_autoinvite_domains oad JOIN org_roles r ON r.id = oad.org_role_id WHERE oad.org_id=$1", orgID)
	if err != nil {
		return nil, parseErr("org whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationWhitelistedDomainsForDomain(ctx context.Context, domain string) ([]*database.OrganizationWhitelistedDomain, error) {
	var res []*database.OrganizationWhitelistedDomain
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM orgs_autoinvite_domains WHERE lower(domain)=lower($1)", domain)
	if err != nil {
		return nil, parseErr("org whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationWhitelistedDomain(ctx context.Context, orgID, domain string) (*database.OrganizationWhitelistedDomain, error) {
	res := &database.OrganizationWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM orgs_autoinvite_domains WHERE org_id=$1 AND lower(domain)=lower($2)", orgID, domain).StructScan(res)
	if err != nil {
		return nil, parseErr("org whitelist domain", err)
	}
	return res, nil
}

func (c *connection) InsertOrganizationWhitelistedDomain(ctx context.Context, opts *database.InsertOrganizationWhitelistedDomainOptions) (*database.OrganizationWhitelistedDomain, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.OrganizationWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO orgs_autoinvite_domains(org_id, org_role_id, domain) VALUES ($1, $2, $3) RETURNING *`, opts.OrgID, opts.OrgRoleID, opts.Domain).StructScan(res)
	if err != nil {
		return nil, parseErr("org whitelist domain", err)
	}
	return res, nil
}

func (c *connection) DeleteOrganizationWhitelistedDomain(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM orgs_autoinvite_domains WHERE id=$1", id)
	return checkDeleteRow("org whitelist domain", res, err)
}

func (c *connection) FindProjects(ctx context.Context, afterName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE lower(name) > lower($1) ORDER BY lower(p.name) LIMIT $2", afterName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsByVersion(ctx context.Context, version, afterName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE p.prod_version = $1 AND lower(name) > lower($2) ORDER BY lower(p.name) LIMIT $3", version, afterName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectPathsByPattern(ctx context.Context, namePattern, afterName string, limit int) ([]string, error) {
	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT concat(o.name,'/',p.name) as project_name FROM projects p JOIN orgs o ON p.org_id = o.id
	WHERE concat(o.name,'/',p.name) ilike $1 AND concat(o.name,'/',p.name) > $2
	ORDER BY project_name
	LIMIT $3`, namePattern, afterName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectPathsByPatternAndAnnotations(ctx context.Context, namePattern, afterName string, annotationKeys []string, annotationPairs map[string]string, limit int) ([]string, error) {
	if annotationKeys == nil {
		annotationKeys = []string{}
	}
	if annotationPairs == nil {
		annotationPairs = map[string]string{}
	}

	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT concat(o.name,'/',p.name) as project_name FROM projects p JOIN orgs o ON p.org_id = o.id
	WHERE concat(o.name,'/',p.name) ilike $1 AND concat(o.name,'/',p.name) > $2 AND p.annotations ?& $3 AND p.annotations @> $4
	ORDER BY project_name
	LIMIT $5`, namePattern, afterName, annotationKeys, annotationPairs, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return res, nil
}

func (c *connection) FindProjectsForUser(ctx context.Context, userID string) ([]*database.Project, error) {
	var res []*projectDTO
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
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsForOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p
		WHERE p.org_id=$1 AND lower(p.name) > lower($2)
		ORDER BY lower(p.name) LIMIT $3
	`, orgID, afterProjectName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsForOrgAndUser(ctx context.Context, orgID, userID, afterProjectName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p
		WHERE p.org_id = $1 AND lower(p.name) > lower($2) AND (p.public = true OR p.id IN (
			SELECT upr.project_id FROM users_projects_roles upr WHERE upr.user_id = $3
			UNION
			SELECT ugpr.project_id FROM usergroups_projects_roles ugpr JOIN usergroups_users uug ON ugpr.usergroup_id = uug.usergroup_id WHERE uug.user_id = $3
		))  ORDER BY lower(p.name) LIMIT $4
	`, orgID, afterProjectName, userID, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindPublicProjectsInOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT p.* FROM projects p
		WHERE p.org_id = $1 AND p.public = true AND lower(p.name) > lower($2)
		ORDER BY lower(p.name) LIMIT $3
	`, orgID, afterProjectName, limit)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsByGithubURL(ctx context.Context, githubURL string) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE lower(p.github_url)=lower($1) ", githubURL)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProjectsByGithubInstallationID(ctx context.Context, id int64) ([]*database.Project, error) {
	var res []*projectDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT p.* FROM projects p WHERE p.github_installation_id=$1", id)
	if err != nil {
		return nil, parseErr("projects", err)
	}
	return c.projectsFromDTOs(res)
}

func (c *connection) FindProject(ctx context.Context, id string) (*database.Project, error) {
	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM projects WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) FindProjectByName(ctx context.Context, orgName, name string) (*database.Project, error) {
	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT p.* FROM projects p JOIN orgs o ON p.org_id = o.id WHERE lower(p.name)=lower($1) AND lower(o.name)=lower($2)", name, orgName).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) InsertProject(ctx context.Context, opts *database.InsertProjectOptions) (*database.Project, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO projects (org_id, name, description, public, created_by_user_id, provisioner, prod_olap_driver, prod_olap_dsn, prod_slots, subpath, prod_branch, archive_asset_id, github_url, github_installation_id, prod_ttl_seconds, prod_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING *`,
		opts.OrganizationID, opts.Name, opts.Description, opts.Public, opts.CreatedByUserID, opts.Provisioner, opts.ProdOLAPDriver, opts.ProdOLAPDSN, opts.ProdSlots, opts.Subpath, opts.ProdBranch, opts.ArchiveAssetID, opts.GithubURL, opts.GithubInstallationID, opts.ProdTTLSeconds, opts.ProdVersion,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) DeleteProject(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM projects WHERE id=$1", id)
	return checkDeleteRow("project", res, err)
}

func (c *connection) UpdateProject(ctx context.Context, id string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}
	if opts.Annotations == nil {
		opts.Annotations = make(map[string]string, 0)
	}

	res := &projectDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE projects SET name=$1, description=$2, public=$3, prod_branch=$4, github_url=$5, github_installation_id=$6, archive_asset_id=$7, prod_deployment_id=$8, provisioner=$9, prod_slots=$10, subpath=$11, prod_ttl_seconds=$12, annotations=$13, prod_version=$14, updated_on=now()
		WHERE id=$15 RETURNING *`,
		opts.Name, opts.Description, opts.Public, opts.ProdBranch, opts.GithubURL, opts.GithubInstallationID, opts.ArchiveAssetID, opts.ProdDeploymentID, opts.Provisioner, opts.ProdSlots, opts.Subpath, opts.ProdTTLSeconds, opts.Annotations, opts.ProdVersion, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("project", err)
	}
	return c.projectFromDTO(res)
}

func (c *connection) CountProjectsQuotaUsage(ctx context.Context, orgID string) (*database.ProjectsQuotaUsage, error) {
	res := &database.ProjectsQuotaUsage{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		WITH t1 AS (SELECT * FROM projects WHERE org_id = $1)
		SELECT
			(SELECT COUNT(*) FROM t1) AS projects,
			(SELECT COUNT(*) FROM deployments d WHERE d.project_id IN (SELECT id FROM t1)) AS deployments,
			(SELECT COALESCE(SUM(prod_slots), 0) FROM t1) AS slots
	`, orgID).StructScan(res)
	if err != nil {
		return nil, parseErr("projects quota usage", err)
	}
	return res, nil
}

func (c *connection) FindProjectWhitelistedDomainForProjectWithJoinedRoleNames(ctx context.Context, projectID string) ([]*database.ProjectWhitelistedDomainWithJoinedRoleNames, error) {
	var res []*database.ProjectWhitelistedDomainWithJoinedRoleNames
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT pad.domain, r.name FROM projects_autoinvite_domains pad JOIN project_roles r ON r.id = pad.project_role_id WHERE pad.project_id=$1", projectID)
	if err != nil {
		return nil, parseErr("project whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindProjectWhitelistedDomainsForDomain(ctx context.Context, domain string) ([]*database.ProjectWhitelistedDomain, error) {
	var res []*database.ProjectWhitelistedDomain
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM projects_autoinvite_domains WHERE lower(domain)=lower($1)", domain)
	if err != nil {
		return nil, parseErr("project whitelist domains", err)
	}
	return res, nil
}

func (c *connection) FindProjectWhitelistedDomain(ctx context.Context, projectID, domain string) (*database.ProjectWhitelistedDomain, error) {
	res := &database.ProjectWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM projects_autoinvite_domains WHERE project_id=$1 AND lower(domain)=lower($2)", projectID, domain).StructScan(res)
	if err != nil {
		return nil, parseErr("project whitelist domain", err)
	}
	return res, nil
}

func (c *connection) InsertProjectWhitelistedDomain(ctx context.Context, opts *database.InsertProjectWhitelistedDomainOptions) (*database.ProjectWhitelistedDomain, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ProjectWhitelistedDomain{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO projects_autoinvite_domains(project_id, project_role_id, domain) VALUES ($1, $2, $3) RETURNING *`, opts.ProjectID, opts.ProjectRoleID, opts.Domain).StructScan(res)
	if err != nil {
		return nil, parseErr("project whitelist domain", err)
	}
	return res, nil
}

func (c *connection) DeleteProjectWhitelistedDomain(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM projects_autoinvite_domains WHERE id=$1", id)
	return checkDeleteRow("project whitelist domain", res, err)
}

func (c *connection) FindDeployments(ctx context.Context, afterID string, limit int) ([]*database.Deployment, error) {
	var qry strings.Builder
	var args []any
	qry.WriteString("SELECT d.* FROM deployments d ")
	if afterID != "" {
		qry.WriteString("WHERE d.id > $1 ORDER BY d.id LIMIT $2")
		args = []any{afterID, limit}
	} else {
		qry.WriteString("ORDER BY d.id LIMIT $1")
		args = []any{limit}
	}
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, qry.String(), args...)
	if err != nil {
		return nil, parseErr("deployments", err)
	}
	return res, nil
}

// FindExpiredDeployments returns all the deployments which are expired as per prod ttl
func (c *connection) FindExpiredDeployments(ctx context.Context) ([]*database.Deployment, error) {
	var res []*database.Deployment
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT d.* FROM deployments d
		JOIN projects p ON d.project_id = p.id
		WHERE p.prod_ttl_seconds IS NOT NULL AND d.used_on + p.prod_ttl_seconds * interval '1 second' < now()
	`)
	if err != nil {
		return nil, parseErr("deployments", err)
	}
	return res, nil
}

func (c *connection) FindDeploymentsForProject(ctx context.Context, projectID string) ([]*database.Deployment, error) {
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

func (c *connection) FindDeploymentByInstanceID(ctx context.Context, instanceID string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM deployments d WHERE d.runtime_instance_id=$1", instanceID).StructScan(res)
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
		INSERT INTO deployments (project_id, branch, runtime_host, runtime_instance_id, runtime_audience, status, status_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`,
		opts.ProjectID, opts.Branch, opts.RuntimeHost, opts.RuntimeInstanceID, opts.RuntimeAudience, opts.Status, opts.StatusMessage,
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

func (c *connection) UpdateDeployment(ctx context.Context, id string, opts *database.UpdateDeploymentOptions) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE deployments
		SET branch=$1, runtime_host=$2, runtime_instance_id=$3, runtime_audience=$4, status=$5, status_message=$6, updated_on=now()
		WHERE id=$7 RETURNING *`,
		opts.Branch, opts.RuntimeHost, opts.RuntimeInstanceID, opts.RuntimeAudience, opts.Status, opts.StatusMessage, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentStatus(ctx context.Context, id string, status database.DeploymentStatus, message string) (*database.Deployment, error) {
	res := &database.Deployment{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE deployments SET status=$1, status_message=$2, updated_on=now() WHERE id=$3 RETURNING *", status, message, id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE deployments SET used_on=now() WHERE id = any($1)", ids)
	if err != nil {
		return parseErr("deployment", err)
	}
	return nil
}

func (c *connection) UpsertStaticRuntimeAssignment(ctx context.Context, id, host string, slots int) error {
	// If slots is 0, delete the assignment if it exists (may not exist due to idempotence, so not checking the affected row count).
	if slots == 0 {
		_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM static_runtime_assignments WHERE resource_id=$1", id)
		if err != nil {
			return parseErr("slots used", err)
		}
		return nil
	}

	// Upsert the assignment.
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO static_runtime_assignments (resource_id, host, slots) VALUES ($1, $2, $3) ON CONFLICT (resource_id) DO UPDATE SET slots = EXCLUDED.slots", id, host, slots)
	if err != nil {
		return parseErr("slots used", err)
	}
	return nil
}

func (c *connection) DeleteStaticRuntimeAssignment(ctx context.Context, id string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM static_runtime_assignments WHERE resource_id=$1", id)
	if err != nil {
		// Not using checkDeleteRow because this is operation must be idempotent, so the row may not exist.
		return parseErr("slots used", err)
	}
	return nil
}

func (c *connection) ResolveStaticRuntimeSlotsUsed(ctx context.Context) ([]*database.StaticRuntimeSlotsUsed, error) {
	var res []*database.StaticRuntimeSlotsUsed
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT host, SUM(slots) as slots FROM static_runtime_assignments GROUP BY host ORDER BY host")
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

func (c *connection) FindUsersByEmailPattern(ctx context.Context, emailPattern, afterEmail string, limit int) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT u.* FROM users u
	WHERE lower(u.email) LIKE lower($1) AND lower(u.email) > lower($2)
	ORDER BY lower(u.email) LIMIT $3`, emailPattern, afterEmail, limit)
	if err != nil {
		return nil, parseErr("users", err)
	}
	return res, nil
}

// SearchProjectUsers searches for users that have access to the project.
func (c *connection) SearchProjectUsers(ctx context.Context, projectID, emailQuery, afterEmail string, limit int) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT u.* FROM users u
		WHERE u.id IN (
			SELECT upr.user_id FROM users_projects_roles upr WHERE upr.project_id=$1
			UNION
			SELECT ugu.user_id FROM usergroups_projects_roles ugpr JOIN usergroups_users ugu ON ugpr.usergroup_id = ugu.usergroup_id WHERE ugpr.project_id=$1
		)
		AND lower(u.email) LIKE lower($2)
		AND lower(u.email) > lower($3)
		ORDER BY lower(u.email) ASC LIMIT $4`, projectID, emailQuery, afterEmail, limit)
	if err != nil {
		return nil, parseErr("users", err)
	}
	return res, nil
}

func (c *connection) InsertUser(ctx context.Context, opts *database.InsertUserOptions) (*database.User, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.User{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO users (email, display_name, photo_url, quota_trial_orgs, quota_singleuser_orgs, superuser) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *", opts.Email, opts.DisplayName, opts.PhotoURL, opts.QuotaTrialOrgs, opts.QuotaSingleuserOrgs, opts.Superuser).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) CheckUsersEmpty(ctx context.Context) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT NOT EXISTS (SELECT 1 FROM users limit 1) ").Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
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
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE users SET display_name=$2, photo_url=$3, github_username=$4, github_refresh_token=$5, quota_singleuser_orgs=$6, quota_trial_orgs=$7, preference_time_zone=$8, updated_on=now() WHERE id=$1 RETURNING *",
		id,
		opts.DisplayName,
		opts.PhotoURL,
		opts.GithubUsername,
		opts.GithubRefreshToken,
		opts.QuotaSingleuserOrgs,
		opts.QuotaTrialOrgs,
		opts.PreferenceTimeZone).StructScan(res)
	if err != nil {
		return nil, parseErr("user", err)
	}
	return res, nil
}

func (c *connection) UpdateUserActiveOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE users SET active_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("user", err)
	}
	return nil
}

func (c *connection) CheckUserIsAnOrganizationMember(ctx context.Context, userID, orgID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT EXISTS (SELECT 1 FROM users_orgs_roles WHERE user_id=$1 AND org_id=$2)", userID, orgID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) CheckUserIsAProjectMember(ctx context.Context, userID, projectID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT EXISTS (SELECT 1 FROM users_projects_roles WHERE user_id=$1 AND project_id=$2)", userID, projectID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) GetCurrentTrialOrgCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT current_trial_orgs_count FROM users WHERE id=$1", userID).Scan(&count)
	if err != nil {
		return 0, parseErr("org count", err)
	}
	return count, nil
}

func (c *connection) IncrementCurrentTrialOrgCount(ctx context.Context, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE users SET current_trial_orgs_count = current_trial_orgs_count + 1 WHERE id=$1", userID)
	if err != nil {
		return parseErr("org count", err)
	}
	return nil
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

func (c *connection) UpdateUsergroupName(ctx context.Context, name, groupID string) (*database.Usergroup, error) {
	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE usergroups SET name=$1, updated_on=now() WHERE id=$2 RETURNING *", name, groupID).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) UpdateUsergroupDescription(ctx context.Context, description, groupID string) (*database.Usergroup, error) {
	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "UPDATE usergroups SET description=$1, updated_on=now() WHERE id=$2 RETURNING *", description, groupID).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) DeleteUsergroup(ctx context.Context, groupID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups WHERE id=$1", groupID)
	return checkDeleteRow("usergroup", res, err)
}

func (c *connection) FindUsergroupByName(ctx context.Context, orgName, name string) (*database.Usergroup, error) {
	res := &database.Usergroup{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT ug.* FROM usergroups ug JOIN orgs o ON ug.org_id = o.id
		WHERE lower(ug.name)=lower($1) AND lower(o.name)=lower($2)
	`, name, orgName).StructScan(res)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) FindUsergroupsForUser(ctx context.Context, userID, orgID string) ([]*database.Usergroup, error) {
	var res []*database.Usergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT ug.* FROM usergroups ug JOIN usergroups_users uug ON ug.id = uug.usergroup_id
		WHERE uug.user_id = $1 AND ug.org_id = $2
	`, userID, orgID)
	if err != nil {
		return nil, parseErr("usergroup", err)
	}
	return res, nil
}

func (c *connection) InsertUsergroupMemberUser(ctx context.Context, groupID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO usergroups_users (user_id, usergroup_id) VALUES ($1, $2)", userID, groupID)
	if err != nil {
		return parseErr("usergroup member", err)
	}
	return nil
}

func (c *connection) FindUsergroupMemberUsers(ctx context.Context, groupID, afterEmail string, limit int) ([]*database.MemberUser, error) {
	var res []*database.MemberUser
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT uug.user_id as "id", u.email, u.display_name, u.photo_url FROM usergroups_users uug
		JOIN users u ON uug.user_id = u.id
		WHERE uug.usergroup_id = $1 AND lower(u.email) > lower($2)
		ORDER BY lower(u.email) LIMIT $3
	`, groupID, afterEmail, limit)
	if err != nil {
		return nil, parseErr("usergroup member", err)
	}
	return res, nil
}

func (c *connection) DeleteUsergroupMemberUser(ctx context.Context, groupID, userID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_users WHERE user_id = $1 AND usergroup_id = $2", userID, groupID)
	return checkDeleteRow("usergroup member", res, err)
}

func (c *connection) CheckUsergroupExists(ctx context.Context, groupID string) (bool, error) {
	var res bool
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT EXISTS (SELECT 1 FROM usergroups WHERE id=$1)", groupID).Scan(&res)
	if err != nil {
		return false, parseErr("check", err)
	}
	return res, nil
}

func (c *connection) DeleteUsergroupsMemberUser(ctx context.Context, orgID, userID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		DELETE FROM usergroups_users WHERE user_id = $1 AND usergroup_id IN (SELECT id FROM usergroups WHERE org_id = $2)
	`, userID, orgID)
	if err != nil {
		return parseErr("usergroup member", err)
	}
	return nil
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
		INSERT INTO user_auth_tokens (id, secret_hash, user_id, display_name, auth_client_id, representing_user_id, expires_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`,
		opts.ID, opts.SecretHash, opts.UserID, opts.DisplayName, opts.AuthClientID, opts.RepresentingUserID, opts.ExpiresOn,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("auth token", err)
	}
	return res, nil
}

func (c *connection) UpdateUserAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE user_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("auth token", err)
	}
	return nil
}

func (c *connection) DeleteUserAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("auth token", res, err)
}

func (c *connection) DeleteExpiredUserAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM user_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("auth token", err)
}

// FindServicesByOrgID returns a list of services in an org.
func (c *connection) FindServicesByOrgID(ctx context.Context, orgID string) ([]*database.Service, error) {
	var res []*database.Service

	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT * FROM service WHERE org_id=$1", orgID)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res, nil
}

// FindService returns a service.
func (c *connection) FindService(ctx context.Context, id string) (*database.Service, error) {
	res := &database.Service{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM service WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res, nil
}

// FindServiceByName returns a service.
func (c *connection) FindServiceByName(ctx context.Context, orgID, name string) (*database.Service, error) {
	res := &database.Service{}

	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM service WHERE org_id=$1 AND name=$2", orgID, name).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res, nil
}

// InsertService inserts a service.
func (c *connection) InsertService(ctx context.Context, opts *database.InsertServiceOptions) (*database.Service, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Service{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO service (org_id, name)
		VALUES ($1, $2) RETURNING *`,
		opts.OrgID, opts.Name,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res, nil
}

// UpdateService updates a service.
func (c *connection) UpdateService(ctx context.Context, id string, opts *database.UpdateServiceOptions) (*database.Service, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Service{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE service
		SET name=$1
		WHERE id=$2 RETURNING *`,
		opts.Name, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("service", err)
	}
	return res, nil
}

// UpdateServiceActiceOn updates a service's active_on timestamp.
func (c *connection) UpdateServiceActiveOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE service SET active_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("service", err)
	}
	return nil
}

// DeleteService deletes a service.
func (c *connection) DeleteService(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service WHERE id=$1", id)
	return checkDeleteRow("service", res, err)
}

// FindSeviceAuthTokens returns a list of service auth tokens.
func (c *connection) FindServiceAuthTokens(ctx context.Context, serviceID string) ([]*database.ServiceAuthToken, error) {
	var res []*database.ServiceAuthToken
	err := c.getDB(ctx).SelectContext(ctx, &res, "SELECT t.* FROM service_auth_tokens t WHERE t.service_id=$1", serviceID)
	if err != nil {
		return nil, parseErr("service auth tokens", err)
	}
	return res, nil
}

// FindServiceAuthToken returns a service auth token.
func (c *connection) FindServiceAuthToken(ctx context.Context, id string) (*database.ServiceAuthToken, error) {
	res := &database.ServiceAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM service_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("service auth token", err)
	}
	return res, nil
}

// InsertServiceAuthToken inserts a service auth token.
func (c *connection) InsertServiceAuthToken(ctx context.Context, opts *database.InsertServiceAuthTokenOptions) (*database.ServiceAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ServiceAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO service_auth_tokens (id, secret_hash, service_id, expires_on)
		VALUES ($1, $2, $3, $4) RETURNING *`,
		opts.ID, opts.SecretHash, opts.ServiceID, opts.ExpiresOn,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("service auth token", err)
	}
	return res, nil
}

func (c *connection) UpdateServiceAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE service_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("service auth token", err)
	}
	return nil
}

// DeleteServiceAuthToken deletes a service auth token.
func (c *connection) DeleteServiceAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("service auth token", res, err)
}

// DeleteExpiredServiceAuthTokens deletes expired service auth tokens.
func (c *connection) DeleteExpiredServiceAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM service_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("service auth token", err)
}

func (c *connection) FindDeploymentAuthToken(ctx context.Context, id string) (*database.DeploymentAuthToken, error) {
	res := &database.DeploymentAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM deployment_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment auth token", err)
	}
	return res, nil
}

func (c *connection) InsertDeploymentAuthToken(ctx context.Context, opts *database.InsertDeploymentAuthTokenOptions) (*database.DeploymentAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.DeploymentAuthToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO deployment_auth_tokens (id, secret_hash, deployment_id, expires_on)
		VALUES ($1, $2, $3, $4) RETURNING *`,
		opts.ID, opts.SecretHash, opts.DeploymentID, opts.ExpiresOn,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("deployment auth token", err)
	}
	return res, nil
}

func (c *connection) UpdateDeploymentAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE deployment_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("deployment auth token", err)
	}
	return nil
}

func (c *connection) DeleteExpiredDeploymentAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM deployment_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("deployment auth token", err)
}

func (c *connection) FindMagicAuthTokensWithUser(ctx context.Context, projectID string, createdByUserID *string, afterID string, limit int) ([]*database.MagicAuthTokenWithUser, error) {
	n := 1
	where := fmt.Sprintf("t.project_id=$%d", n)
	args := []any{projectID}
	n++

	if createdByUserID != nil {
		where = fmt.Sprintf("%s AND t.created_by_user_id=$%d", where, n)
		args = append(args, *createdByUserID)
		n++
	}

	if afterID != "" {
		where = fmt.Sprintf("%s AND t.id>$%d", where, n)
		args = append(args, afterID)
		n++
	}

	where += " AND (t.expires_on IS NULL OR t.expires_on > now()) AND t.internal=false"

	qry := fmt.Sprintf("SELECT t.*, u.email AS created_by_user_email FROM magic_auth_tokens t LEFT JOIN users u ON t.created_by_user_id=u.id WHERE %s ORDER BY t.id LIMIT $%d", where, n)
	args = append(args, limit)

	var dtos []*magicAuthTokenWithUserDTO
	err := c.getDB(ctx).SelectContext(ctx, &dtos, qry, args...)
	if err != nil {
		return nil, parseErr("magic auth tokens", err)
	}

	res := make([]*database.MagicAuthTokenWithUser, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = c.magicAuthTokenWithUserFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c *connection) FindMagicAuthToken(ctx context.Context, id string, withSecret bool) (*database.MagicAuthToken, error) {
	res := &magicAuthTokenDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.* FROM magic_auth_tokens t WHERE t.id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("magic auth token", err)
	}
	return c.magicAuthTokenFromDTO(res, withSecret)
}

func (c *connection) FindMagicAuthTokenWithUser(ctx context.Context, id string) (*database.MagicAuthTokenWithUser, error) {
	res := &magicAuthTokenWithUserDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT t.*, u.email AS created_by_user_email FROM magic_auth_tokens t LEFT JOIN users u ON t.created_by_user_id=u.id WHERE t.id=$1 AND t.internal=false", id).StructScan(res)
	if err != nil {
		return nil, parseErr("magic auth token", err)
	}
	return c.magicAuthTokenWithUserFromDTO(res)
}

func (c *connection) InsertMagicAuthToken(ctx context.Context, opts *database.InsertMagicAuthTokenOptions) (*database.MagicAuthToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	if opts.Fields == nil {
		opts.Fields = []string{}
	}

	encSecret, encKeyID, err := c.encrypt(opts.Secret)
	if err != nil {
		return nil, err
	}

	res := &magicAuthTokenDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO magic_auth_tokens (id, secret_hash, secret, secret_encryption_key_id, project_id, expires_on, created_by_user_id, attributes, resource_type, resource_name, filter_json, fields, state, display_name, internal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING *`,
		opts.ID, opts.SecretHash, encSecret, encKeyID, opts.ProjectID, opts.ExpiresOn, opts.CreatedByUserID, opts.Attributes, opts.ResourceType, opts.ResourceName, opts.FilterJSON, opts.Fields, opts.State, opts.DisplayName, opts.Internal,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("magic auth token", err)
	}
	return c.magicAuthTokenFromDTO(res, true)
}

func (c *connection) UpdateMagicAuthTokenUsedOn(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "UPDATE magic_auth_tokens SET used_on=now() WHERE id=ANY($1)", ids)
	if err != nil {
		return parseErr("magic auth token", err)
	}
	return nil
}

func (c *connection) DeleteMagicAuthToken(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM magic_auth_tokens WHERE id=$1", id)
	return checkDeleteRow("magic auth token", res, err)
}

func (c *connection) DeleteMagicAuthTokens(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM magic_auth_tokens WHERE id=ANY($1)", ids)
	return parseErr("magic auth token", err)
}

func (c *connection) DeleteExpiredMagicAuthTokens(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM magic_auth_tokens WHERE expires_on IS NOT NULL AND expires_on + $1 < now()", retention)
	return parseErr("magic auth token", err)
}

func (c *connection) FindReportTokens(ctx context.Context, reportName string) ([]*database.ReportToken, error) {
	var res []*database.ReportToken
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT * FROM report_tokens WHERE report_name=$1`, reportName)
	if err != nil {
		return nil, parseErr("report tokens", err)
	}
	return res, nil
}

func (c *connection) FindReportTokensWithSecret(ctx context.Context, reportName string) ([]*database.ReportTokenWithSecret, error) {
	var res []*reportTokenWithSecretDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT t.*, m.secret as magic_auth_token_secret, m.secret_encryption_key_id FROM report_tokens t JOIN magic_auth_tokens m ON t.magic_auth_token_id=m.id WHERE t.report_name=$1`, reportName)
	if err != nil {
		return nil, parseErr("report tokens", err)
	}

	ret := make([]*database.ReportTokenWithSecret, len(res))
	for i, dto := range res {
		ret[i], err = c.reportTokenWithSecretFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (c *connection) FindReportTokenForMagicAuthToken(ctx context.Context, magicAuthTokenID string) (*database.ReportToken, error) {
	res := &database.ReportToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM report_tokens WHERE magic_auth_token_id=$1`, magicAuthTokenID).StructScan(res)
	if err != nil {
		return nil, parseErr("report token", err)
	}
	return res, nil
}

func (c *connection) InsertReportToken(ctx context.Context, opts *database.InsertReportTokenOptions) (*database.ReportToken, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ReportToken{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO report_tokens (report_name, recipient_email, magic_auth_token_id) VALUES ($1, $2, $3) RETURNING *`, opts.ReportName, opts.RecipientEmail, opts.MagicAuthTokenID).StructScan(res)
	if err != nil {
		return nil, parseErr("report token", err)
	}
	return res, nil
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

func (c *connection) DeleteExpiredDeviceAuthCodes(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM device_auth_codes WHERE expires_on + $1 < now()", retention)
	return parseErr("device auth code", err)
}

func (c *connection) FindAuthorizationCode(ctx context.Context, code string) (*database.AuthorizationCode, error) {
	authCode := &database.AuthorizationCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM authorization_codes WHERE code = $1", code).StructScan(authCode)
	if err != nil {
		return nil, parseErr("authorization code", err)
	}
	return authCode, nil
}

func (c *connection) InsertAuthorizationCode(ctx context.Context, code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod string, expiration time.Time) (*database.AuthorizationCode, error) {
	res := &database.AuthorizationCode{}
	err := c.getDB(ctx).QueryRowxContext(ctx,
		`INSERT INTO authorization_codes (code, user_id, client_id, redirect_uri, code_challenge, code_challenge_method, expires_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`, code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod, expiration).StructScan(res)
	if err != nil {
		return nil, parseErr("authorization code", err)
	}
	return res, nil
}

func (c *connection) DeleteAuthorizationCode(ctx context.Context, code string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM authorization_codes WHERE code=$1", code)
	return checkDeleteRow("authorization code", res, err)
}

func (c *connection) DeleteExpiredAuthorizationCodes(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM authorization_codes WHERE expires_on + $1 < now()", retention)
	return parseErr("authorization code", err)
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

func (c *connection) FindOrganizationMemberUsers(ctx context.Context, orgID, afterEmail string, limit int) ([]*database.MemberUser, error) {
	var res []*database.MemberUser
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on, r.name FROM users u
    	JOIN users_orgs_roles uor ON u.id = uor.user_id
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.org_id=$1 AND lower(u.email) > lower($2)
		ORDER BY lower(u.email) LIMIT $3
	`, orgID, afterEmail, limit)
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

func (c *connection) FindOrganizationMembersWithManageUsersRole(ctx context.Context, orgID string) ([]*database.MemberUser, error) {
	var res []*database.MemberUser
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on, r.name FROM users u
			JOIN users_orgs_roles uor ON u.id = uor.user_id
		JOIN org_roles r ON r.id = uor.org_role_id
		WHERE uor.org_id=$1 AND r.manage_org_members=true
		ORDER BY lower(u.email)
	`, orgID)
	if err != nil {
		return nil, parseErr("org members", err)
	}
	return res, nil
}

func (c *connection) FindProjectMemberUsers(ctx context.Context, projectID, afterEmail string, limit int) ([]*database.MemberUser, error) {
	var res []*database.MemberUser
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT u.id, u.email, u.display_name, u.photo_url, u.created_on, u.updated_on, r.name FROM users u
    	JOIN users_projects_roles upr ON u.id = upr.user_id
		JOIN project_roles r ON r.id = upr.project_role_id
		WHERE upr.project_id=$1 AND lower(u.email) > lower($2)
		ORDER BY lower(u.email) LIMIT $3
	`, projectID, afterEmail, limit)
	if err != nil {
		return nil, parseErr("project members", err)
	}
	return res, nil
}

func (c *connection) FindSuperusers(ctx context.Context) ([]*database.User, error) {
	var res []*database.User
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT u.* FROM users u WHERE u.superuser = true`)
	if err != nil {
		return nil, parseErr("project members", err)
	}
	return res, nil
}

func (c *connection) UpdateSuperuser(ctx context.Context, userID string, superuser bool) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE users SET superuser=$2, updated_on=now() WHERE id=$1`, userID, superuser)
	return checkUpdateRow("superuser", res, err)
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

// FindOrganizationMemberUsergroups returns org user groups as a collection of MemberUsergroup.
// If a user group has no org role then RoleName is empty.
func (c *connection) FindOrganizationMemberUsergroups(ctx context.Context, orgID, afterName string, limit int) ([]*database.MemberUsergroup, error) {
	var res []*database.MemberUsergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT ug.id, ug.name, ug.created_on, ug.updated_on, COALESCE(r.name, '') as "role_name" FROM usergroups ug
		LEFT JOIN usergroups_orgs_roles uor ON ug.id = uor.usergroup_id
		LEFT JOIN org_roles r ON uor.org_role_id = r.id
		WHERE ug.org_id=$1 AND lower(ug.name) > lower($2)
		ORDER BY lower(ug.name) LIMIT $3
	`, orgID, afterName, limit)
	if err != nil {
		return nil, parseErr("org groups", err)
	}
	return res, nil
}

func (c *connection) InsertOrganizationMemberUsergroup(ctx context.Context, groupID, orgID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO usergroups_orgs_roles (usergroup_id, org_id, org_role_id) VALUES ($1, $2, $3)
	`, groupID, orgID, roleID)
	if err != nil {
		return parseErr("org group member", err)
	}
	return nil
}

func (c *connection) UpdateOrganizationMemberUsergroup(ctx context.Context, groupID, orgID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE usergroups_orgs_roles SET org_role_id = $3 WHERE usergroup_id = $1 AND org_id = $2
	`, groupID, orgID, roleID)
	return checkUpdateRow("org group member", res, err)
}

func (c *connection) DeleteOrganizationMemberUsergroup(ctx context.Context, groupID, orgID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_orgs_roles WHERE usergroup_id = $1 AND org_id = $2", groupID, orgID)
	return checkDeleteRow("org group member", res, err)
}

func (c *connection) FindProjectMemberUsergroups(ctx context.Context, projectID, afterName string, limit int) ([]*database.MemberUsergroup, error) {
	var res []*database.MemberUsergroup
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT ug.id, ug.name, ug.created_on, ug.updated_on, r.name as "role_name" FROM usergroups ug
		JOIN usergroups_projects_roles upr ON ug.id = upr.usergroup_id
		JOIN project_roles r ON upr.project_role_id = r.id
		WHERE upr.project_id=$1 AND lower(ug.name) > lower($2)
		ORDER BY lower(ug.name) LIMIT $3
	`, projectID, afterName, limit)
	if err != nil {
		return nil, parseErr("project groups", err)
	}
	return res, nil
}

func (c *connection) InsertProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO usergroups_projects_roles (usergroup_id, project_id, project_role_id) VALUES ($1, $2, $3)
	`, groupID, projectID, roleID)
	if err != nil {
		return parseErr("project group member", err)
	}
	return nil
}

func (c *connection) UpdateProjectMemberUsergroup(ctx context.Context, groupID, projectID, roleID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE usergroups_projects_roles SET project_role_id = $3 WHERE usergroup_id = $1 AND project_id = $2
	`, groupID, projectID, roleID)
	return checkUpdateRow("project group member", res, err)
}

func (c *connection) DeleteProjectMemberUsergroup(ctx context.Context, groupID, projectID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM usergroups_projects_roles WHERE usergroup_id = $1 AND project_id = $2", groupID, projectID)
	return checkDeleteRow("project group member", res, err)
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

func (c *connection) FindOrganizationInvites(ctx context.Context, orgID, afterEmail string, limit int) ([]*database.Invite, error) {
	var res []*database.Invite
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT uoi.email, ur.name as role, u.email as invited_by
		FROM org_invites uoi JOIN org_roles ur ON uoi.org_role_id = ur.id JOIN users u ON uoi.invited_by_user_id = u.id
		WHERE uoi.org_id = $1 AND lower(uoi.email) > lower($2)
		ORDER BY lower(uoi.email) LIMIT $3
	`, orgID, afterEmail, limit)
	if err != nil {
		return nil, parseErr("org invites", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationInvitesByEmail(ctx context.Context, userEmail string) ([]*database.OrganizationInvite, error) {
	var dtos []*organizationInviteDTO
	err := c.getDB(ctx).SelectContext(ctx, &dtos, "SELECT * FROM org_invites WHERE lower(email) = lower($1)", userEmail)
	if err != nil {
		return nil, parseErr("org invites", err)
	}
	res := make([]*database.OrganizationInvite, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = dto.AsModel()
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c *connection) FindOrganizationInvite(ctx context.Context, orgID, userEmail string) (*database.OrganizationInvite, error) {
	dto := &organizationInviteDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM org_invites WHERE lower(email) = lower($1) AND org_id = $2", userEmail, orgID).StructScan(dto)
	if err != nil {
		return nil, parseErr("org invite", err)
	}
	return dto.AsModel()
}

func (c *connection) InsertOrganizationInvite(ctx context.Context, opts *database.InsertOrganizationInviteOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}

	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO org_invites (email, invited_by_user_id, org_id, org_role_id) VALUES ($1, $2, $3, $4)", opts.Email, opts.InviterID, opts.OrgID, opts.RoleID)
	if err != nil {
		return parseErr("org invite", err)
	}
	return nil
}

func (c *connection) UpdateOrganizationInviteUsergroups(ctx context.Context, id string, groupIDs []string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE org_invites SET usergroup_ids = $1 WHERE id = $2`, groupIDs, id)
	return checkUpdateRow("org invite", res, err)
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

func (c *connection) FindProjectInvites(ctx context.Context, projectID, afterEmail string, limit int) ([]*database.Invite, error) {
	var res []*database.Invite
	err := c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT upi.email, ur.name as role, u.email as invited_by
			FROM project_invites upi JOIN project_roles ur ON upi.project_role_id = ur.id JOIN users u ON upi.invited_by_user_id = u.id
			WHERE upi.project_id = $1 AND lower(upi.email) > lower($2)
			ORDER BY lower(upi.email) LIMIT $3
	`, projectID, afterEmail, limit)
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

func (c *connection) InsertProjectInvite(ctx context.Context, opts *database.InsertProjectInviteOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}

	_, err := c.getDB(ctx).ExecContext(ctx, "INSERT INTO project_invites (email, invited_by_user_id, project_id, project_role_id) VALUES ($1, $2, $3, $4)", opts.Email, opts.InviterID, opts.ProjectID, opts.RoleID)
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

func (c *connection) FindProjectAccessRequests(ctx context.Context, projectID, afterID string, limit int) ([]*database.ProjectAccessRequest, error) {
	var res []*database.ProjectAccessRequest
	var err error
	if afterID != "" {
		err = c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT par.user_id
			FROM project_access_requests par
			WHERE par.project_id = $1 AND par.user_id > $2
			ORDER BY par.user_id LIMIT $3
		`, projectID, afterID, limit)
	} else {
		err = c.getDB(ctx).SelectContext(ctx, &res, `
			SELECT par.user_id
			FROM project_access_requests par
			WHERE par.project_id = $1
			ORDER BY par.user_id LIMIT $2
		`, projectID, limit)
	}
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) FindProjectAccessRequest(ctx context.Context, projectID, userID string) (*database.ProjectAccessRequest, error) {
	res := &database.ProjectAccessRequest{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_access_requests WHERE user_id = $1 AND project_id = $2", userID, projectID).StructScan(res)
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) FindProjectAccessRequestByID(ctx context.Context, id string) (*database.ProjectAccessRequest, error) {
	res := &database.ProjectAccessRequest{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM project_access_requests WHERE id=$1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) InsertProjectAccessRequest(ctx context.Context, opts *database.InsertProjectAccessRequestOptions) (*database.ProjectAccessRequest, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.ProjectAccessRequest{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "INSERT INTO project_access_requests (user_id, project_id) VALUES ($1, $2) RETURNING *", opts.UserID, opts.ProjectID).StructScan(res)
	if err != nil {
		return nil, parseErr("project access request", err)
	}
	return res, nil
}

func (c *connection) DeleteProjectAccessRequest(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM project_access_requests WHERE id = $1", id)
	return checkDeleteRow("project access request", res, err)
}

// FindBookmarks returns a list of bookmarks for a user per project
func (c *connection) FindBookmarks(ctx context.Context, projectID, resourceKind, resourceName, userID string) ([]*database.Bookmark, error) {
	var res []*database.Bookmark
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT * FROM bookmarks WHERE project_id = $1 and resource_kind = $2 and lower(resource_name) = lower($3) and (user_id = $4 or shared = true or "default" = true)`,
		projectID, resourceKind, resourceName, userID)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

// FindBookmark returns a bookmark for given bookmark id
func (c *connection) FindBookmark(ctx context.Context, bookmarkID string) (*database.Bookmark, error) {
	res := &database.Bookmark{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM bookmarks WHERE id = $1", bookmarkID).StructScan(res)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

func (c *connection) FindDefaultBookmark(ctx context.Context, projectID, resourceKind, resourceName string) (*database.Bookmark, error) {
	res := &database.Bookmark{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM bookmarks WHERE project_id = $1 and resource_kind = $2 and lower(resource_name) = lower($3) and "default" = true`,
		projectID, resourceKind, resourceName).StructScan(res)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

// InsertBookmark inserts a bookmark for a user per project
func (c *connection) InsertBookmark(ctx context.Context, opts *database.InsertBookmarkOptions) (*database.Bookmark, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	res := &database.Bookmark{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO bookmarks (display_name, description, data, resource_kind, resource_name, project_id, user_id, "default", shared)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *`,
		opts.DisplayName, opts.Description, opts.Data, opts.ResourceKind, opts.ResourceName, opts.ProjectID, opts.UserID, opts.Default, opts.Shared).StructScan(res)
	if err != nil {
		return nil, parseErr("bookmarks", err)
	}
	return res, nil
}

func (c *connection) UpdateBookmark(ctx context.Context, opts *database.UpdateBookmarkOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE bookmarks SET display_name=$1, description=$2, data=$3, shared=$4 WHERE id=$5`,
		opts.DisplayName, opts.Description, opts.Data, opts.Shared, opts.BookmarkID)
	return checkUpdateRow("bookmark", res, err)
}

// DeleteBookmark deletes a bookmark for a given bookmark id
func (c *connection) DeleteBookmark(ctx context.Context, bookmarkID string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM bookmarks WHERE id = $1", bookmarkID)
	return checkDeleteRow("bookmarks", res, err)
}

func (c *connection) FindVirtualFiles(ctx context.Context, projectID, branch string, afterUpdatedOn time.Time, afterPath string, limit int) ([]*database.VirtualFile, error) {
	var res []*database.VirtualFile
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT path, data, deleted, updated_on
		FROM virtual_files
		WHERE project_id=$1 AND branch=$2 AND (updated_on>$3 OR updated_on=$3 AND path>$4)
		ORDER BY updated_on, path LIMIT $5
	`, projectID, branch, afterUpdatedOn, afterPath, limit)
	if err != nil {
		return nil, parseErr("virtual files", err)
	}
	return res, nil
}

func (c *connection) FindVirtualFile(ctx context.Context, projectID, branch, path string) (*database.VirtualFile, error) {
	res := &database.VirtualFile{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		SELECT path, data, deleted, updated_on
		FROM virtual_files
		WHERE project_id=$1 AND branch=$2 AND path=$3
	`, projectID, branch, path).StructScan(res)
	if err != nil {
		return nil, parseErr("virtual files", err)
	}
	return res, nil
}

func (c *connection) UpsertVirtualFile(ctx context.Context, opts *database.InsertVirtualFileOptions) error {
	if err := database.Validate(opts); err != nil {
		return err
	}

	_, err := c.getDB(ctx).ExecContext(ctx, `
		INSERT INTO virtual_files (project_id, branch, path, data, deleted)
		VALUES ($1, $2, $3, $4, FALSE)
		ON CONFLICT (project_id, branch, path) DO UPDATE SET
			data = EXCLUDED.data,
			deleted = FALSE,
			updated_on = now()
	`, opts.ProjectID, opts.Branch, opts.Path, opts.Data)
	if err != nil {
		return parseErr("virtual file", err)
	}
	return nil
}

func (c *connection) UpdateVirtualFileDeleted(ctx context.Context, projectID, branch, path string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `
		UPDATE virtual_files SET
			data = ''::BYTEA,
			deleted = TRUE,
			updated_on = now()
		WHERE project_id=$1 AND branch=$2 AND path=$3`, projectID, branch, path)
	return checkUpdateRow("virtual file", res, err)
}

func (c *connection) DeleteExpiredVirtualFiles(ctx context.Context, retention time.Duration) error {
	_, err := c.getDB(ctx).ExecContext(ctx, `DELETE FROM virtual_files WHERE deleted AND updated_on + $1 < now()`, retention)
	return parseErr("virtual files", err)
}

func (c *connection) FindAsset(ctx context.Context, id string) (*database.Asset, error) {
	res := &database.Asset{}
	err := c.getDB(ctx).QueryRowxContext(ctx, "SELECT * FROM assets WHERE id = $1", id).StructScan(res)
	if err != nil {
		return nil, parseErr("asset", err)
	}
	return res, nil
}

func (c *connection) InsertAsset(ctx context.Context, id, organizationID, path, ownerID string, public bool) (*database.Asset, error) {
	res := &database.Asset{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO assets (id, org_id, path, owner_id, public)
		VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		id, organizationID, path, ownerID, public,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("asset", err)
	}
	return res, nil
}

func (c *connection) FindUnusedAssets(ctx context.Context, limit int) ([]*database.Asset, error) {
	var res []*database.Asset
	// find assets that are not associated with any project or org
	// skip assets that are less than 7 days old to avoid deleting assets for projects
	// that were accidentally deleted and may need to be restored
	err := c.getDB(ctx).SelectContext(ctx, &res, `
		SELECT a.* FROM assets a 
		WHERE a.created_on < now() - INTERVAL '7 DAYS'
		AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.archive_asset_id = a.id)
		AND NOT EXISTS (SELECT 1 FROM orgs o WHERE o.logo_asset_id = a.id)
		ORDER BY a.created_on DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, parseErr("assets", err)
	}
	return res, nil
}

func (c *connection) DeleteAssets(ctx context.Context, ids []string) error {
	_, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM assets WHERE id=ANY($1)", ids)
	return parseErr("asset", err)
}

func (c *connection) FindOrganizationIDsWithBilling(ctx context.Context) ([]string, error) {
	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT id FROM orgs WHERE billing_customer_id <> ''`)
	if err != nil {
		return nil, parseErr("billing orgs", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationIDsWithoutBilling(ctx context.Context) ([]string, error) {
	var res []string
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT id FROM orgs WHERE billing_customer_id = ''`)
	if err != nil {
		return nil, parseErr("billing orgs without billing or payment info", err)
	}
	return res, nil
}

func (c *connection) CountBillingProjectsForOrganization(ctx context.Context, orgID string, createdBefore time.Time) (int, error) {
	var count int
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT COUNT(*) FROM projects WHERE org_id = $1 AND prod_deployment_id IS NOT NULL AND created_on < $2`, orgID, createdBefore).Scan(&count)
	if err != nil {
		return 0, parseErr("billing projects", err)
	}
	return count, nil
}

func (c *connection) FindBillingUsageReportedOn(ctx context.Context) (time.Time, error) {
	var usageReportedOn sql.NullTime
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT usage_reported_on FROM billing_reporting_time`).Scan(&usageReportedOn)
	if err != nil {
		return time.Time{}, parseErr("billing usage", err)
	}
	if !usageReportedOn.Valid {
		return time.Time{}, nil
	}
	return usageReportedOn.Time, nil
}

func (c *connection) UpdateBillingUsageReportedOn(ctx context.Context, usageReportedOn time.Time) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE billing_reporting_time SET usage_reported_on=$1`, usageReportedOn)
	return checkUpdateRow("billing usage", res, err)
}

func (c *connection) FindOrganizationForPaymentCustomerID(ctx context.Context, customerID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM orgs WHERE payment_customer_id = $1`, customerID).StructScan(res)
	if err != nil {
		return nil, parseErr("billing org for payment id", err)
	}
	return res, nil
}

func (c *connection) FindOrganizationForBillingCustomerID(ctx context.Context, customerID string) (*database.Organization, error) {
	res := &database.Organization{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM orgs WHERE billing_customer_id = $1`, customerID).StructScan(res)
	if err != nil {
		return nil, parseErr("billing org for billing id", err)
	}
	return res, nil
}

func (c *connection) FindBillingIssuesForOrg(ctx context.Context, orgID string) ([]*database.BillingIssue, error) {
	var res []*billingIssueDTO
	err := c.db.SelectContext(ctx, &res, `SELECT * FROM billing_issues WHERE org_id = $1`, orgID)
	if err != nil {
		return nil, parseErr("billing issues", err)
	}

	var billingErrors []*database.BillingIssue
	for _, dto := range res {
		billingErrors = append(billingErrors, dto.AsModel())
	}
	return billingErrors, nil
}

func (c *connection) FindBillingIssueByTypeForOrg(ctx context.Context, orgID string, errorType database.BillingIssueType) (*database.BillingIssue, error) {
	res := &billingIssueDTO{}
	err := c.db.GetContext(ctx, res, `SELECT * FROM billing_issues WHERE org_id = $1 AND type = $2`, orgID, errorType)
	if err != nil {
		return nil, parseErr("billing issue", err)
	}
	return res.AsModel(), nil
}

func (c *connection) FindBillingIssueByType(ctx context.Context, errorType database.BillingIssueType) ([]*database.BillingIssue, error) {
	var res []*billingIssueDTO
	err := c.db.SelectContext(ctx, &res, `SELECT * FROM billing_issues WHERE type = $1`, errorType)
	if err != nil {
		return nil, parseErr("billing issues", err)
	}

	var billingErrors []*database.BillingIssue
	for _, dto := range res {
		billingErrors = append(billingErrors, dto.AsModel())
	}
	return billingErrors, nil
}

func (c *connection) FindBillingIssueByTypeAndOverdueProcessed(ctx context.Context, errorType database.BillingIssueType, overdueProcessed bool) ([]*database.BillingIssue, error) {
	var res []*billingIssueDTO
	err := c.db.SelectContext(ctx, &res, `SELECT * FROM billing_issues WHERE type = $1 AND overdue_processed = $2`, errorType, overdueProcessed)
	if err != nil {
		return nil, parseErr("billing issues", err)
	}

	var billingErrors []*database.BillingIssue
	for _, dto := range res {
		billingErrors = append(billingErrors, dto.AsModel())
	}
	return billingErrors, nil
}

func (c *connection) UpsertBillingIssue(ctx context.Context, opts *database.UpsertBillingIssueOptions) (*database.BillingIssue, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(opts.Metadata)
	if err != nil {
		return nil, err
	}

	temp := &billingIssueDTO{
		OrgID:     opts.OrgID,
		Type:      opts.Type,
		Metadata:  metadata,
		EventTime: opts.EventTime,
	}

	temp.Level = temp.getBillingIssueLevel()

	res := &billingIssueDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `INSERT INTO billing_issues (org_id, type, level, metadata, event_time) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (org_id, type) DO UPDATE SET metadata = $4, event_time = $5 RETURNING *`, temp.OrgID, temp.Type, temp.Level, temp.Metadata, temp.EventTime).StructScan(res)
	if err != nil {
		return nil, parseErr("billing issue", err)
	}
	return res.AsModel(), nil
}

func (c *connection) UpdateBillingIssueOverdueAsProcessed(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, `UPDATE billing_issues SET overdue_processed = true WHERE id = $1`, id)
	return checkUpdateRow("billing issue", res, err)
}

func (c *connection) DeleteBillingIssue(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM billing_issues WHERE id = $1", id)
	return checkDeleteRow("billing issue", res, err)
}

func (c *connection) DeleteBillingIssueByTypeForOrg(ctx context.Context, orgID string, errorType database.BillingIssueType) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM billing_issues WHERE org_id = $1 AND type = $2", orgID, errorType)
	return checkDeleteRow("billing issue", res, err)
}

func (c *connection) FindProjectVariables(ctx context.Context, projectID string, environment *string) ([]*database.ProjectVariable, error) {
	q := `SELECT * FROM project_variables p WHERE p.project_id = $1`
	args := []interface{}{projectID}
	if environment != nil {
		// Also include variables that are not environment specific and not set for the given environment
		q += `
			AND (
				p.environment = $2
				OR (
					p.environment = ''
					AND NOT EXISTS (
						SELECT 1
						FROM project_variables p2
						WHERE p2.project_id = p.project_id
						AND p2.environment = $2
						AND lower(p2.name) = lower(p.name)
					)
				)
			)
		`
		args = append(args, environment)
	}
	var res []*database.ProjectVariable
	err := c.getDB(ctx).SelectContext(ctx, &res, q, args...)
	if err != nil {
		return nil, parseErr("project variables", err)
	}

	// Decrypt the variables
	err = c.decryptProjectVariables(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) UpsertProjectVariable(ctx context.Context, projectID, environment string, vars map[string]string, userID string) ([]*database.ProjectVariable, error) {
	query := `INSERT INTO project_variables (project_id, environment, name, value, value_encryption_key_id, updated_by_user_id, updated_on)
	VALUES %s
	ON CONFLICT (project_id, environment, lower(name)) DO UPDATE SET
		value = EXCLUDED.value,
		value_encryption_key_id = EXCLUDED.value_encryption_key_id,
		updated_by_user_id = EXCLUDED.updated_by_user_id,
		updated_on = now() RETURNING *`

	var placeholders strings.Builder
	args := []any{projectID, environment, userID}
	i := 3
	for key, value := range vars {
		// Encrypt the variables
		encryptedValue, valueEncryptionKeyID, err := c.encrypt([]byte(value))
		if err != nil {
			return nil, err
		}

		if valueEncryptionKeyID != "" {
			value = base64.StdEncoding.EncodeToString(encryptedValue)
		}
		args = append(args, key, value, valueEncryptionKeyID)
		if placeholders.Len() > 0 {
			placeholders.WriteString(", ")
		}
		fmt.Fprintf(&placeholders, "($1, $2, $%d, $%d, $%d, $3, now())", i+1, i+2, i+3) // project_id, environment, name, value, value_encryption_key_id, updated_by_user_id, updated_on
		i += 3
	}

	var res []*database.ProjectVariable
	err := c.getDB(ctx).SelectContext(ctx, &res, fmt.Sprintf(query, placeholders.String()), args...)
	if err != nil {
		return nil, parseErr("project variables", err)
	}

	// Decrypt the variables
	err = c.decryptProjectVariables(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) DeleteProjectVariables(ctx context.Context, projectID, environment string, names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("no names provided to delete project variables")
	}

	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names)+2)
	args[0] = projectID
	args[1] = environment
	for i, name := range names {
		placeholders[i] = fmt.Sprintf("lower($%d)", i+3)
		args[i+2] = name
	}

	query := fmt.Sprintf("DELETE FROM project_variables WHERE project_id = $1 AND environment = $2 AND lower(name) IN (%s)", strings.Join(placeholders, ","))
	_, err := c.getDB(ctx).ExecContext(ctx, query, args...)
	return err
}

func (c *connection) FindProvisionerResourcesForDeployment(ctx context.Context, deploymentID string) ([]*database.ProvisionerResource, error) {
	var res []*provisionerResourceDTO
	err := c.getDB(ctx).SelectContext(ctx, &res, `SELECT * FROM provisioner_resources WHERE deployment_id = $1`, deploymentID)
	if err != nil {
		return nil, parseErr("provisioner resources", err)
	}
	return c.provisionerResourcesFromDTOs(res)
}

func (c *connection) FindProvisionerResourceByTypeAndName(ctx context.Context, deploymentID, typ, name string) (*database.ProvisionerResource, error) {
	res := &provisionerResourceDTO{}
	err := c.getDB(ctx).QueryRowxContext(ctx, `SELECT * FROM provisioner_resources WHERE deployment_id = $1 AND "type" = $2 AND name = $3`, deploymentID, typ, name).StructScan(res)
	if err != nil {
		return nil, parseErr("provisioner resource", err)
	}
	return c.provisionerResourceFromDTO(res)
}

func (c *connection) InsertProvisionerResource(ctx context.Context, opts *database.InsertProvisionerResourceOptions) (*database.ProvisionerResource, error) {
	if err := database.Validate(opts); err != nil {
		return nil, err
	}

	args, err := json.Marshal(opts.Args)
	if err != nil {
		return nil, err
	}
	state, err := json.Marshal(opts.State)
	if err != nil {
		return nil, err
	}
	config, err := json.Marshal(opts.Config)
	if err != nil {
		return nil, err
	}

	res := &provisionerResourceDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		INSERT INTO provisioner_resources (id, deployment_id, "type", name, status, status_message, provisioner, args_json, state_json, config_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *`,
		opts.ID, opts.DeploymentID, opts.Type, opts.Name, opts.Status, opts.StatusMessage, opts.Provisioner, args, state, config,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("provisioner resource", err)
	}
	return c.provisionerResourceFromDTO(res)
}

func (c *connection) UpdateProvisionerResource(ctx context.Context, id string, opts *database.UpdateProvisionerResourceOptions) (*database.ProvisionerResource, error) {
	args, err := json.Marshal(opts.Args)
	if err != nil {
		return nil, err
	}
	state, err := json.Marshal(opts.State)
	if err != nil {
		return nil, err
	}
	config, err := json.Marshal(opts.Config)
	if err != nil {
		return nil, err
	}

	res := &provisionerResourceDTO{}
	err = c.getDB(ctx).QueryRowxContext(ctx, `
		UPDATE provisioner_resources SET status = $1, status_message = $2, args_json = $3, state_json = $4, config_json = $5, updated_on = now() WHERE id = $6 RETURNING *`,
		opts.Status, opts.StatusMessage, args, state, config, id,
	).StructScan(res)
	if err != nil {
		return nil, parseErr("provisioner resource", err)
	}
	return c.provisionerResourceFromDTO(res)
}

func (c *connection) DeleteProvisionerResource(ctx context.Context, id string) error {
	res, err := c.getDB(ctx).ExecContext(ctx, "DELETE FROM provisioner_resources WHERE id = $1", id)
	return checkDeleteRow("provisioner resource", res, err)
}

// projectDTO wraps database.Project, using the pgtype package to handle types that pgx can't read directly into their native Go types.
type projectDTO struct {
	*database.Project
	ProdVariables pgtype.JSON `db:"prod_variables"`
	Annotations   pgtype.JSON `db:"annotations"`
}

func (c *connection) projectFromDTO(dto *projectDTO) (*database.Project, error) {
	err := dto.Annotations.AssignTo(&dto.Project.Annotations)
	if err != nil {
		return nil, err
	}

	return dto.Project, nil
}

func (c *connection) projectsFromDTOs(dtos []*projectDTO) ([]*database.Project, error) {
	res := make([]*database.Project, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = c.projectFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// provisionerResourceDTO wraps database.ProvisionerResource, using the pgtype package to handle types that pgx can't read directly into their native Go types.
type provisionerResourceDTO struct {
	*database.ProvisionerResource
	Args   pgtype.JSON `db:"args_json"`
	State  pgtype.JSON `db:"state_json"`
	Config pgtype.JSON `db:"config_json"`
}

func (c *connection) provisionerResourceFromDTO(dto *provisionerResourceDTO) (*database.ProvisionerResource, error) {
	err := dto.Args.AssignTo(&dto.ProvisionerResource.Args)
	if err != nil {
		return nil, err
	}
	err = dto.State.AssignTo(&dto.ProvisionerResource.State)
	if err != nil {
		return nil, err
	}
	err = dto.Config.AssignTo(&dto.ProvisionerResource.Config)
	if err != nil {
		return nil, err
	}
	return dto.ProvisionerResource, nil
}

func (c *connection) provisionerResourcesFromDTOs(dtos []*provisionerResourceDTO) ([]*database.ProvisionerResource, error) {
	res := make([]*database.ProvisionerResource, len(dtos))
	for i, dto := range dtos {
		var err error
		res[i], err = c.provisionerResourceFromDTO(dto)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// magicAuthTokenDTO wraps database.MagicAuthToken, using the pgtype package to handly types that pgx can't read directly into their native Go types.
type magicAuthTokenDTO struct {
	*database.MagicAuthToken
	Attributes pgtype.JSON      `db:"attributes"`
	Fields     pgtype.TextArray `db:"fields"`
}

func (c *connection) magicAuthTokenFromDTO(dto *magicAuthTokenDTO, fetchSecret bool) (*database.MagicAuthToken, error) {
	err := dto.Attributes.AssignTo(&dto.MagicAuthToken.Attributes)
	if err != nil {
		return nil, err
	}
	err = dto.Fields.AssignTo(&dto.MagicAuthToken.Fields)
	if err != nil {
		return nil, err
	}

	if fetchSecret {
		dto.MagicAuthToken.Secret, err = c.decrypt(dto.MagicAuthToken.Secret, dto.MagicAuthToken.SecretEncryptionKeyID)
		if err != nil {
			return nil, err
		}
	} else {
		dto.MagicAuthToken.Secret = nil
		dto.MagicAuthToken.SecretEncryptionKeyID = ""
	}

	return dto.MagicAuthToken, nil
}

// magicAuthTokenWithUserDTO wraps database.MagicAuthTokenWithUser, using the pgtype package to handly types that pgx can't read directly into their native Go types.
type magicAuthTokenWithUserDTO struct {
	*database.MagicAuthTokenWithUser
	Attributes pgtype.JSON      `db:"attributes"`
	Fields     pgtype.TextArray `db:"fields"`
}

func (c *connection) magicAuthTokenWithUserFromDTO(dto *magicAuthTokenWithUserDTO) (*database.MagicAuthTokenWithUser, error) {
	err := dto.Attributes.AssignTo(&dto.MagicAuthTokenWithUser.Attributes)
	if err != nil {
		return nil, err
	}
	err = dto.Fields.AssignTo(&dto.MagicAuthToken.Fields)
	if err != nil {
		return nil, err
	}

	dto.MagicAuthTokenWithUser.Secret, err = c.decrypt(dto.MagicAuthTokenWithUser.Secret, dto.MagicAuthTokenWithUser.SecretEncryptionKeyID)
	if err != nil {
		return nil, err
	}

	return dto.MagicAuthTokenWithUser, nil
}

type reportTokenWithSecretDTO struct {
	*database.ReportTokenWithSecret
	SecretEncryptionKeyID string `db:"secret_encryption_key_id"`
}

func (c *connection) reportTokenWithSecretFromDTO(dto *reportTokenWithSecretDTO) (*database.ReportTokenWithSecret, error) {
	if dto.SecretEncryptionKeyID == "" {
		return dto.ReportTokenWithSecret, nil
	}
	decrypted, err := c.decrypt(dto.ReportTokenWithSecret.MagicAuthTokenSecret, dto.SecretEncryptionKeyID)
	if err != nil {
		return nil, err
	}
	dto.ReportTokenWithSecret.MagicAuthTokenSecret = decrypted
	return dto.ReportTokenWithSecret, nil
}

type organizationInviteDTO struct {
	*database.OrganizationInvite
	UsergroupIDs pgtype.TextArray `db:"usergroup_ids"`
}

func (o *organizationInviteDTO) AsModel() (*database.OrganizationInvite, error) {
	err := o.UsergroupIDs.AssignTo(&o.OrganizationInvite.UsergroupIDs)
	if err != nil {
		return nil, err
	}

	return o.OrganizationInvite, nil
}

type billingIssueDTO struct {
	ID               string                     `db:"id"`
	OrgID            string                     `db:"org_id"`
	Type             database.BillingIssueType  `db:"type"`
	Level            database.BillingIssueLevel `db:"level"`
	Metadata         json.RawMessage            `db:"metadata"`
	OverdueProcessed bool                       `db:"overdue_processed"`
	EventTime        time.Time                  `db:"event_time"`
	CreatedOn        time.Time                  `db:"created_on"`
}

func (b *billingIssueDTO) AsModel() *database.BillingIssue {
	var metadata database.BillingIssueMetadata
	switch b.Type {
	case database.BillingIssueTypeOnTrial:
		metadata = &database.BillingIssueMetadataOnTrial{}
	case database.BillingIssueTypeTrialEnded:
		metadata = &database.BillingIssueMetadataTrialEnded{}
	case database.BillingIssueTypeNoPaymentMethod:
		metadata = &database.BillingIssueMetadataNoPaymentMethod{}
	case database.BillingIssueTypeNoBillableAddress:
		metadata = &database.BillingIssueMetadataNoBillableAddress{}
	case database.BillingIssueTypePaymentFailed:
		metadata = &database.BillingIssueMetadataPaymentFailed{}
	case database.BillingIssueTypeSubscriptionCancelled:
		metadata = &database.BillingIssueMetadataSubscriptionCancelled{}
	case database.BillingIssueTypeNeverSubscribed:
		metadata = &database.BillingIssueMetadataNeverSubscribed{}
	default:
	}
	if err := json.Unmarshal(b.Metadata, &metadata); err != nil {
		return nil
	}
	return &database.BillingIssue{
		ID:        b.ID,
		OrgID:     b.OrgID,
		Type:      b.Type,
		Level:     b.Level,
		Metadata:  metadata,
		EventTime: b.EventTime,
		CreatedOn: b.CreatedOn,
	}
}

func (b *billingIssueDTO) getBillingIssueLevel() database.BillingIssueLevel {
	if b.Type == database.BillingIssueTypeUnspecified {
		return database.BillingIssueLevelUnspecified
	}
	if b.Type == database.BillingIssueTypeOnTrial {
		return database.BillingIssueLevelWarning
	}
	return database.BillingIssueLevelError
}

func (c *connection) decryptProjectVariables(res []*database.ProjectVariable) error {
	for _, v := range res {
		if v.ValueEncryptionKeyID == "" {
			continue
		}
		dec, err := base64.StdEncoding.DecodeString(v.Value)
		if err != nil {
			return err
		}

		decryptedValue, err := c.decrypt(dec, v.ValueEncryptionKeyID)
		if err != nil {
			return err
		}

		v.Value = string(decryptedValue)
	}
	return nil
}

// returns the encrypted text and the encryption key id used. The first key in the keyring is used for encryption. If the keyring is empty, the text is returned as is along with an empty key id.
func (c *connection) encrypt(text []byte) ([]byte, string, error) {
	if len(c.encKeyring) == 0 {
		return text, "", nil
	}
	// use the first key in the keyring for encryption
	encrypted, err := encrypt(text, c.encKeyring[0].Secret)
	if err != nil {
		return nil, "", err
	}
	return encrypted, c.encKeyring[0].ID, nil
}

// returns the decrypted text, using the encryption key id provided. If the key id is empty, the text is returned as is.
func (c *connection) decrypt(text []byte, encKeyID string) ([]byte, error) {
	if encKeyID == "" {
		return text, nil
	}
	var encKey *database.EncryptionKey
	for _, key := range c.encKeyring {
		if key.ID == encKeyID {
			encKey = key
			break
		}
	}
	if encKey == nil {
		return nil, fmt.Errorf("encryption key id %s not found in keyring", encKeyID)
	}
	return decrypt(text, encKey.Secret)
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
		return database.NewNotFoundError(fmt.Sprintf("%s not found", target))
	}
	var pgerr *pgconn.PgError
	if !errors.As(err, &pgerr) {
		return err
	}
	if pgerr.Code == "23505" { // unique_violation
		switch pgerr.ConstraintName {
		case "orgs_name_idx":
			return database.NewNotUniqueError("an org with that name already exists")
		case "projects_name_idx":
			return database.NewNotUniqueError("a project with that name already exists in the org")
		case "users_email_idx":
			return database.NewNotUniqueError("a user with that email already exists")
		case "usergroups_name_idx":
			return database.NewNotUniqueError("a usergroup with that name already exists in the org")
		case "usergroups_users_pkey":
			return database.NewNotUniqueError("user is already a member of the usergroup")
		case "users_orgs_roles_pkey":
			return database.NewNotUniqueError("user is already a member of the org")
		case "users_projects_roles_pkey":
			return database.NewNotUniqueError("user is already a member of the project")
		case "usergroups_orgs_roles_pkey":
			return database.NewNotUniqueError("usergroup is already a member of the org")
		case "usergroups_projects_roles_pkey":
			return database.NewNotUniqueError("usergroup is already a member of the project")
		case "org_invites_email_org_idx":
			return database.NewNotUniqueError("email has already been invited to the org")
		case "project_invites_email_project_idx":
			return database.NewNotUniqueError("email has already been invited to the project")
		case "orgs_autoinvite_domains_org_id_domain_idx":
			return database.NewNotUniqueError("domain has already been added for the org")
		case "service_name_idx":
			return database.NewNotUniqueError("a service with that name already exists in the org")
		case "virtual_files_pkey":
			return database.NewNotUniqueError("a virtual file already exists at that path")
		default:
			if target == "" {
				return database.ErrNotUnique
			}
			return database.NewNotUniqueError(fmt.Sprintf("%s already exists", target))
		}
	}
	return err
}

// encrypts plaintext using AES-GCM with the given key and returns the base64 encoded ciphertext
func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// nonce is prepended to the ciphertext, so it can be used for decryption
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypts the ciphertext using AES-GCM with the given key and returns the plaintext
func decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	d, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return d, nil
}
